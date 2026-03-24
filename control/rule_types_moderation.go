package control

// ---------------------------------------------------------------------------
// Hive Text Model Only (ruleType: "hive/text-model-only")
// ---------------------------------------------------------------------------

// HiveTextModelOnlyTarget is the target configuration for Hive text-model-only rules.
type HiveTextModelOnlyTarget struct {
	APIKey     string         `json:"apiKey"`
	ModelURL   string         `json:"modelUrl,omitempty"`
	Thresholds map[string]int `json:"thresholds,omitempty"`
}

// HiveTextModelOnlyRulePost is the request body for creating a Hive text-model-only rule.
type HiveTextModelOnlyRulePost struct {
	Status              string                  `json:"status,omitempty"`
	RuleType            string                  `json:"ruleType"`
	BeforePublishConfig BeforePublishConfig     `json:"beforePublishConfig"`
	ChatRoomFilter      string                  `json:"chatRoomFilter,omitempty"`
	InvocationMode      string                  `json:"invocationMode"`
	Target              HiveTextModelOnlyTarget `json:"target"`
}

// HiveTextModelOnlyTargetPatch is the patch-specific target for Hive text-model-only rules.
// Fields use pointer types so that omitted fields are not sent in the PATCH request.
type HiveTextModelOnlyTargetPatch struct {
	APIKey     *string        `json:"apiKey,omitempty"`
	ModelURL   *string        `json:"modelUrl,omitempty"`
	Thresholds map[string]int `json:"thresholds,omitempty"`
}

// HiveTextModelOnlyRulePatch is the request body for updating a Hive text-model-only rule.
type HiveTextModelOnlyRulePatch struct {
	Status              string                        `json:"status,omitempty"`
	RuleType            string                        `json:"ruleType"`
	BeforePublishConfig *BeforePublishConfigPatch     `json:"beforePublishConfig,omitempty"`
	ChatRoomFilter      string                        `json:"chatRoomFilter,omitempty"`
	InvocationMode      string                        `json:"invocationMode,omitempty"`
	Target              *HiveTextModelOnlyTargetPatch `json:"target,omitempty"`
}

// ---------------------------------------------------------------------------
// Hive Dashboard (ruleType: "hive/dashboard")
// ---------------------------------------------------------------------------

// HiveDashboardTarget is the target configuration for Hive dashboard rules.
type HiveDashboardTarget struct {
	APIKey          string `json:"apiKey"`
	CheckWatchLists *bool  `json:"checkWatchLists,omitempty"`
}

// HiveDashboardRulePost is the request body for creating a Hive dashboard rule.
type HiveDashboardRulePost struct {
	Status         string              `json:"status,omitempty"`
	RuleType       string              `json:"ruleType"`
	InvocationMode string              `json:"invocationMode"`
	ChatRoomFilter string              `json:"chatRoomFilter,omitempty"`
	Target         HiveDashboardTarget `json:"target"`
}

// HiveDashboardTargetPatch is the patch-specific target for Hive dashboard rules.
// Fields use pointer types so that omitted fields are not sent in the PATCH request.
type HiveDashboardTargetPatch struct {
	APIKey          *string `json:"apiKey,omitempty"`
	CheckWatchLists *bool   `json:"checkWatchLists,omitempty"`
}

// HiveDashboardRulePatch is the request body for updating a Hive dashboard rule.
type HiveDashboardRulePatch struct {
	Status         string                    `json:"status,omitempty"`
	RuleType       string                    `json:"ruleType"`
	InvocationMode string                    `json:"invocationMode,omitempty"`
	ChatRoomFilter string                    `json:"chatRoomFilter,omitempty"`
	Target         *HiveDashboardTargetPatch `json:"target,omitempty"`
}

// ---------------------------------------------------------------------------
// Bodyguard Text Moderation (ruleType: "bodyguard/text-moderation")
// ---------------------------------------------------------------------------

// BodyguardTextModerationTarget is the target configuration for Bodyguard text moderation rules.
type BodyguardTextModerationTarget struct {
	APIKey          string `json:"apiKey"`
	ChannelID       string `json:"channelId,omitempty"`
	APIURL          string `json:"apiUrl,omitempty"`
	DefaultLanguage string `json:"defaultLanguage,omitempty"`
}

// BodyguardTextModerationRulePost is the request body for creating a Bodyguard text moderation rule.
type BodyguardTextModerationRulePost struct {
	Status              string                        `json:"status,omitempty"`
	RuleType            string                        `json:"ruleType"`
	BeforePublishConfig BeforePublishConfig           `json:"beforePublishConfig"`
	InvocationMode      string                        `json:"invocationMode"`
	ChatRoomFilter      string                        `json:"chatRoomFilter,omitempty"`
	Target              BodyguardTextModerationTarget `json:"target"`
}

