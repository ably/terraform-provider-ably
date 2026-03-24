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
// Hive Text Model Only: CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_HiveTextModelOnly_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := HiveTextModelOnlyRulePost{
		Status:   "enabled",
		RuleType: "hive/text-model-only",
		BeforePublishConfig: BeforePublishConfig{
			RetryTimeout: 30, MaxRetries: 3,
			FailedAction: "reject", TooManyRequestsAction: "enqueue",
		},
		ChatRoomFilter: ".*",
		InvocationMode: "before-publish",
		Target: HiveTextModelOnlyTarget{
			APIKey:   "hive-key-123",
			ModelURL: "https://api.hive.ai/model",
			Thresholds: map[string]int{
				"hate": 80, "spam": 90, "violence": 70,
			},
		},
	}
	wantResp := RuleResponse{
		ID: "rule030", AppID: "app123", Status: "enabled", RuleType: "hive/text-model-only",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got HiveTextModelOnlyRulePost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "hive/text-model-only", got.RuleType)
		assert.Equal(t, "before-publish", got.InvocationMode)
		assert.Equal(t, ".*", got.ChatRoomFilter)
		assert.Equal(t, "hive-key-123", got.Target.APIKey)
		assert.Equal(t, "https://api.hive.ai/model", got.Target.ModelURL)
		assert.Equal(t, 80, got.Target.Thresholds["hate"])
		assert.Equal(t, 90, got.Target.Thresholds["spam"])
		assert.Equal(t, 30, got.BeforePublishConfig.RetryTimeout)
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule030", got.ID)
	assert.Equal(t, "hive/text-model-only", got.RuleType)
}

// ---------------------------------------------------------------------------
// Hive Text Model Only: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_HiveTextModelOnly_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := HiveTextModelOnlyRulePatch{
		Status:   "disabled",
		RuleType: "hive/text-model-only",
		BeforePublishConfig: &BeforePublishConfigPatch{
			RetryTimeout: ptr(10), MaxRetries: ptr(1), FailedAction: ptr("allow"),
		},
		InvocationMode: "before-publish",
		Target: &HiveTextModelOnlyTargetPatch{
			APIKey:     ptr("new-key"),
			Thresholds: map[string]int{"profanity": 50},
		},
	}
	wantResp := RuleResponse{
		ID: "rule030", AppID: "app123", Status: "disabled", RuleType: "hive/text-model-only",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule030", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got HiveTextModelOnlyRulePatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "hive/text-model-only", got.RuleType)
		assert.Equal(t, "disabled", got.Status)
		require.NotNil(t, got.Target)
		require.NotNil(t, got.Target.APIKey)
		assert.Equal(t, "new-key", *got.Target.APIKey)
		assert.Equal(t, 50, got.Target.Thresholds["profanity"])
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule030", body)
	require.NoError(t, err)
	assert.Equal(t, "rule030", got.ID)
	assert.Equal(t, "disabled", got.Status)
}

func TestUpdateRule_HiveTextModelOnlyPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := HiveTextModelOnlyRulePatch{RuleType: "hive/text-model-only"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule030", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "hive/text-model-only", raw["ruleType"])
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "beforePublishConfig")
		assert.NotContains(t, raw, "chatRoomFilter")
		assert.NotContains(t, raw, "invocationMode")
		assert.NotContains(t, raw, "target")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule030", RuleType: "hive/text-model-only"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule030", body)
	require.NoError(t, err)
	assert.Equal(t, "rule030", got.ID)
}

// ---------------------------------------------------------------------------
// Hive Text Model Only: Thresholds map[string]int edge cases
// ---------------------------------------------------------------------------

