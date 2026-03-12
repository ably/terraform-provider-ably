package ably

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Ingress Postgres Outbox: CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_IngressPostgresOutbox_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	sslRootCert := "-----BEGIN CERTIFICATE-----\nMIIBxTCCAW..."
	body := IngressPostgresOutboxRulePost{
		RuleType: "ingress-postgres-outbox",
		Target: IngressPostgresOutboxTarget{
			URL:               "postgres://user:pass@host:5432/db",
			OutboxTableSchema: "public",
			OutboxTableName:   "outbox",
			NodesTableSchema:  "public",
			NodesTableName:    "nodes",
			SSLMode:           "prefer",
			SSLRootCert:       &sslRootCert,
			PrimarySite:       "us-east-1-A",
		},
	}
	wantResp := RuleResponse{
		ID: "rule010", AppID: "app123", Status: "enabled", RuleType: "ingress-postgres-outbox",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		// Verify no source/requestMode in the wire format.
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.NotContains(t, raw, "source")
		assert.NotContains(t, raw, "requestMode")
		assert.Equal(t, "ingress-postgres-outbox", raw["ruleType"])

		target := raw["target"].(map[string]interface{})
		assert.Equal(t, "postgres://user:pass@host:5432/db", target["url"])
		assert.Equal(t, "public", target["outboxTableSchema"])
		assert.Equal(t, "prefer", target["sslMode"])
		assert.Contains(t, target, "sslRootCert")
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule010", got.ID)
	assert.Equal(t, "ingress-postgres-outbox", got.RuleType)
}

func TestCreateRule_IngressPostgresOutbox_NilSSLRootCert(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := IngressPostgresOutboxRulePost{
		RuleType: "ingress-postgres-outbox",
		Target: IngressPostgresOutboxTarget{
			URL:               "postgres://host/db",
			OutboxTableSchema: "public",
			OutboxTableName:   "outbox",
			NodesTableSchema:  "public",
			NodesTableName:    "nodes",
			SSLMode:           "prefer",
			PrimarySite:       "us-east-1-A",
		},
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		target := raw["target"].(map[string]interface{})
		assert.NotContains(t, target, "sslRootCert")
		writeJSON(w, http.StatusCreated, RuleResponse{ID: "rule010", RuleType: "ingress-postgres-outbox"})
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule010", got.ID)
}

// ---------------------------------------------------------------------------
// Ingress Postgres Outbox: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_IngressPostgresOutbox_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := IngressPostgresOutboxRulePatch{
		RuleType: "ingress-postgres-outbox",
		Target: &IngressPostgresOutboxTarget{
			URL:               "postgres://user:pass@host:5432/db",
			OutboxTableSchema: "public",
			OutboxTableName:   "outbox",
			NodesTableSchema:  "public",
			NodesTableName:    "nodes",
			SSLMode:           "verify-full",
			PrimarySite:       "eu-west-1-A",
		},
	}
	wantResp := RuleResponse{
		ID: "rule010", AppID: "app123", Status: "enabled", RuleType: "ingress-postgres-outbox",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule010", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.NotContains(t, raw, "source")
		assert.NotContains(t, raw, "requestMode")
		assert.Equal(t, "ingress-postgres-outbox", raw["ruleType"])

		target := raw["target"].(map[string]interface{})
		assert.Equal(t, "verify-full", target["sslMode"])
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule010", body)
	require.NoError(t, err)
	assert.Equal(t, "rule010", got.ID)
}

func TestUpdateRule_IngressPostgresOutboxPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := IngressPostgresOutboxRulePatch{RuleType: "ingress-postgres-outbox"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule010", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Contains(t, raw, "ruleType")
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "target")
		assert.NotContains(t, raw, "source")
		assert.NotContains(t, raw, "requestMode")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule010", RuleType: "ingress-postgres-outbox"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule010", body)
	require.NoError(t, err)
	assert.Equal(t, "rule010", got.ID)
}

// ---------------------------------------------------------------------------
// Ingress MongoDB: CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_IngressMongoDB_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := IngressMongoDBRulePost{
		RuleType: "ingress/mongodb",
		Target: IngressMongoDBTarget{
			URL:                      "mongodb+srv://user:pass@cluster.mongodb.net",
			Database:                 "mydb",
			Collection:               "events",
			Pipeline:                 `[{"$match": {"operationType": "insert"}}]`,
			FullDocument:             "updateLookup",
			FullDocumentBeforeChange: "whenAvailable",
			PrimarySite:              "us-east-1-A",
		},
	}
	wantResp := RuleResponse{
		ID: "rule011", AppID: "app123", Status: "enabled", RuleType: "ingress/mongodb",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.NotContains(t, raw, "source")
		assert.NotContains(t, raw, "requestMode")
		assert.Equal(t, "ingress/mongodb", raw["ruleType"])

		target := raw["target"].(map[string]interface{})
		assert.Equal(t, "mongodb+srv://user:pass@cluster.mongodb.net", target["url"])
		assert.Equal(t, "mydb", target["database"])
		assert.Equal(t, "events", target["collection"])
		assert.Equal(t, "updateLookup", target["fullDocument"])
		assert.Equal(t, "whenAvailable", target["fullDocumentBeforeChange"])
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule011", got.ID)
	assert.Equal(t, "ingress/mongodb", got.RuleType)
}

func TestCreateRule_IngressMongoDB_OmitsOptionalFields(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := IngressMongoDBRulePost{
		RuleType: "ingress/mongodb",
		Target: IngressMongoDBTarget{
			URL:         "mongodb+srv://host",
			Database:    "db",
			Collection:  "col",
			Pipeline:    "[]",
			PrimarySite: "us-east-1-A",
		},
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		target := raw["target"].(map[string]interface{})
		assert.NotContains(t, target, "fullDocument")
		assert.NotContains(t, target, "fullDocumentBeforeChange")
		writeJSON(w, http.StatusCreated, RuleResponse{ID: "rule011", RuleType: "ingress/mongodb"})
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule011", got.ID)
}

// ---------------------------------------------------------------------------
// Ingress MongoDB: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_IngressMongoDB_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := IngressMongoDBRulePatch{
		RuleType: "ingress/mongodb",
		Target: &IngressMongoDBTarget{
			URL:         "mongodb+srv://user:pass@cluster.mongodb.net",
			Database:    "mydb",
			Collection:  "events",
			Pipeline:    "[]",
			PrimarySite: "eu-west-1-A",
		},
	}
	wantResp := RuleResponse{
		ID: "rule011", AppID: "app123", Status: "enabled", RuleType: "ingress/mongodb",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule011", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.NotContains(t, raw, "source")
		assert.NotContains(t, raw, "requestMode")

		target := raw["target"].(map[string]interface{})
		assert.Equal(t, "eu-west-1-A", target["primarySite"])
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule011", body)
	require.NoError(t, err)
	assert.Equal(t, "rule011", got.ID)
}

func TestUpdateRule_IngressMongoDBPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := IngressMongoDBRulePatch{RuleType: "ingress/mongodb"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule011", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Contains(t, raw, "ruleType")
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "target")
		assert.NotContains(t, raw, "source")
		assert.NotContains(t, raw, "requestMode")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule011", RuleType: "ingress/mongodb"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule011", body)
	require.NoError(t, err)
	assert.Equal(t, "rule011", got.ID)
}