// BodyguardTextModerationTargetPatch is the patch-specific target for Bodyguard text moderation rules.
// Fields use pointer types so that omitted fields are not sent in the PATCH request.
type BodyguardTextModerationTargetPatch struct {
	APIKey          *string `json:"apiKey,omitempty"`
	ChannelID       *string `json:"channelId,omitempty"`
	APIURL          *string `json:"apiUrl,omitempty"`
	DefaultLanguage *string `json:"defaultLanguage,omitempty"`
}

// BodyguardTextModerationRulePatch is the request body for updating a Bodyguard text moderation rule.
type BodyguardTextModerationRulePatch struct {
	Status              string                              `json:"status,omitempty"`
	RuleType            string                              `json:"ruleType"`
	BeforePublishConfig *BeforePublishConfigPatch           `json:"beforePublishConfig,omitempty"`
	InvocationMode      string                              `json:"invocationMode,omitempty"`
	ChatRoomFilter      string                              `json:"chatRoomFilter,omitempty"`
	Target              *BodyguardTextModerationTargetPatch `json:"target,omitempty"`
}

// ---------------------------------------------------------------------------
// Tisane Text Moderation (ruleType: "tisane/text-moderation")
// ---------------------------------------------------------------------------

// TisaneTextModerationTarget is the target configuration for Tisane text moderation rules.
type TisaneTextModerationTarget struct {
	APIKey          string         `json:"apiKey"`
	ModelURL        string         `json:"modelUrl,omitempty"`
	Thresholds      map[string]int `json:"thresholds,omitempty"`
	DefaultLanguage string         `json:"defaultLanguage"`
}

// TisaneTextModerationRulePost is the request body for creating a Tisane text moderation rule.
type TisaneTextModerationRulePost struct {
	Status              string                     `json:"status,omitempty"`
	RuleType            string                     `json:"ruleType"`
	BeforePublishConfig BeforePublishConfig        `json:"beforePublishConfig"`
	InvocationMode      string                     `json:"invocationMode"`
	ChatRoomFilter      string                     `json:"chatRoomFilter,omitempty"`
	Target              TisaneTextModerationTarget `json:"target"`
}

// TisaneTextModerationTargetPatch is the patch-specific target for Tisane text moderation rules.
// Fields use pointer types so that omitted fields are not sent in the PATCH request.
type TisaneTextModerationTargetPatch struct {
	APIKey          *string        `json:"apiKey,omitempty"`
	ModelURL        *string        `json:"modelUrl,omitempty"`
	Thresholds      map[string]int `json:"thresholds,omitempty"`
	DefaultLanguage *string        `json:"defaultLanguage,omitempty"`
}

// TisaneTextModerationRulePatch is the request body for updating a Tisane text moderation rule.
type TisaneTextModerationRulePatch struct {
	Status              string                           `json:"status,omitempty"`
	RuleType            string                           `json:"ruleType"`
	BeforePublishConfig *BeforePublishConfigPatch        `json:"beforePublishConfig,omitempty"`
	InvocationMode      string                           `json:"invocationMode,omitempty"`
	ChatRoomFilter      string                           `json:"chatRoomFilter,omitempty"`
	Target              *TisaneTextModerationTargetPatch `json:"target,omitempty"`
}

// ---------------------------------------------------------------------------
// Azure Text Moderation (ruleType: "azure/text-moderation")
// ---------------------------------------------------------------------------

// AzureTextModerationTarget is the target configuration for Azure text moderation rules.
type AzureTextModerationTarget struct {
	APIKey     string         `json:"apiKey"`
	Endpoint   string         `json:"endpoint"`
	Thresholds map[string]int `json:"thresholds,omitempty"`
}

// AzureTextModerationRulePost is the request body for creating an Azure text moderation rule.
type AzureTextModerationRulePost struct {
	Status              string                    `json:"status,omitempty"`
	RuleType            string                    `json:"ruleType"`
	BeforePublishConfig BeforePublishConfig       `json:"beforePublishConfig"`
	InvocationMode      string                    `json:"invocationMode"`
	ChatRoomFilter      string                    `json:"chatRoomFilter,omitempty"`
	Target              AzureTextModerationTarget `json:"target"`
}

// AzureTextModerationTargetPatch is the patch-specific target for Azure text moderation rules.
// Fields use pointer types so that omitted fields are not sent in the PATCH request.
type AzureTextModerationTargetPatch struct {
	APIKey     *string        `json:"apiKey,omitempty"`
	Endpoint   *string        `json:"endpoint,omitempty"`
	Thresholds map[string]int `json:"thresholds,omitempty"`
}

// AzureTextModerationRulePatch is the request body for updating an Azure text moderation rule.
type AzureTextModerationRulePatch struct {
	Status              string                          `json:"status,omitempty"`
	RuleType            string                          `json:"ruleType"`
	BeforePublishConfig *BeforePublishConfigPatch       `json:"beforePublishConfig,omitempty"`
	InvocationMode      string                          `json:"invocationMode,omitempty"`
	ChatRoomFilter      string                          `json:"chatRoomFilter,omitempty"`
	Target              *AzureTextModerationTargetPatch `json:"target,omitempty"`
}