func TestCreateRule_HiveTextModelOnly_NilThresholds(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := HiveTextModelOnlyRulePost{
		RuleType: "hive/text-model-only",
		BeforePublishConfig: BeforePublishConfig{
			RetryTimeout: 10, MaxRetries: 1,
			FailedAction: "allow", TooManyRequestsAction: "allow",
		},
		InvocationMode: "before-publish",
		Target:         HiveTextModelOnlyTarget{APIKey: "k"},
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		target := raw["target"].(map[string]interface{})
		assert.NotContains(t, target, "thresholds")
		writeJSON(w, http.StatusCreated, RuleResponse{ID: "rule030", RuleType: "hive/text-model-only"})
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule030", got.ID)
}

// ---------------------------------------------------------------------------
// Hive Dashboard: CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_HiveDashboard_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := HiveDashboardRulePost{
		Status:         "enabled",
		RuleType:       "hive/dashboard",
		InvocationMode: "before-publish",
		ChatRoomFilter: "room-*",
		Target: HiveDashboardTarget{
			APIKey:          "dashboard-key",
			CheckWatchLists: ptr(true),
		},
	}
	wantResp := RuleResponse{
		ID: "rule031", AppID: "app123", Status: "enabled", RuleType: "hive/dashboard",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "hive/dashboard", raw["ruleType"])
		assert.Equal(t, "before-publish", raw["invocationMode"])
		assert.Equal(t, "room-*", raw["chatRoomFilter"])
		// Hive Dashboard must NOT have beforePublishConfig.
		assert.NotContains(t, raw, "beforePublishConfig")

		target := raw["target"].(map[string]interface{})
		assert.Equal(t, "dashboard-key", target["apiKey"])
		assert.Equal(t, true, target["checkWatchLists"])
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule031", got.ID)
	assert.Equal(t, "hive/dashboard", got.RuleType)
}

// ---------------------------------------------------------------------------
// Hive Dashboard: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_HiveDashboard_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := HiveDashboardRulePatch{
		RuleType: "hive/dashboard",
		Target:   &HiveDashboardTargetPatch{APIKey: ptr("updated-key"), CheckWatchLists: ptr(false)},
	}
	wantResp := RuleResponse{
		ID: "rule031", AppID: "app123", Status: "enabled", RuleType: "hive/dashboard",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule031", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "hive/dashboard", raw["ruleType"])
		target := raw["target"].(map[string]interface{})
		assert.Equal(t, "updated-key", target["apiKey"])
		assert.Equal(t, false, target["checkWatchLists"])
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule031", body)
	require.NoError(t, err)
	assert.Equal(t, "rule031", got.ID)
}

func TestUpdateRule_HiveDashboardPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := HiveDashboardRulePatch{RuleType: "hive/dashboard"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule031", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "hive/dashboard", raw["ruleType"])
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "invocationMode")
		assert.NotContains(t, raw, "chatRoomFilter")
		assert.NotContains(t, raw, "target")
		assert.NotContains(t, raw, "beforePublishConfig")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule031", RuleType: "hive/dashboard"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule031", body)
	require.NoError(t, err)
	assert.Equal(t, "rule031", got.ID)
}

// ---------------------------------------------------------------------------
// Hive Dashboard: nullable bool (CheckWatchLists)
// ---------------------------------------------------------------------------

