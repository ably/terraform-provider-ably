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
// Before-Publish Webhook: CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_BeforePublishWebhook_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := BeforePublishWebhookRulePost{
		Status:   "enabled",
		RuleType: "http/before-publish",
		BeforePublishConfig: BeforePublishConfig{
			RetryTimeout:          30,
			MaxRetries:            5,
			FailedAction:          "reject",
			TooManyRequestsAction: "reject",
		},
		InvocationMode: "single",
		ChatRoomFilter: "room-*",
		Target: BeforePublishWebhookTarget{
			URL:     "https://example.com/hook",
			Headers: []RuleHeader{{Name: "X-Custom", Value: "abc"}},
		},
	}
	wantResp := RuleResponse{
		ID: "rule020", AppID: "app123", Status: "enabled", RuleType: "http/before-publish",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "http/before-publish", raw["ruleType"])
		assert.Equal(t, "single", raw["invocationMode"])
		assert.Equal(t, "room-*", raw["chatRoomFilter"])
		// Before-publish webhook should NOT have source or requestMode.
		assert.NotContains(t, raw, "source")
		assert.NotContains(t, raw, "requestMode")

		bpc := raw["beforePublishConfig"].(map[string]interface{})
		assert.Equal(t, float64(30), bpc["retryTimeout"])
		assert.Equal(t, float64(5), bpc["maxRetries"])
		assert.Equal(t, "reject", bpc["failedAction"])

		target := raw["target"].(map[string]interface{})
		assert.Equal(t, "https://example.com/hook", target["url"])
		// Target should NOT have format field.
		assert.NotContains(t, target, "format")
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule020", got.ID)
	assert.Equal(t, "http/before-publish", got.RuleType)
}

// ---------------------------------------------------------------------------
// Before-Publish Webhook: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_BeforePublishWebhook_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	bpc := BeforePublishConfig{
		RetryTimeout:          10,
		MaxRetries:            2,
		FailedAction:          "reject",
		TooManyRequestsAction: "reject",
	}
	body := BeforePublishWebhookRulePatch{
		Status:              "disabled",
		RuleType:            "http/before-publish",
		BeforePublishConfig: &bpc,
		InvocationMode:      "single",
		ChatRoomFilter:      "chat-*",
		Target: &BeforePublishWebhookTargetPatch{
			URL:     ptr("https://example.com/updated"),
			Headers: []RuleHeader{{Name: "Authorization", Value: "Bearer tok"}},
		},
	}
	wantResp := RuleResponse{
		ID: "rule020", AppID: "app123", Status: "disabled", RuleType: "http/before-publish",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule020", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got BeforePublishWebhookRulePatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "http/before-publish", got.RuleType)
		assert.Equal(t, "disabled", got.Status)
		require.NotNil(t, got.BeforePublishConfig)
		assert.Equal(t, 10, got.BeforePublishConfig.RetryTimeout)
		require.NotNil(t, got.Target)
		require.NotNil(t, got.Target.URL)
		assert.Equal(t, "https://example.com/updated", *got.Target.URL)
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule020", body)
	require.NoError(t, err)
	assert.Equal(t, "rule020", got.ID)
	assert.Equal(t, "disabled", got.Status)
}

func TestUpdateRule_BeforePublishWebhookPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := BeforePublishWebhookRulePatch{RuleType: "http/before-publish"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule020", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "http/before-publish", raw["ruleType"])
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "beforePublishConfig")
		assert.NotContains(t, raw, "invocationMode")
		assert.NotContains(t, raw, "chatRoomFilter")
		assert.NotContains(t, raw, "target")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule020", RuleType: "http/before-publish"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule020", body)
	require.NoError(t, err)
	assert.Equal(t, "rule020", got.ID)
}

// ---------------------------------------------------------------------------
// Before-Publish AWS Lambda: CreateRule (credentials mode)
// ---------------------------------------------------------------------------

func TestCreateRule_BeforePublishAWSLambda_Credentials(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := BeforePublishAWSLambdaRulePost{
		Status:   "enabled",
		RuleType: "aws/lambda/before-publish",
		BeforePublishConfig: BeforePublishConfig{
			RetryTimeout:          20,
			MaxRetries:            3,
			FailedAction:          "reject",
			TooManyRequestsAction: "allow",
		},
		InvocationMode: "single",
		Target: BeforePublishAWSLambdaTarget{
			Region:       "us-east-1",
			FunctionName: "my-function",
			Authentication: AWSAuthentication{
				AuthenticationMode: "credentials",
				AccessKeyID:        "AKIAIOSFODNN7EXAMPLE",
				SecretAccessKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
		},
	}
	wantResp := RuleResponse{
		ID: "rule021", AppID: "app123", Status: "enabled", RuleType: "aws/lambda/before-publish",
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "aws/lambda/before-publish", raw["ruleType"])
		assert.Equal(t, "single", raw["invocationMode"])

		target := raw["target"].(map[string]interface{})
		assert.Equal(t, "us-east-1", target["region"])
		assert.Equal(t, "my-function", target["functionName"])

		auth := target["authentication"].(map[string]interface{})
		assert.Equal(t, "credentials", auth["authenticationMode"])
		assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", auth["accessKeyId"])
		assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", auth["secretAccessKey"])
		// assumeRoleArn should be omitted in credentials mode.
		assert.NotContains(t, auth, "assumeRoleArn")
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule021", got.ID)
	assert.Equal(t, "aws/lambda/before-publish", got.RuleType)
}

// ---------------------------------------------------------------------------
// Before-Publish AWS Lambda: CreateRule (assumeRole mode)
// ---------------------------------------------------------------------------

func TestCreateRule_BeforePublishAWSLambda_AssumeRole(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := BeforePublishAWSLambdaRulePost{
		RuleType: "aws/lambda/before-publish",
		BeforePublishConfig: BeforePublishConfig{
			RetryTimeout:          10,
			MaxRetries:            1,
			FailedAction:          "allow",
			TooManyRequestsAction: "allow",
		},
		InvocationMode: "single",
		Target: BeforePublishAWSLambdaTarget{
			Region:       "eu-west-1",
			FunctionName: "other-function",
			Authentication: AWSAuthentication{
				AuthenticationMode: "assumeRole",
				AssumeRoleArn:      "arn:aws:iam::123456789012:role/my-role",
			},
		},
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))

		target := raw["target"].(map[string]interface{})
		auth := target["authentication"].(map[string]interface{})
		assert.Equal(t, "assumeRole", auth["authenticationMode"])
		assert.Equal(t, "arn:aws:iam::123456789012:role/my-role", auth["assumeRoleArn"])
		assert.NotContains(t, auth, "accessKeyId")
		assert.NotContains(t, auth, "secretAccessKey")
		writeJSON(w, http.StatusCreated, RuleResponse{ID: "rule021", RuleType: "aws/lambda/before-publish"})
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule021", got.ID)
}

