// Package provider implements the Ably provider for Terraform.
//
// This file provides an in-process, stateful stand-in for the Ably Control
// API so the provider's acceptance tests can run with NO credentials and NO
// network access. It is the "Tier 1" hermetic loop described in
// CODEGEN_STRATEGY.md: the loop an AI agent (or CI on a fork) can run on every
// change to prove the provider's CRUD/import/diff logic is internally
// consistent.
//
// How it wires in: TestMain below starts the fake by default and stands aside
// only when TF_ACC is already set (as `make testacc` and CI do, pointing at a
// real Control API). In fake mode it points the provider at the fake via
// ABLY_URL, sets a dummy ABLY_ACCOUNT_TOKEN, and sets TF_ACC=1 so the existing
// acceptance tests run unchanged. No provider source changes are needed: the
// provider already honours ABLY_URL and the client does a plain BaseURL+path
// with no host allow-listing.
//
// What it deliberately does NOT do: validate rule-type-specific fields or
// reproduce real API business rules. It stores whatever JSON body it is sent,
// stamps the server-assigned fields (id, appId, created, modified), and echoes
// the record back on read. That is enough to exercise schema validation, plan
// stability, full CRUD wiring, import, and attribute mapping. It proves the
// provider is internally consistent, not that it matches production; the
// staging-backed acceptance suite (Tier 2) is what keeps it honest.
package provider

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeAccountID is the account ID the fake reports from GET /me. The provider
// discovers the account ID at runtime via /me (see provider.go Configure), so
// tests never hardcode it.
const fakeAccountID = "fakeAccId"

// fakeControlAPI is an in-memory, stateful Control API stand-in.
//
// It is concurrency-safe: every map access is guarded by mu, and resources are
// scoped by their parent app ID, so even if tests are made parallel (each
// creating its own app) they do not interfere with one another.
type fakeControlAPI struct {
	server *httptest.Server

	mu  sync.Mutex
	seq int64

	apps       map[string]record            // appID -> app
	keys       map[string]map[string]record // appID -> keyID -> key
	namespaces map[string]map[string]record // appID -> nsID -> namespace
	queues     map[string]map[string]record // appID -> queueID -> queue
	rules      map[string]map[string]record // appID -> ruleID -> rule
}

// record is a single resource stored as decoded JSON.
type record = map[string]any

func newFakeControlAPI() *fakeControlAPI {
	f := &fakeControlAPI{
		apps:       map[string]record{},
		keys:       map[string]map[string]record{},
		namespaces: map[string]map[string]record{},
		queues:     map[string]map[string]record{},
		rules:      map[string]map[string]record{},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /me", f.handleMe)

	mux.HandleFunc("GET /accounts/{accountID}/apps", f.listApps)
	mux.HandleFunc("POST /accounts/{accountID}/apps", f.createApp)
	mux.HandleFunc("PATCH /apps/{appID}", f.updateApp)
	mux.HandleFunc("DELETE /apps/{appID}", f.deleteApp)

	mux.HandleFunc("GET /apps/{appID}/keys", f.listKeys)
	mux.HandleFunc("POST /apps/{appID}/keys", f.createKey)
	mux.HandleFunc("PATCH /apps/{appID}/keys/{keyID}", f.updateKey)
	mux.HandleFunc("POST /apps/{appID}/keys/{keyID}/revoke", f.revokeKey)

	mux.HandleFunc("GET /apps/{appID}/namespaces", f.listNamespaces)
	mux.HandleFunc("POST /apps/{appID}/namespaces", f.createNamespace)
	mux.HandleFunc("PATCH /apps/{appID}/namespaces/{nsID}", f.updateNamespace)
	mux.HandleFunc("DELETE /apps/{appID}/namespaces/{nsID}", f.deleteNamespace)

	mux.HandleFunc("GET /apps/{appID}/queues", f.listQueues)
	mux.HandleFunc("POST /apps/{appID}/queues", f.createQueue)
	mux.HandleFunc("DELETE /apps/{appID}/queues/{queueID}", f.deleteQueue)

	mux.HandleFunc("GET /apps/{appID}/rules", f.listRules)
	mux.HandleFunc("POST /apps/{appID}/rules", f.createRule)
	mux.HandleFunc("GET /apps/{appID}/rules/{ruleID}", f.getRule)
	mux.HandleFunc("PATCH /apps/{appID}/rules/{ruleID}", f.updateRule)
	mux.HandleFunc("DELETE /apps/{appID}/rules/{ruleID}", f.deleteRule)

	// Read-only/stats endpoints the provider may call. Return empty so they
	// never 404.
	mux.HandleFunc("GET /apps/{appID}/stats", f.emptyArray)
	mux.HandleFunc("GET /accounts/{accountID}/stats", f.emptyArray)

	var h http.Handler = mux
	if os.Getenv("FAKE_DEBUG") != "" {
		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &dbgRecorder{ResponseWriter: w, status: 200}
			mux.ServeHTTP(rec, r)
			fmt.Fprintf(os.Stderr, "[fake] %s %s -> %d %s\n", r.Method, r.URL.Path, rec.status, rec.buf.String())
		})
	}
	f.server = httptest.NewServer(h)
	return f
}