func TestCreateRule_HiveDashboard_CheckWatchLists(t *testing.T) {
	t.Parallel()

	t.Run("nil omits field", func(t *testing.T) {
		t.Parallel()
		mux, client := newTestMux(t)

		body := HiveDashboardRulePost{
			RuleType:       "hive/dashboard",
			InvocationMode: "before-publish",
			Target:         HiveDashboardTarget{APIKey: "k"},
		}

		mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
			if !requireBearerToken(t, w, r, "test-token") {
				return
			}
			var raw map[string]interface{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
			target := raw["target"].(map[string]interface{})
			assert.NotContains(t, target, "checkWatchLists")
			writeJSON(w, http.StatusCreated, RuleResponse{ID: "rule031", RuleType: "hive/dashboard"})
		})

		_, err := client.CreateRule(context.Background(), "app123", body)
		require.NoError(t, err)
	})

	t.Run("true", func(t *testing.T) {
		t.Parallel()
		mux, client := newTestMux(t)

		body := HiveDashboardRulePost{
			RuleType:       "hive/dashboard",
			InvocationMode: "before-publish",
			Target:         HiveDashboardTarget{APIKey: "k", CheckWatchLists: ptr(true)},
		}

		mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
			if !requireBearerToken(t, w, r, "test-token") {
				return
			}
			var raw map[string]interface{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
			target := raw["target"].(map[string]interface{})
			assert.Equal(t, true, target["checkWatchLists"])
			writeJSON(w, http.StatusCreated, RuleResponse{ID: "rule031", RuleType: "hive/dashboard"})
		})

		_, err := client.CreateRule(context.Background(), "app123", body)
		require.NoError(t, err)
	})

	t.Run("false", func(t *testing.T) {
		t.Parallel()
		mux, client := newTestMux(t)

		body := HiveDashboardRulePost{
			RuleType:       "hive/dashboard",
			InvocationMode: "before-publish",
			Target:         HiveDashboardTarget{APIKey: "k", CheckWatchLists: ptr(false)},
		}

		mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
			if !requireBearerToken(t, w, r, "test-token") {
				return
			}
			var raw map[string]interface{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
			target := raw["target"].(map[string]interface{})
			assert.Equal(t, false, target["checkWatchLists"])
			writeJSON(w, http.StatusCreated, RuleResponse{ID: "rule031", RuleType: "hive/dashboard"})
		})

		_, err := client.CreateRule(context.Background(), "app123", body)
		require.NoError(t, err)
	})
}

// ---------------------------------------------------------------------------
// Bodyguard Text Moderation: CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_BodyguardTextModeration_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := BodyguardTextModerationRulePost{
		Status:   "enabled",
		RuleType: "bodyguard/text-moderation",
		BeforePublishConfig: BeforePublishConfig{
			RetryTimeout: 20, MaxRetries: 2,
			FailedAction: "reject", TooManyRequestsAction: "enqueue",
		},
		InvocationMode: "before-publish",
		ChatRoomFilter: "chat-*",
		Target: BodyguardTextModerationTarget{
			APIKey:          "bg-key",
			ChannelID:       "chan-1",
			APIURL:          "https://api.bodyguard.ai",
			DefaultLanguage: "en",
		},
	}
	wantResp := RuleResponse{
		ID: "rule032", AppID: "app123", Status: "enabled", RuleType: "bodyguard/text-moderation",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got BodyguardTextModerationRulePost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "bodyguard/text-moderation", got.RuleType)
		assert.Equal(t, "before-publish", got.InvocationMode)
		assert.Equal(t, "bg-key", got.Target.APIKey)
		assert.Equal(t, "chan-1", got.Target.ChannelID)
		assert.Equal(t, "https://api.bodyguard.ai", got.Target.APIURL)
		assert.Equal(t, "en", got.Target.DefaultLanguage)
		assert.Equal(t, 20, got.BeforePublishConfig.RetryTimeout)
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule032", got.ID)
	assert.Equal(t, "bodyguard/text-moderation", got.RuleType)
}

// ---------------------------------------------------------------------------
// Bodyguard Text Moderation: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_BodyguardTextModeration_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := BodyguardTextModerationRulePatch{
		Status:   "disabled",
		RuleType: "bodyguard/text-moderation",
		BeforePublishConfig: &BeforePublishConfigPatch{
			RetryTimeout: ptr(15), MaxRetries: ptr(5), FailedAction: ptr("allow"),
		},
		InvocationMode: "before-publish",
		Target: &BodyguardTextModerationTargetPatch{
			APIKey:          ptr("new-bg-key"),
			DefaultLanguage: ptr("fr"),
		},
	}
	wantResp := RuleResponse{
		ID: "rule032", AppID: "app123", Status: "disabled", RuleType: "bodyguard/text-moderation",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule032", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got BodyguardTextModerationRulePatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "bodyguard/text-moderation", got.RuleType)
		assert.Equal(t, "disabled", got.Status)
		require.NotNil(t, got.Target)
		require.NotNil(t, got.Target.APIKey)
		assert.Equal(t, "new-bg-key", *got.Target.APIKey)
		require.NotNil(t, got.Target.DefaultLanguage)
		assert.Equal(t, "fr", *got.Target.DefaultLanguage)
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule032", body)
	require.NoError(t, err)
	assert.Equal(t, "rule032", got.ID)
	assert.Equal(t, "disabled", got.Status)
}