// ---------------------------------------------------------------------------
// Before-Publish AWS Lambda: CreateRule (with optional Source)
// ---------------------------------------------------------------------------

func TestCreateRule_BeforePublishAWSLambda_WithSource(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := BeforePublishAWSLambdaRulePost{
		RuleType: "aws/lambda/before-publish",
		BeforePublishConfig: BeforePublishConfig{
			RetryTimeout: 10, MaxRetries: 1,
			FailedAction: "allow", TooManyRequestsAction: "allow",
		},
		InvocationMode: "single",
		Source:         &RuleSource{ChannelFilter: "^my-channel", Type: "channel.message"},
		Target: BeforePublishAWSLambdaTarget{
			Region: "us-west-2", FunctionName: "handler",
			Authentication: AWSAuthentication{
				AuthenticationMode: "credentials",
				AccessKeyID:        "AKID", SecretAccessKey: "SECRET",
			},
		},
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		source, ok := raw["source"].(map[string]interface{})
		require.True(t, ok, "source should be present when set")
		assert.Equal(t, "^my-channel", source["channelFilter"])
		assert.Equal(t, "channel.message", source["type"])
		writeJSON(w, http.StatusCreated, RuleResponse{ID: "rule021", RuleType: "aws/lambda/before-publish"})
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule021", got.ID)
}

// ---------------------------------------------------------------------------
// Before-Publish AWS Lambda: UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_BeforePublishAWSLambda_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	bpc := BeforePublishConfig{
		RetryTimeout: 20, MaxRetries: 4,
		FailedAction: "allow", TooManyRequestsAction: "reject",
	}
	body := BeforePublishAWSLambdaRulePatch{
		Status:              "disabled",
		RuleType:            "aws/lambda/before-publish",
		BeforePublishConfig: &bpc,
		InvocationMode:      "single",
		ChatRoomFilter:      "lobby-*",
		Source:              &RuleSource{ChannelFilter: "^filtered", Type: "channel.message"},
		Target: &BeforePublishAWSLambdaTargetPatch{
			Region: ptr("us-east-1"), FunctionName: ptr("updated-fn"),
			Authentication: &AWSAuthentication{
				AuthenticationMode: "assumeRole",
				AssumeRoleArn:      "arn:aws:iam::111111111111:role/test",
			},
		},
	}
	wantResp := RuleResponse{
		ID: "rule021", AppID: "app123", Status: "disabled", RuleType: "aws/lambda/before-publish",
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule021", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got BeforePublishAWSLambdaRulePatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "aws/lambda/before-publish", got.RuleType)
		assert.Equal(t, "disabled", got.Status)
		require.NotNil(t, got.BeforePublishConfig)
		assert.Equal(t, 20, got.BeforePublishConfig.RetryTimeout)
		require.NotNil(t, got.Source)
		assert.Equal(t, "^filtered", got.Source.ChannelFilter)
		require.NotNil(t, got.Target)
		require.NotNil(t, got.Target.Authentication)
		assert.Equal(t, "assumeRole", got.Target.Authentication.AuthenticationMode)
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule021", body)
	require.NoError(t, err)
	assert.Equal(t, "rule021", got.ID)
	assert.Equal(t, "disabled", got.Status)
}

func TestUpdateRule_BeforePublishAWSLambdaPatch_OmitEmpty(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := BeforePublishAWSLambdaRulePatch{RuleType: "aws/lambda/before-publish"}

	mux.HandleFunc("PATCH /apps/app123/rules/rule021", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var raw map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&raw))
		assert.Equal(t, "aws/lambda/before-publish", raw["ruleType"])
		assert.NotContains(t, raw, "status")
		assert.NotContains(t, raw, "beforePublishConfig")
		assert.NotContains(t, raw, "invocationMode")
		assert.NotContains(t, raw, "chatRoomFilter")
		assert.NotContains(t, raw, "source")
		assert.NotContains(t, raw, "target")
		writeJSON(w, http.StatusOK, RuleResponse{ID: "rule021", RuleType: "aws/lambda/before-publish"})
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule021", body)
	require.NoError(t, err)
	assert.Equal(t, "rule021", got.ID)
}
