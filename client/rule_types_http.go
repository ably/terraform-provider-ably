package ably

// RuleSourcePatch is a partial source update (all fields optional).
type RuleSourcePatch struct {
	ChannelFilter string `json:"channelFilter,omitempty"`
	Type          string `json:"type,omitempty"`
}

// ---------------------------------------------------------------------------
// IFTTT (ruleType: "http/ifttt")
// ---------------------------------------------------------------------------

// IFTTTRuleTarget is the target configuration for IFTTT rules.
type IFTTTRuleTarget struct {
	WebhookKey string `json:"webhookKey"`
	EventName  string `json:"eventName"`
}

// IFTTTRulePost is the request body for creating an IFTTT rule.
type IFTTTRulePost struct {
	Status      string          `json:"status,omitempty"`
	RuleType    string          `json:"ruleType"`
	RequestMode string          `json:"requestMode"`
	Source      RuleSource      `json:"source"`
	Target      IFTTTRuleTarget `json:"target"`
}

// IFTTTRuleTargetPatch is the patch-specific target for IFTTT rules.
// Fields use pointer types so that omitted fields are not sent in the PATCH request.
type IFTTTRuleTargetPatch struct {
	WebhookKey *string `json:"webhookKey,omitempty"`
	EventName  *string `json:"eventName,omitempty"`
}

// IFTTTRulePatch is the request body for updating an IFTTT rule.
type IFTTTRulePatch struct {
	Status      string                `json:"status,omitempty"`
	RuleType    string                `json:"ruleType"`
	RequestMode string                `json:"requestMode,omitempty"`
	Source      *RuleSourcePatch      `json:"source,omitempty"`
	Target      *IFTTTRuleTargetPatch `json:"target,omitempty"`
}

// ---------------------------------------------------------------------------
// Zapier (ruleType: "http/zapier")
// ---------------------------------------------------------------------------

// ZapierRuleTarget is the target configuration for Zapier rules.
type ZapierRuleTarget struct {
	URL          string       `json:"url"`
	Headers      []RuleHeader `json:"headers,omitempty"`
	SigningKeyID *string      `json:"signingKeyId,omitempty"`
}

// ZapierRulePost is the request body for creating a Zapier rule.
type ZapierRulePost struct {
	Status      string           `json:"status,omitempty"`
	RuleType    string           `json:"ruleType"`
	RequestMode string           `json:"requestMode"`
	Source      RuleSource       `json:"source"`
	Target      ZapierRuleTarget `json:"target"`
}

// ZapierRuleTargetPatch is the patch-specific target for Zapier rules.
// Fields use pointer types so that omitted fields are not sent in the PATCH request.
type ZapierRuleTargetPatch struct {
	URL          *string      `json:"url,omitempty"`
	Headers      []RuleHeader `json:"headers,omitempty"`
	SigningKeyID *string      `json:"signingKeyId,omitempty"`
}

// ZapierRulePatch is the request body for updating a Zapier rule.
type ZapierRulePatch struct {
	Status      string                 `json:"status,omitempty"`
	RuleType    string                 `json:"ruleType"`
	RequestMode string                 `json:"requestMode,omitempty"`
	Source      *RuleSourcePatch       `json:"source,omitempty"`
	Target      *ZapierRuleTargetPatch `json:"target,omitempty"`
}

// ---------------------------------------------------------------------------
// Cloudflare Worker (ruleType: "http/cloudflare-worker")
// ---------------------------------------------------------------------------

// CloudflareWorkerRuleTarget is the target configuration for Cloudflare Worker rules.
type CloudflareWorkerRuleTarget struct {
	URL          string       `json:"url"`
	Headers      []RuleHeader `json:"headers,omitempty"`
	SigningKeyID *string      `json:"signingKeyId,omitempty"`
}

// CloudflareWorkerRulePost is the request body for creating a Cloudflare Worker rule.
type CloudflareWorkerRulePost struct {
	Status      string                     `json:"status,omitempty"`
	RuleType    string                     `json:"ruleType"`
	RequestMode string                     `json:"requestMode"`
	Source      RuleSource                 `json:"source"`
	Target      CloudflareWorkerRuleTarget `json:"target"`
}

// CloudflareWorkerRuleTargetPatch is the patch-specific target for Cloudflare Worker rules.
// Fields use pointer types so that omitted fields are not sent in the PATCH request.
type CloudflareWorkerRuleTargetPatch struct {
	URL          *string      `json:"url,omitempty"`
	Headers      []RuleHeader `json:"headers,omitempty"`
	SigningKeyID *string      `json:"signingKeyId,omitempty"`
}

