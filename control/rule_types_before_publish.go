package control

// ---------------------------------------------------------------------------
// AWSAuthentication (nested auth object for before-publish AWS rules)
// ---------------------------------------------------------------------------

// AWSAuthentication holds AWS authentication as a nested object matching the API schema.
type AWSAuthentication struct {
	AuthenticationMode string `json:"authenticationMode"`
	AccessKeyID        string `json:"accessKeyId,omitempty"`
	SecretAccessKey    string `json:"secretAccessKey,omitempty"`
	AssumeRoleArn      string `json:"assumeRoleArn,omitempty"`
}

// AWSAuthenticationPatch is the patch-safe variant of AWSAuthentication.
// All fields are pointer types with omitempty so that omitted fields are not
// serialized, preventing partial PATCH updates from overwriting existing values.
type AWSAuthenticationPatch struct {
	AuthenticationMode *string `json:"authenticationMode,omitempty"`
	AccessKeyID        *string `json:"accessKeyId,omitempty"`
	SecretAccessKey    *string `json:"secretAccessKey,omitempty"`
	AssumeRoleArn      *string `json:"assumeRoleArn,omitempty"`
}

// ---------------------------------------------------------------------------
// Before-Publish Webhook (ruleType: "http/before-publish")
// ---------------------------------------------------------------------------

// BeforePublishWebhookTarget is the target configuration for before-publish webhook rules.
type BeforePublishWebhookTarget struct {
	URL     string       `json:"url"`
	Headers []RuleHeader `json:"headers,omitempty"`
}

// BeforePublishWebhookRulePost is the request body for creating a before-publish webhook rule.
type BeforePublishWebhookRulePost struct {
	Status              string                     `json:"status,omitempty"`
	RuleType            string                     `json:"ruleType"`
	BeforePublishConfig BeforePublishConfig        `json:"beforePublishConfig"`
	InvocationMode      string                     `json:"invocationMode"`
	ChatRoomFilter      string                     `json:"chatRoomFilter,omitempty"`
	Target              BeforePublishWebhookTarget `json:"target"`
}

// BeforePublishWebhookTargetPatch is the patch-specific target for before-publish webhook rules.
// Fields use pointer types so that omitted fields are not sent in the PATCH request.
type BeforePublishWebhookTargetPatch struct {
	URL     *string      `json:"url,omitempty"`
	Headers []RuleHeader `json:"headers,omitempty"`
}

// BeforePublishWebhookRulePatch is the request body for updating a before-publish webhook rule.
type BeforePublishWebhookRulePatch struct {
	Status              string                           `json:"status,omitempty"`
	RuleType            string                           `json:"ruleType"`
	BeforePublishConfig *BeforePublishConfigPatch        `json:"beforePublishConfig,omitempty"`
	InvocationMode      string                           `json:"invocationMode,omitempty"`
	ChatRoomFilter      string                           `json:"chatRoomFilter,omitempty"`
	Target              *BeforePublishWebhookTargetPatch `json:"target,omitempty"`
}

// ---------------------------------------------------------------------------
// Before-Publish AWS Lambda (ruleType: "aws/lambda/before-publish")
// ---------------------------------------------------------------------------

// BeforePublishAWSLambdaTarget is the target configuration for before-publish AWS Lambda rules.
type BeforePublishAWSLambdaTarget struct {
	Region         string            `json:"region"`
	FunctionName   string            `json:"functionName"`
	Authentication AWSAuthentication `json:"authentication"`
}

// BeforePublishAWSLambdaRulePost is the request body for creating a before-publish AWS Lambda rule.
type BeforePublishAWSLambdaRulePost struct {
	Status              string                       `json:"status,omitempty"`
	RuleType            string                       `json:"ruleType"`
	BeforePublishConfig BeforePublishConfig          `json:"beforePublishConfig"`
	InvocationMode      string                       `json:"invocationMode"`
	ChatRoomFilter      string                       `json:"chatRoomFilter,omitempty"`
	Source              *RuleSource                  `json:"source,omitempty"`
	Target              BeforePublishAWSLambdaTarget `json:"target"`
}

// BeforePublishAWSLambdaTargetPatch is the patch-specific target for before-publish AWS Lambda rules.
// Fields use pointer types so that omitted fields are not sent in the PATCH request.
type BeforePublishAWSLambdaTargetPatch struct {
	Region         *string                 `json:"region,omitempty"`
	FunctionName   *string                 `json:"functionName,omitempty"`
	Authentication *AWSAuthenticationPatch `json:"authentication,omitempty"`
}

// BeforePublishAWSLambdaRulePatch is the request body for updating a before-publish AWS Lambda rule.
type BeforePublishAWSLambdaRulePatch struct {
	Status              string                             `json:"status,omitempty"`
	RuleType            string                             `json:"ruleType"`
	BeforePublishConfig *BeforePublishConfigPatch          `json:"beforePublishConfig,omitempty"`
	InvocationMode      string                             `json:"invocationMode,omitempty"`
	ChatRoomFilter      string                             `json:"chatRoomFilter,omitempty"`
	Source              *RuleSource                        `json:"source,omitempty"`
	Target              *BeforePublishAWSLambdaTargetPatch `json:"target,omitempty"`
}
