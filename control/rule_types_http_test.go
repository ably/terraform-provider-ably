package control

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// RuleSourcePatch
// ---------------------------------------------------------------------------

func TestRuleSourcePatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	patch := RuleSourcePatch{}

	data, err := json.Marshal(patch)
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(data))
}

// ---------------------------------------------------------------------------
// IFTTT: CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_IFTTT_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := IFTTTRulePost{
		RuleType: "http/ifttt", RequestMode: "single",
		Source: RuleSource{ChannelFilter: "^test", Type: "channel.message"},
		Target: IFTTTRuleTarget{WebhookKey: "key123", EventName: "event1"},
	}
	wantResp := RuleResponse{
		ID: "rule001", AppID: "app123", Status: "enabled", RuleType: "http/ifttt",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got IFTTTRulePost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "http/ifttt", got.RuleType)
		assert.Equal(t, "single", got.RequestMode)
		assert.Equal(t, "^test", got.Source.ChannelFilter)
		assert.Equal(t, "key123", got.Target.WebhookKey)
		assert.Equal(t, "event1", got.Target.EventName)
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule001", got.ID)
	assert.Equal(t, "http/ifttt", got.RuleType)
}

// ---------------------------------------------------------------------------
// IFTTT: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_IFTTT_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := IFTTTRulePatch{
		RuleType: "http/ifttt",
		Source:   &RuleSourcePatch{ChannelFilter: "^updated", Type: "channel.message"},
		Target:   &IFTTTRuleTargetPatch{WebhookKey: ptr("newkey"), EventName: ptr("newevent")},
	}
	wantResp := RuleResponse{
		ID: "rule001", AppID: "app123", Status: "enabled", RuleType: "http/ifttt",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule001", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got IFTTTRulePatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "http/ifttt", got.RuleType)
		require.NotNil(t, got.Source)
		assert.Equal(t, "^updated", got.Source.ChannelFilter)
		require.NotNil(t, got.Target)
		require.NotNil(t, got.Target.WebhookKey)
		assert.Equal(t, "newkey", *got.Target.WebhookKey)
		require.NotNil(t, got.Target.EventName)
		assert.Equal(t, "newevent", *got.Target.EventName)
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule001", body)
	require.NoError(t, err)
	assert.Equal(t, "rule001", got.ID)
	assert.Equal(t, "http/ifttt", got.RuleType)
}

func TestUpdateRule_IFTTTPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := IFTTTRulePatch{RuleType: "http/ifttt"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule001", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "http/ifttt", raw["ruleType"])
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "requestMode")
		assert.NotContains(t, raw, "source")
		assert.NotContains(t, raw, "target")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule001", RuleType: "http/ifttt"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule001", body)
	require.NoError(t, err)
	assert.Equal(t, "rule001", got.ID)
}

// ---------------------------------------------------------------------------
// Zapier: CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_Zapier_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	signingKey := "key-abc"
	body := ZapierRulePost{
		RuleType: "http/zapier", RequestMode: "single",
		Source: RuleSource{ChannelFilter: "^test", Type: "channel.message"},
		Target: ZapierRuleTarget{
			URL:          "https://hooks.zapier.com/test",
			Headers:      []RuleHeader{{Name: "X-Custom", Value: "val"}},
			SigningKeyID: &signingKey,
		},
	}
	wantResp := RuleResponse{
		ID: "rule002", AppID: "app123", Status: "enabled", RuleType: "http/zapier",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got ZapierRulePost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "http/zapier", got.RuleType)
		assert.Equal(t, "single", got.RequestMode)
		assert.Equal(t, "https://hooks.zapier.com/test", got.Target.URL)
		require.Len(t, got.Target.Headers, 1)
		assert.Equal(t, "X-Custom", got.Target.Headers[0].Name)
		require.NotNil(t, got.Target.SigningKeyID)
		assert.Equal(t, "key-abc", *got.Target.SigningKeyID)
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule002", got.ID)
	assert.Equal(t, "http/zapier", got.RuleType)
}

// ---------------------------------------------------------------------------
// Zapier: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_Zapier_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := ZapierRulePatch{
		RuleType: "http/zapier",
		Target:   &ZapierRuleTargetPatch{URL: ptr("https://hooks.zapier.com/updated")},
	}
	wantResp := RuleResponse{
		ID: "rule002", AppID: "app123", Status: "enabled", RuleType: "http/zapier",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule002", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got ZapierRulePatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "http/zapier", got.RuleType)
		require.NotNil(t, got.Target)
		require.NotNil(t, got.Target.URL)
		assert.Equal(t, "https://hooks.zapier.com/updated", *got.Target.URL)
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule002", body)
	require.NoError(t, err)
	assert.Equal(t, "rule002", got.ID)
}