// CloudflareWorkerRulePatch is the request body for updating a Cloudflare Worker rule.
type CloudflareWorkerRulePatch struct {
	Status      string                           `json:"status,omitempty"`
	RuleType    string                           `json:"ruleType"`
	RequestMode string                           `json:"requestMode,omitempty"`
	Source      *RuleSourcePatch                 `json:"source,omitempty"`
	Target      *CloudflareWorkerRuleTargetPatch `json:"target,omitempty"`
}

// ---------------------------------------------------------------------------
// Azure Function (ruleType: "http/azure-function")
// ---------------------------------------------------------------------------

// AzureFunctionRuleTarget is the target configuration for Azure Function rules.
type AzureFunctionRuleTarget struct {
	AzureAppID        string       `json:"azureAppId"`
	AzureFunctionName string       `json:"azureFunctionName"`
	Headers           []RuleHeader `json:"headers,omitempty"`
	SigningKeyID      *string      `json:"signingKeyId,omitempty"`
	Enveloped         *bool        `json:"enveloped,omitempty"`
	Format            string       `json:"format,omitempty"`
}

// AzureFunctionRulePost is the request body for creating an Azure Function rule.
type AzureFunctionRulePost struct {
	Status      string                  `json:"status,omitempty"`
	RuleType    string                  `json:"ruleType"`
	RequestMode string                  `json:"requestMode"`
	Source      RuleSource              `json:"source"`
	Target      AzureFunctionRuleTarget `json:"target"`
}

// AzureFunctionRuleTargetPatch is the patch-specific target for Azure Function rules.
// Fields use pointer types so that omitted fields are not sent in the PATCH request.
type AzureFunctionRuleTargetPatch struct {
	AzureAppID        *string      `json:"azureAppId,omitempty"`
	AzureFunctionName *string      `json:"azureFunctionName,omitempty"`
	Headers           []RuleHeader `json:"headers,omitempty"`
	SigningKeyID      *string      `json:"signingKeyId,omitempty"`
	Enveloped         *bool        `json:"enveloped,omitempty"`
	Format            *string      `json:"format,omitempty"`
}

// AzureFunctionRulePatch is the request body for updating an Azure Function rule.
type AzureFunctionRulePatch struct {
	Status      string                        `json:"status,omitempty"`
	RuleType    string                        `json:"ruleType"`
	RequestMode string                        `json:"requestMode,omitempty"`
	Source      *RuleSourcePatch              `json:"source,omitempty"`
	Target      *AzureFunctionRuleTargetPatch `json:"target,omitempty"`
}

// ---------------------------------------------------------------------------
// Google Cloud Function (ruleType: "http/google-cloud-function")
// ---------------------------------------------------------------------------

// GoogleCloudFunctionRuleTarget is the target configuration for Google Cloud Function rules.
type GoogleCloudFunctionRuleTarget struct {
	Region       string       `json:"region"`
	ProjectID    string       `json:"projectId"`
	FunctionName string       `json:"functionName"`
	Headers      []RuleHeader `json:"headers,omitempty"`
	SigningKeyID *string      `json:"signingKeyId,omitempty"`
	Enveloped    *bool        `json:"enveloped,omitempty"`
	Format       string       `json:"format,omitempty"`
}

// GoogleCloudFunctionRulePost is the request body for creating a Google Cloud Function rule.
type GoogleCloudFunctionRulePost struct {
	Status      string                        `json:"status,omitempty"`
	RuleType    string                        `json:"ruleType"`
	RequestMode string                        `json:"requestMode"`
	Source      RuleSource                    `json:"source"`
	Target      GoogleCloudFunctionRuleTarget `json:"target"`
}

// GoogleCloudFunctionRuleTargetPatch is the patch-specific target for Google Cloud Function rules.
// Fields use pointer types so that omitted fields are not sent in the PATCH request.
type GoogleCloudFunctionRuleTargetPatch struct {
	Region       *string      `json:"region,omitempty"`
	ProjectID    *string      `json:"projectId,omitempty"`
	FunctionName *string      `json:"functionName,omitempty"`
	Headers      []RuleHeader `json:"headers,omitempty"`
	SigningKeyID *string      `json:"signingKeyId,omitempty"`
	Enveloped    *bool        `json:"enveloped,omitempty"`
	Format       *string      `json:"format,omitempty"`
}

// GoogleCloudFunctionRulePatch is the request body for updating a Google Cloud Function rule.
type GoogleCloudFunctionRulePatch struct {
	Status      string                              `json:"status,omitempty"`
	RuleType    string                              `json:"ruleType"`
	RequestMode string                              `json:"requestMode,omitempty"`
	Source      *RuleSourcePatch                    `json:"source,omitempty"`
	Target      *GoogleCloudFunctionRuleTargetPatch `json:"target,omitempty"`
}