func TestUpdateRule_BodyguardTextModerationPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := BodyguardTextModerationRulePatch{RuleType: "bodyguard/text-moderation"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule032", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "bodyguard/text-moderation", raw["ruleType"])
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "beforePublishConfig")
		assert.NotContains(t, raw, "invocationMode")
		assert.NotContains(t, raw, "chatRoomFilter")
		assert.NotContains(t, raw, "target")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule032", RuleType: "bodyguard/text-moderation"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule032", body)
	require.NoError(t, err)
	assert.Equal(t, "rule032", got.ID)
}

// ---------------------------------------------------------------------------
// Tisane Text Moderation: CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_TisaneTextModeration_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := TisaneTextModerationRulePost{
		Status:   "enabled",
		RuleType: "tisane/text-moderation",
		BeforePublishConfig: BeforePublishConfig{
			RetryTimeout: 25, MaxRetries: 4,
			FailedAction: "reject", TooManyRequestsAction: "enqueue",
		},
		InvocationMode: "before-publish",
		Target: TisaneTextModerationTarget{
			APIKey:          "tisane-key",
			ModelURL:        "https://api.tisane.ai/model",
			Thresholds:      map[string]int{"abuse": 60, "hate_speech": 75},
			DefaultLanguage: "en",
		},
	}
	wantResp := RuleResponse{
		ID: "rule033", AppID: "app123", Status: "enabled", RuleType: "tisane/text-moderation",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got TisaneTextModerationRulePost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "tisane/text-moderation", got.RuleType)
		assert.Equal(t, "tisane-key", got.Target.APIKey)
		assert.Equal(t, "en", got.Target.DefaultLanguage)
		assert.Equal(t, 60, got.Target.Thresholds["abuse"])
		assert.Equal(t, 75, got.Target.Thresholds["hate_speech"])
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule033", got.ID)
	assert.Equal(t, "tisane/text-moderation", got.RuleType)
}

// ---------------------------------------------------------------------------
// Tisane Text Moderation: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_TisaneTextModeration_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := TisaneTextModerationRulePatch{
		Status:   "disabled",
		RuleType: "tisane/text-moderation",
		BeforePublishConfig: &BeforePublishConfigPatch{
			RetryTimeout: ptr(10), MaxRetries: ptr(2),
			FailedAction: ptr("allow"), TooManyRequestsAction: ptr("reject"),
		},
		InvocationMode: "before-publish",
		ChatRoomFilter: "room-*",
		Target: &TisaneTextModerationTargetPatch{
			APIKey:          ptr("updated-key"),
			Thresholds:      map[string]int{"toxicity": 85},
			DefaultLanguage: ptr("de"),
		},
	}
	wantResp := RuleResponse{
		ID: "rule033", AppID: "app123", Status: "disabled", RuleType: "tisane/text-moderation",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule033", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got TisaneTextModerationRulePatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "tisane/text-moderation", got.RuleType)
		assert.Equal(t, "disabled", got.Status)
		require.NotNil(t, got.Target)
		require.NotNil(t, got.Target.APIKey)
		assert.Equal(t, "updated-key", *got.Target.APIKey)
		require.NotNil(t, got.Target.DefaultLanguage)
		assert.Equal(t, "de", *got.Target.DefaultLanguage)
		assert.Equal(t, 85, got.Target.Thresholds["toxicity"])
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule033", body)
	require.NoError(t, err)
	assert.Equal(t, "rule033", got.ID)
	assert.Equal(t, "disabled", got.Status)
}