func TestUpdateRule_ZapierPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := ZapierRulePatch{RuleType: "http/zapier"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule002", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "http/zapier", raw["ruleType"])
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "requestMode")
		assert.NotContains(t, raw, "source")
		assert.NotContains(t, raw, "target")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule002", RuleType: "http/zapier"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule002", body)
	require.NoError(t, err)
	assert.Equal(t, "rule002", got.ID)
}

// ---------------------------------------------------------------------------
// Cloudflare Worker: CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_CloudflareWorker_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	signingKey := "key-cf"
	body := CloudflareWorkerRulePost{
		RuleType: "http/cloudflare-worker", RequestMode: "single",
		Source: RuleSource{ChannelFilter: "^test", Type: "channel.message"},
		Target: CloudflareWorkerRuleTarget{
			URL:          "https://worker.example.com",
			Headers:      []RuleHeader{{Name: "X-CF", Value: "123"}},
			SigningKeyID: &signingKey,
		},
	}
	wantResp := RuleResponse{
		ID: "rule003", AppID: "app123", Status: "enabled", RuleType: "http/cloudflare-worker",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got CloudflareWorkerRulePost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "http/cloudflare-worker", got.RuleType)
		assert.Equal(t, "https://worker.example.com", got.Target.URL)
		require.NotNil(t, got.Target.SigningKeyID)
		assert.Equal(t, "key-cf", *got.Target.SigningKeyID)
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule003", got.ID)
	assert.Equal(t, "http/cloudflare-worker", got.RuleType)
}

// ---------------------------------------------------------------------------
// Cloudflare Worker: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_CloudflareWorker_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := CloudflareWorkerRulePatch{
		RuleType: "http/cloudflare-worker",
		Target:   &CloudflareWorkerRuleTargetPatch{URL: ptr("https://worker.example.com/v2")},
	}
	wantResp := RuleResponse{
		ID: "rule003", AppID: "app123", Status: "enabled", RuleType: "http/cloudflare-worker",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule003", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got CloudflareWorkerRulePatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "http/cloudflare-worker", got.RuleType)
		require.NotNil(t, got.Target)
		require.NotNil(t, got.Target.URL)
		assert.Equal(t, "https://worker.example.com/v2", *got.Target.URL)
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule003", body)
	require.NoError(t, err)
	assert.Equal(t, "rule003", got.ID)
}

func TestUpdateRule_CloudflareWorkerPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := CloudflareWorkerRulePatch{RuleType: "http/cloudflare-worker"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule003", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "http/cloudflare-worker", raw["ruleType"])
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "requestMode")
		assert.NotContains(t, raw, "source")
		assert.NotContains(t, raw, "target")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule003", RuleType: "http/cloudflare-worker"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule003", body)
	require.NoError(t, err)
	assert.Equal(t, "rule003", got.ID)
}

// ---------------------------------------------------------------------------
// Azure Function: CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_AzureFunction_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	signingKey := "key-az"
	enveloped := true
	body := AzureFunctionRulePost{
		RuleType: "http/azure-function", RequestMode: "single",
		Source: RuleSource{ChannelFilter: "^test", Type: "channel.message"},
		Target: AzureFunctionRuleTarget{
			AzureAppID:        "my-app-id",
			AzureFunctionName: "my-function",
			Headers:           []RuleHeader{{Name: "X-Az", Value: "hdr"}},
			SigningKeyID:      &signingKey,
			Enveloped:         &enveloped,
			Format:            "json",
		},
	}
	wantResp := RuleResponse{
		ID: "rule004", AppID: "app123", Status: "enabled", RuleType: "http/azure-function",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got AzureFunctionRulePost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "http/azure-function", got.RuleType)
		assert.Equal(t, "my-app-id", got.Target.AzureAppID)
		assert.Equal(t, "my-function", got.Target.AzureFunctionName)
		require.NotNil(t, got.Target.Enveloped)
		assert.True(t, *got.Target.Enveloped)
		assert.Equal(t, "json", got.Target.Format)
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule004", got.ID)
	assert.Equal(t, "http/azure-function", got.RuleType)
}

