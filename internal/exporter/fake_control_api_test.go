package exporter

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
)

// fakeControlAPI is a read-only, in-process stand-in for the Control API, serving
// a fixed account so the exporter can be tested with no credentials and no
// network. The exporter only reads, so a fixture is enough; the provider's own
// fake is the stateful one used for CRUD.
type fakeControlAPI struct {
	server *httptest.Server

	// mu guards requests, which is appended to on the server's goroutines and
	// read from the test's.
	mu sync.Mutex
	// requests records the paths that were requested, so tests can assert the
	// exporter is not making calls it should not.
	requests []string
}

const fakeAccountID = "acc1"

// newFakeControlAPI serves one account with one app holding a key, a namespace, a
// queue and five rules. One rule has an unsupported ruleType; one is a Kafka rule
// missing the SASL credentials the API never returns.
func newFakeControlAPI(t *testing.T) *fakeControlAPI {
	t.Helper()
	return newFakeControlAPIWith(t, nil)
}

// newFakeControlAPIWith lets a test alter the rules first, for cases needing the
// API to return something the provider won't accept.
func newFakeControlAPIWith(t *testing.T, alter func(rules map[string]map[string]any)) *fakeControlAPI {
	t.Helper()

	fake := &fakeControlAPI{}

	rules := map[string]map[string]any{
		"rule1": {
			"id":          "rule1",
			"appId":       "app1",
			"status":      "enabled",
			"ruleType":    "http",
			"requestMode": "single",
			"source": map[string]any{
				"channelFilter": "^chat:",
				"type":          "channel.message",
			},
			"target": map[string]any{
				"url":          "https://example.com/hook?a=1",
				"headers":      []any{map[string]any{"name": "X-Ably", "value": "yes"}},
				"signingKeyId": "key1",
				"format":       "json",
				"enveloped":    true,
			},
		},
		"rule2": {
			"id":          "rule2",
			"appId":       "app1",
			"status":      "enabled",
			"ruleType":    "amqp",
			"requestMode": "single",
			"source": map[string]any{
				"channelFilter": "",
				"type":          "channel.message",
			},
			"target": map[string]any{
				"queueId":   "queue1",
				"enveloped": false,
				"format":    "json",
			},
		},
		"rule3": {
			"id":             "rule3",
			"appId":          "app1",
			"status":         "enabled",
			"ruleType":       "bodyguard/text-moderation",
			"invocationMode": "BEFORE_PUBLISH",
			"chatRoomFilter": "/room:.*/",
			"beforePublishConfig": map[string]any{
				"retryTimeout":          3000,
				"maxRetries":            2,
				"failedAction":          "PUBLISH",
				"tooManyRequestsAction": "RETRY",
			},
			"target": map[string]any{
				"apiKey":          "bodyguard-secret",
				"channelId":       "chan-1",
				"apiUrl":          "https://api.bodyguard.ai",
				"defaultLanguage": "en",
			},
		},
		"rule5": {
			"id":          "rule5",
			"appId":       "app1",
			"status":      "enabled",
			"ruleType":    "kafka",
			"requestMode": "single",
			"source": map[string]any{
				"channelFilter": "",
				"type":          "channel.message",
			},
			// The API returns the SASL block without the credentials it was given.
			"target": map[string]any{
				"routingKey": "key",
				"brokers":    []any{"broker.example:9092"},
				"auth":       map[string]any{"sasl": map[string]any{"mechanism": "scram-sha-256"}},
				"enveloped":  false,
				"format":     "json",
			},
		},
		// A rule type the provider has no resource for. The exporter should skip
		// it with a warning rather than failing the whole export.
		"rule4": {
			"id":       "rule4",
			"appId":    "app1",
			"status":   "enabled",
			"ruleType": "hive/text-moderation",
			"target":   map[string]any{"apiKey": "hive-secret"},
		},
	}

	if alter != nil {
		alter(rules)
	}

	responses := map[string]any{
		"/me": map[string]any{
			"account": map[string]any{"id": fakeAccountID, "name": "Test Account"},
		},
		"/accounts/" + fakeAccountID + "/apps": []any{
			map[string]any{
				"id":        "app1",
				"accountId": fakeAccountID,
				"name":      "Chat Service",
				"status":    "enabled",
				"tlsOnly":   true,
				"created":   1600000000000,
				"modified":  1600000000000,
			},
		},
		"/apps/app1/keys": []any{
			map[string]any{
				"id":              "key1",
				"appId":           "app1",
				"name":            "root key",
				"key":             "app1.key1:secret",
				"status":          0,
				"revocableTokens": false,
				"capability":      map[string]any{"chat:*": []any{"publish", "subscribe"}},
				"created":         1600000000000,
				"modified":        1600000000000,
			},
			// Revoked keys are not exportable: the provider reads them as gone.
			map[string]any{
				"id":     "key2",
				"appId":  "app1",
				"name":   "revoked key",
				"status": 1,
			},
		},
		"/apps/app1/namespaces": []any{
			map[string]any{
				"id":          "chat",
				"appId":       "app1",
				"persisted":   true,
				"persistLast": false,
				"pushEnabled": false,
				"tlsOnly":     false,
			},
		},
		"/apps/app1/queues": []any{
			map[string]any{
				"id":         "queue1",
				"appId":      "app1",
				"name":       "orders",
				"region":     "eu-west-1-a",
				"ttl":        60,
				"maxLength":  10000,
				"state":      "active",
				"deadletter": false,
				"amqp":       map[string]any{"uri": "amqps://example", "queueName": "orders"},
				"stomp":      map[string]any{"uri": "stomp://example", "host": "shared", "destination": "/orders"},
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fake.mu.Lock()
		fake.requests = append(fake.requests, r.Method+" "+r.URL.Path)
		fake.mu.Unlock()

		if r.Method != http.MethodGet {
			http.Error(w, `{"message":"the exporter must only read"}`, http.StatusMethodNotAllowed)
			return
		}

		if body, ok := responses[r.URL.Path]; ok {
			writeJSON(t, w, body)
			return
		}

		switch {
		case r.URL.Path == "/apps/app1/rules":
			list := make([]any, 0, len(rules))
			for _, id := range slices.Sorted(maps.Keys(rules)) {
				list = append(list, rules[id])
			}
			writeJSON(t, w, list)
		case strings.HasPrefix(r.URL.Path, "/apps/app1/rules/"):
			id := strings.TrimPrefix(r.URL.Path, "/apps/app1/rules/")
			rule, ok := rules[id]
			if !ok {
				http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
				return
			}
			writeJSON(t, w, rule)
		default:
			http.Error(w, fmt.Sprintf(`{"message":"unexpected path %s"}`, r.URL.Path), http.StatusNotFound)
		}
	})

	fake.server = httptest.NewServer(mux)
	t.Cleanup(fake.server.Close)
	return fake
}

// url is the base URL to give the exporter.
func (f *fakeControlAPI) url() string {
	return f.server.URL
}

// recorded returns a copy of the requests the fake has served.
func (f *fakeControlAPI) recorded() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.requests)
}

// writeJSON encodes a fixture. Errorf rather than Fatalf: FailNow from a handler
// goroutine kills the response mid-flight and surfaces as a transport error.
func writeJSON(t *testing.T, w http.ResponseWriter, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Errorf("encoding fake response: %s", err)
	}
}