func TestUpdateRule_TisaneTextModerationPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := TisaneTextModerationRulePatch{RuleType: "tisane/text-moderation"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule033", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "tisane/text-moderation", raw["ruleType"])
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "beforePublishConfig")
		assert.NotContains(t, raw, "invocationMode")
		assert.NotContains(t, raw, "chatRoomFilter")
		assert.NotContains(t, raw, "target")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule033", RuleType: "tisane/text-moderation"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule033", body)
	require.NoError(t, err)
	assert.Equal(t, "rule033", got.ID)
}

// ---------------------------------------------------------------------------
// Azure Text Moderation: CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_AzureTextModeration_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := AzureTextModerationRulePost{
		Status:   "enabled",
		RuleType: "azure/text-moderation",
		BeforePublishConfig: BeforePublishConfig{
			RetryTimeout: 30, MaxRetries: 3,
			FailedAction: "reject", TooManyRequestsAction: "enqueue",
		},
		InvocationMode: "before-publish",
		ChatRoomFilter: "public-*",
		Target: AzureTextModerationTarget{
			APIKey:     "azure-key-456",
			Endpoint:   "https://eastus.api.cognitive.microsoft.com",
			Thresholds: map[string]int{"sexual": 50, "violence": 60, "hate": 40},
		},
	}
	wantResp := RuleResponse{
		ID: "rule034", AppID: "app123", Status: "enabled", RuleType: "azure/text-moderation",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got AzureTextModerationRulePost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "azure/text-moderation", got.RuleType)
		assert.Equal(t, "before-publish", got.InvocationMode)
		assert.Equal(t, "azure-key-456", got.Target.APIKey)
		assert.Equal(t, "https://eastus.api.cognitive.microsoft.com", got.Target.Endpoint)
		assert.Equal(t, 50, got.Target.Thresholds["sexual"])
		assert.Equal(t, 60, got.Target.Thresholds["violence"])
		assert.Equal(t, 40, got.Target.Thresholds["hate"])
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule034", got.ID)
	assert.Equal(t, "azure/text-moderation", got.RuleType)
}

// ---------------------------------------------------------------------------
// Azure Text Moderation: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_AzureTextModeration_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := AzureTextModerationRulePatch{
		Status:   "disabled",
		RuleType: "azure/text-moderation",
		BeforePublishConfig: &BeforePublishConfigPatch{
			RetryTimeout: ptr(5), MaxRetries: ptr(1), FailedAction: ptr("allow"),
		},
		InvocationMode: "before-publish",
		ChatRoomFilter: "moderated-*",
		Target: &AzureTextModerationTargetPatch{
			APIKey:     ptr("new-azure-key"),
			Endpoint:   ptr("https://westus.api.cognitive.microsoft.com"),
			Thresholds: map[string]int{"self_harm": 30},
		},
	}
	wantResp := RuleResponse{
		ID: "rule034", AppID: "app123", Status: "disabled", RuleType: "azure/text-moderation",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule034", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got AzureTextModerationRulePatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "azure/text-moderation", got.RuleType)
		assert.Equal(t, "disabled", got.Status)
		require.NotNil(t, got.Target)
		require.NotNil(t, got.Target.APIKey)
		assert.Equal(t, "new-azure-key", *got.Target.APIKey)
		require.NotNil(t, got.Target.Endpoint)
		assert.Equal(t, "https://westus.api.cognitive.microsoft.com", *got.Target.Endpoint)
		assert.Equal(t, 30, got.Target.Thresholds["self_harm"])
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule034", body)
	require.NoError(t, err)
	assert.Equal(t, "rule034", got.ID)
	assert.Equal(t, "disabled", got.Status)
}

func TestUpdateRule_AzureTextModerationPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := AzureTextModerationRulePatch{RuleType: "azure/text-moderation"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule034", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "azure/text-moderation", raw["ruleType"])
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "beforePublishConfig")
		assert.NotContains(t, raw, "invocationMode")
		assert.NotContains(t, raw, "chatRoomFilter")
		assert.NotContains(t, raw, "target")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule034", RuleType: "azure/text-moderation"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule034", body)
	require.NoError(t, err)
	assert.Equal(t, "rule034", got.ID)
}