// ---------------------------------------------------------------------------
// Azure Function: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_AzureFunction_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	enveloped := false
	body := AzureFunctionRulePatch{
		RuleType: "http/azure-function",
		Target: &AzureFunctionRuleTargetPatch{
			AzureAppID:        ptr("updated-app"),
			AzureFunctionName: ptr("updated-func"),
			Enveloped:         &enveloped,
		},
	}
	wantResp := RuleResponse{
		ID: "rule004", AppID: "app123", Status: "enabled", RuleType: "http/azure-function",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule004", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got AzureFunctionRulePatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "http/azure-function", got.RuleType)
		require.NotNil(t, got.Target)
		require.NotNil(t, got.Target.AzureAppID)
		assert.Equal(t, "updated-app", *got.Target.AzureAppID)
		require.NotNil(t, got.Target.Enveloped)
		assert.False(t, *got.Target.Enveloped)
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule004", body)
	require.NoError(t, err)
	assert.Equal(t, "rule004", got.ID)
}

func TestUpdateRule_AzureFunctionPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := AzureFunctionRulePatch{RuleType: "http/azure-function"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule004", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "http/azure-function", raw["ruleType"])
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "requestMode")
		assert.NotContains(t, raw, "source")
		assert.NotContains(t, raw, "target")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule004", RuleType: "http/azure-function"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule004", body)
	require.NoError(t, err)
	assert.Equal(t, "rule004", got.ID)
}

// ---------------------------------------------------------------------------
// Google Cloud Function: CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_GoogleCloudFunction_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	signingKey := "key-gcf"
	enveloped := true
	body := GoogleCloudFunctionRulePost{
		RuleType: "http/google-cloud-function", RequestMode: "single",
		Source: RuleSource{ChannelFilter: "^test", Type: "channel.message"},
		Target: GoogleCloudFunctionRuleTarget{
			Region:       "us-central1",
			ProjectID:    "my-project",
			FunctionName: "my-function",
			Headers:      []RuleHeader{{Name: "X-GCF", Value: "hdr"}},
			SigningKeyID: &signingKey,
			Enveloped:    &enveloped,
			Format:       "json",
		},
	}
	wantResp := RuleResponse{
		ID: "rule005", AppID: "app123", Status: "enabled", RuleType: "http/google-cloud-function",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got GoogleCloudFunctionRulePost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "http/google-cloud-function", got.RuleType)
		assert.Equal(t, "us-central1", got.Target.Region)
		assert.Equal(t, "my-project", got.Target.ProjectID)
		assert.Equal(t, "my-function", got.Target.FunctionName)
		require.NotNil(t, got.Target.Enveloped)
		assert.True(t, *got.Target.Enveloped)
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule005", got.ID)
	assert.Equal(t, "http/google-cloud-function", got.RuleType)
}

// ---------------------------------------------------------------------------
// Google Cloud Function: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_GoogleCloudFunction_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	enveloped := false
	body := GoogleCloudFunctionRulePatch{
		RuleType: "http/google-cloud-function",
		Source:   &RuleSourcePatch{ChannelFilter: "^updated"},
		Target: &GoogleCloudFunctionRuleTargetPatch{
			Region:       ptr("europe-west1"),
			ProjectID:    ptr("updated-project"),
			FunctionName: ptr("updated-function"),
			Enveloped:    &enveloped,
			Format:       ptr("msgpack"),
		},
	}
	wantResp := RuleResponse{
		ID: "rule005", AppID: "app123", Status: "enabled", RuleType: "http/google-cloud-function",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule005", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got GoogleCloudFunctionRulePatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "http/google-cloud-function", got.RuleType)
		require.NotNil(t, got.Source)
		assert.Equal(t, "^updated", got.Source.ChannelFilter)
		require.NotNil(t, got.Target)
		require.NotNil(t, got.Target.Region)
		assert.Equal(t, "europe-west1", *got.Target.Region)
		require.NotNil(t, got.Target.ProjectID)
		assert.Equal(t, "updated-project", *got.Target.ProjectID)
		require.NotNil(t, got.Target.Enveloped)
		assert.False(t, *got.Target.Enveloped)
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule005", body)
	require.NoError(t, err)
	assert.Equal(t, "rule005", got.ID)
}

func TestUpdateRule_GoogleCloudFunctionPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := GoogleCloudFunctionRulePatch{RuleType: "http/google-cloud-function"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule005", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "http/google-cloud-function", raw["ruleType"])
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "requestMode")
		assert.NotContains(t, raw, "source")
		assert.NotContains(t, raw, "target")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule005", RuleType: "http/google-cloud-function"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule005", body)
	require.NoError(t, err)
	assert.Equal(t, "rule005", got.ID)
}