type dbgRecorder struct {
	http.ResponseWriter
	status int
	buf    strings.Builder
}

func (d *dbgRecorder) WriteHeader(s int) { d.status = s; d.ResponseWriter.WriteHeader(s) }
func (d *dbgRecorder) Write(b []byte) (int, error) {
	d.buf.Write(b)
	return d.ResponseWriter.Write(b)
}

// --- helpers ---------------------------------------------------------------

func (f *fakeControlAPI) nextID(prefix string) string {
	n := atomic.AddInt64(&f.seq, 1)
	return fmt.Sprintf("%s%05d", prefix, n)
}

func fakeNow() int64 { return time.Now().UnixMilli() }

func fakeWriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func fakeWriteError(w http.ResponseWriter, status int, message string) {
	fakeWriteJSON(w, status, map[string]any{
		"message":    message,
		"statusCode": status,
		"code":       status * 100,
	})
}

func fakeDecodeBody(r *http.Request) record {
	m := record{}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&m)
	}
	return m
}

// values returns the records in a sub-store as a slice (for List endpoints).
func values(m map[string]record) []record {
	out := make([]record, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

// --- /me -------------------------------------------------------------------

func (f *fakeControlAPI) handleMe(w http.ResponseWriter, _ *http.Request) {
	fakeWriteJSON(w, http.StatusOK, map[string]any{
		"account": map[string]any{"id": fakeAccountID, "name": "Fake Account"},
		"user":    map[string]any{"id": 1, "email": "fake@ably.invalid"},
		"token":   map[string]any{"id": "fake", "name": "fake", "capabilities": []string{}},
	})
}

// --- apps ------------------------------------------------------------------

func (f *fakeControlAPI) createApp(w http.ResponseWriter, r *http.Request) {
	body := fakeDecodeBody(r)
	f.mu.Lock()
	defer f.mu.Unlock()

	id := f.nextID("app")
	ts := fakeNow()
	body["id"] = id
	body["accountId"] = fakeAccountID
	body["created"] = ts
	body["modified"] = ts
	f.apps[id] = body
	fakeWriteJSON(w, http.StatusCreated, body)
}

func (f *fakeControlAPI) listApps(w http.ResponseWriter, _ *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fakeWriteJSON(w, http.StatusOK, values(f.apps))
}

func (f *fakeControlAPI) updateApp(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	body := fakeDecodeBody(r)
	f.mu.Lock()
	defer f.mu.Unlock()

	rec, ok := f.apps[appID]
	if !ok {
		fakeWriteError(w, http.StatusNotFound, "App not found")
		return
	}
	maps.Copy(rec, body)
	rec["modified"] = fakeNow()
	fakeWriteJSON(w, http.StatusOK, rec)
}

func (f *fakeControlAPI) deleteApp(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.apps[appID]; !ok {
		fakeWriteError(w, http.StatusNotFound, "App not found")
		return
	}
	delete(f.apps, appID)
	delete(f.keys, appID)
	delete(f.namespaces, appID)
	delete(f.queues, appID)
	delete(f.rules, appID)
	w.WriteHeader(http.StatusNoContent)
}

// --- keys ------------------------------------------------------------------

func (f *fakeControlAPI) createKey(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	body := fakeDecodeBody(r)
	f.mu.Lock()
	defer f.mu.Unlock()

	id := f.nextID("key")
	ts := fakeNow()
	body["id"] = id
	body["appId"] = appID
	body["status"] = 0
	// The full key string is only ever returned by the create endpoint in the
	// real API; the resource preserves it from state thereafter.
	body["key"] = fmt.Sprintf("%s.%s:%s", appID, id, f.nextID("secret"))
	body["created"] = ts
	body["modified"] = ts
	if f.keys[appID] == nil {
		f.keys[appID] = map[string]record{}
	}
	f.keys[appID][id] = body
	fakeWriteJSON(w, http.StatusCreated, body)
}

func (f *fakeControlAPI) listKeys(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	f.mu.Lock()
	defer f.mu.Unlock()
	fakeWriteJSON(w, http.StatusOK, values(f.keys[appID]))
}

func (f *fakeControlAPI) updateKey(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	keyID := r.PathValue("keyID")
	body := fakeDecodeBody(r)
	f.mu.Lock()
	defer f.mu.Unlock()

	rec, ok := f.keys[appID][keyID]
	if !ok {
		fakeWriteError(w, http.StatusNotFound, "Key not found")
		return
	}
	maps.Copy(rec, body)
	rec["modified"] = fakeNow()
	fakeWriteJSON(w, http.StatusOK, rec)
}

func (f *fakeControlAPI) revokeKey(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	keyID := r.PathValue("keyID")
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.keys[appID][keyID]; !ok {
		fakeWriteError(w, http.StatusNotFound, "Key not found")
		return
	}
	// The provider's key delete path calls revoke; drop it so subsequent reads
	// and destroy checks see it gone.
	delete(f.keys[appID], keyID)
	w.WriteHeader(http.StatusOK)
}

// --- namespaces ------------------------------------------------------------

func (f *fakeControlAPI) createNamespace(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	body := fakeDecodeBody(r)
	f.mu.Lock()
	defer f.mu.Unlock()

	// Namespace ID is client-supplied (the channel namespace prefix).
	id, _ := body["id"].(string)
	ts := fakeNow()
	body["appId"] = appID
	body["created"] = ts
	body["modified"] = ts
	// The real API always returns these flags (defaulting to false). Omitting
	// them makes the provider record null, which then drifts to false on the
	// next plan.
	if _, ok := body["batchingEnabled"]; !ok {
		body["batchingEnabled"] = false
	}
	if _, ok := body["conflationEnabled"]; !ok {
		body["conflationEnabled"] = false
	}
	if f.namespaces[appID] == nil {
		f.namespaces[appID] = map[string]record{}
	}
	f.namespaces[appID][id] = body
	fakeWriteJSON(w, http.StatusCreated, body)
}

func (f *fakeControlAPI) listNamespaces(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	f.mu.Lock()
	defer f.mu.Unlock()
	fakeWriteJSON(w, http.StatusOK, values(f.namespaces[appID]))
}

func (f *fakeControlAPI) updateNamespace(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	nsID := r.PathValue("nsID")
	body := fakeDecodeBody(r)
	f.mu.Lock()
	defer f.mu.Unlock()

	rec, ok := f.namespaces[appID][nsID]
	if !ok {
		fakeWriteError(w, http.StatusNotFound, "Namespace not found")
		return
	}
	maps.Copy(rec, body)
	rec["modified"] = fakeNow()
	fakeWriteJSON(w, http.StatusOK, rec)
}

func (f *fakeControlAPI) deleteNamespace(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	nsID := r.PathValue("nsID")
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.namespaces[appID][nsID]; !ok {
		fakeWriteError(w, http.StatusNotFound, "Namespace not found")
		return
	}
	delete(f.namespaces[appID], nsID)
	w.WriteHeader(http.StatusNoContent)
}

// --- queues ----------------------------------------------------------------

func (f *fakeControlAPI) createQueue(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	body := fakeDecodeBody(r)
	f.mu.Lock()
	defer f.mu.Unlock()

	id := f.nextID("queue")
	name, _ := body["name"].(string)
	body["id"] = id
	body["appId"] = appID
	body["state"] = "running"
	body["deadletter"] = false
	body["amqp"] = map[string]any{
		"uri":       "amqps://fake.ably.invalid",
		"queueName": fmt.Sprintf("%s:%s", appID, name),
	}
	body["stomp"] = map[string]any{
		"uri":         "stomp://fake.ably.invalid",
		"host":        "shared.ably.invalid",
		"destination": fmt.Sprintf("/amq/queue/%s:%s", appID, name),
	}
	body["messages"] = map[string]any{}
	body["stats"] = map[string]any{}
	if f.queues[appID] == nil {
		f.queues[appID] = map[string]record{}
	}
	f.queues[appID][id] = body
	fakeWriteJSON(w, http.StatusCreated, body)
}

func (f *fakeControlAPI) listQueues(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	f.mu.Lock()
	defer f.mu.Unlock()
	fakeWriteJSON(w, http.StatusOK, values(f.queues[appID]))
}

func (f *fakeControlAPI) deleteQueue(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	queueID := r.PathValue("queueID")
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.queues[appID][queueID]; !ok {
		fakeWriteError(w, http.StatusNotFound, "Queue not found")
		return
	}
	delete(f.queues[appID], queueID)
	w.WriteHeader(http.StatusNoContent)
}

// --- rules -----------------------------------------------------------------
//
// All rule variants (http, aws/*, moderation, ingress, ...) share these
// handlers. The fake stores the posted body verbatim and echoes it, so the
// polymorphic target round-trips without the fake needing to know the rule
// type.

func (f *fakeControlAPI) createRule(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	body := fakeDecodeBody(r)
	f.mu.Lock()
	defer f.mu.Unlock()

	id := f.nextID("rule")
	ts := fakeNow()
	body["id"] = id
	body["appId"] = appID
	body["version"] = "1"
	body["created"] = ts
	body["modified"] = ts
	if _, ok := body["status"]; !ok {
		body["status"] = "enabled"
	}
	// HTTP-family rule targets default their message format to "json" on the
	// real API when not specified. Mirror that so computed `target.format`
	// attributes are populated.
	if t, ok := body["target"].(map[string]any); ok {
		if _, has := t["format"]; !has {
			t["format"] = "json"
		}
	}
	if f.rules[appID] == nil {
		f.rules[appID] = map[string]record{}
	}
	f.rules[appID][id] = body
	fakeWriteJSON(w, http.StatusCreated, body)
}

func (f *fakeControlAPI) getRule(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	ruleID := r.PathValue("ruleID")
	f.mu.Lock()
	defer f.mu.Unlock()

	rec, ok := f.rules[appID][ruleID]
	if !ok {
		fakeWriteError(w, http.StatusNotFound, "Rule not found")
		return
	}
	fakeWriteJSON(w, http.StatusOK, rec)
}

func (f *fakeControlAPI) listRules(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	f.mu.Lock()
	defer f.mu.Unlock()
	fakeWriteJSON(w, http.StatusOK, values(f.rules[appID]))
}

func (f *fakeControlAPI) updateRule(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	ruleID := r.PathValue("ruleID")
	body := fakeDecodeBody(r)
	f.mu.Lock()
	defer f.mu.Unlock()

	rec, ok := f.rules[appID][ruleID]
	if !ok {
		fakeWriteError(w, http.StatusNotFound, "Rule not found")
		return
	}
	maps.Copy(rec, body)
	if t, ok := rec["target"].(map[string]any); ok {
		if _, has := t["format"]; !has {
			t["format"] = "json"
		}
	}
	rec["modified"] = fakeNow()
	fakeWriteJSON(w, http.StatusOK, rec)
}

func (f *fakeControlAPI) deleteRule(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("appID")
	ruleID := r.PathValue("ruleID")
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.rules[appID][ruleID]; !ok {
		fakeWriteError(w, http.StatusNotFound, "Rule not found")
		return
	}
	delete(f.rules[appID], ruleID)
	w.WriteHeader(http.StatusNoContent)
}

// --- misc ------------------------------------------------------------------

func (f *fakeControlAPI) emptyArray(w http.ResponseWriter, _ *http.Request) {
	fakeWriteJSON(w, http.StatusOK, []any{})
}

// TestMain runs the acceptance suite against the hermetic fake by default.
// Unless an explicit real run is requested via TF_ACC (as `make testacc` and CI
// do, pointing at a real Control API), it stands up the in-process fake, builds
// the current provider, and points Terraform at it via a clean dev_overrides
// config. With TF_ACC already set it stands aside entirely.
func TestMain(m *testing.M) {
	var fake *fakeControlAPI
	var hermeticDir string
	if os.Getenv("TF_ACC") == "" {
		fake = newFakeControlAPI()
		_ = os.Setenv("ABLY_URL", fake.server.URL)
		_ = os.Setenv("ABLY_ACCOUNT_TOKEN", "fake-token")
		_ = os.Setenv("TF_ACC", "1")
		dir, err := setupHermeticProvider()
		if err != nil {
			fmt.Fprintln(os.Stderr, "hermetic setup failed:", err)
			os.Exit(1)
		}
		hermeticDir = dir
	}
	code := m.Run()
	// os.Exit skips deferred cleanup, so tear down explicitly here.
	if fake != nil {
		fake.server.Close()
	}
	if hermeticDir != "" {
		_ = os.RemoveAll(hermeticDir)
	}
	os.Exit(code)
}

// setupHermeticProvider builds the provider from the current source into a temp
// directory and writes a Terraform CLI config whose dev_overrides points at it.
//
// This is deliberately a fresh build, not a reattach: the provider source is
// pinned to the "ably/ably" namespace, which the in-process reattach factory
// (keyed by the bare type "ably") cannot satisfy, and any ambient
// ~/.terraformrc dev_overrides would otherwise run a stale installed binary
// (the trap that the Phase 0c spike and CODEGEN_STRATEGY.md call out). Building
// here guarantees the tests exercise THIS code with no network install.
//
// It returns the temp directory it created so the caller can remove it.
func setupHermeticProvider() (string, error) {
	dir, err := os.MkdirTemp("", "ably-hermetic")
	if err != nil {
		return "", err
	}
	bin := filepath.Join(dir, "terraform-provider-ably")
	build := exec.Command("go", "build", "-o", bin, "github.com/ably/terraform-provider-ably")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return dir, fmt.Errorf("build provider: %w", err)
	}
	rc := filepath.Join(dir, "dev.tfrc")
	cfg := fmt.Sprintf("disable_checkpoint = true\n"+
		"provider_installation {\n"+
		"  dev_overrides {\n"+
		"    \"ably/ably\" = %q\n"+
		"  }\n"+
		"  direct {}\n"+
		"}\n", dir)
	if err := os.WriteFile(rc, []byte(cfg), 0o644); err != nil {
		return dir, err
	}
	// Override any ambient ~/.terraformrc and stale plugin cache so only our
	// freshly built provider is used.
	_ = os.Setenv("TF_CLI_CONFIG_FILE", rc)
	_ = os.Unsetenv("TF_PLUGIN_CACHE_DIR")
	return dir, nil
}
