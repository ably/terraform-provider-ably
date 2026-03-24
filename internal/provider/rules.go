// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func GetPlanAwsAuth(plan AblyRule) control.AWSAuthentication {
	var auth AwsAuth

	switch t := plan.Target.(type) {
	case *AblyRuleTargetKinesis:
		if t != nil {
			auth = t.AwsAuth
		}
	case *AblyRuleTargetSqs:
		if t != nil {
			auth = t.AwsAuth
		}
	case *AblyRuleTargetLambda:
		if t != nil {
			auth = t.AwsAuth
		}
	}

	var controlAuth control.AWSAuthentication
	if auth.AuthenticationMode.ValueString() == "assumeRole" {
		controlAuth = control.AWSAuthentication{
			AuthenticationMode: string(control.AWSAuthModeAssumeRole),
			AssumeRoleArn:      auth.RoleArn.ValueString(),
		}
	} else if auth.AuthenticationMode.ValueString() == "credentials" {
		controlAuth = control.AWSAuthentication{
			AuthenticationMode: string(control.AWSAuthModeCredentials),
			AccessKeyID:        auth.AccessKeyId.ValueString(),
			SecretAccessKey:    auth.SecretAccessKey.ValueString(),
		}
	}

	return controlAuth
}

// webhookEnveloped returns the enveloped *bool for HTTP-type webhook targets.
// In batch mode the API rejects enveloped=true, so we force it to false to
// ensure any existing enveloped value is cleared on update (PATCH).
func webhookEnveloped(enveloped types.Bool, requestMode string) *bool {
	if requestMode == "batch" {
		return ptr(false)
	}
	return ptr(enveloped.ValueBool())
}

// GetPlanRule converts rule from terraform format to the appropriate typed rule post struct.
// Returns (nil, diagnostics) with an error diagnostic if the target type is unrecognized.
func GetPlanRule(plan AblyRule) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	source := control.RuleSource{
		ChannelFilter: plan.Source.ChannelFilter.ValueString(),
		Type:          plan.Source.Type.ValueString(),
	}
	status := plan.Status.ValueString()
	requestMode := GetRequestMode(plan)

	switch t := plan.Target.(type) {
	case *AblyRuleTargetKinesis:
		if requestMode != "single" {
			diags.AddError(
				"Unsupported request_mode",
				"aws/kinesis rules do not support request_mode.",
			)
			return nil, diags
		}
		return control.AWSKinesisRulePost{
			Status:      status,
			RuleType:    "aws/kinesis",
			RequestMode: requestMode,
			Source:      source,
			Target: control.AWSKinesisTarget{
				Region:         t.Region.ValueString(),
				StreamName:     t.StreamName.ValueString(),
				PartitionKey:   t.PartitionKey.ValueString(),
				Enveloped:      ptr(t.Enveloped.ValueBool()),
				Format:         t.Format.ValueString(),
				Authentication: GetPlanAwsAuth(plan),
			},
		}, diags
	case *AblyRuleTargetSqs:
		if requestMode != "single" {
			diags.AddError(
				"Unsupported request_mode",
				"aws/sqs rules do not support request_mode.",
			)
			return nil, diags
		}
		return control.AWSSQSRulePost{
			Status:      status,
			RuleType:    "aws/sqs",
			RequestMode: requestMode,
			Source:      source,
			Target: control.AWSSQSTarget{
				Region:         t.Region.ValueString(),
				AWSAccountID:   t.AwsAccountID.ValueString(),
				QueueName:      t.QueueName.ValueString(),
				Enveloped:      ptr(t.Enveloped.ValueBool()),
				Format:         t.Format.ValueString(),
				Authentication: GetPlanAwsAuth(plan),
			},
		}, diags
	case *AblyRuleTargetLambda:
		return control.AWSLambdaRulePost{
			Status:      status,
			RuleType:    "aws/lambda",
			RequestMode: requestMode,
			Source:      source,
			Target: control.AWSLambdaTarget{
				Region:         t.Region.ValueString(),
				FunctionName:   t.FunctionName.ValueString(),
				Enveloped:      ptr(t.Enveloped.ValueBool()),
				Authentication: GetPlanAwsAuth(plan),
			},
		}, diags
	case *AblyRuleTargetZapier:
		return control.ZapierRulePost{
			Status:      status,
			RuleType:    "http/zapier",
			RequestMode: requestMode,
			Source:      source,
			Target: control.ZapierRuleTarget{
				URL:          t.Url.ValueString(),
				Headers:      GetHeaders(t.Headers),
				SigningKeyID: optionalStringPtr(t.SigningKeyId),
			},
		}, diags
	case *AblyRuleTargetCloudflareWorker:
		return control.CloudflareWorkerRulePost{
			Status:      status,
			RuleType:    "http/cloudflare-worker",
			RequestMode: requestMode,
			Source:      source,
			Target: control.CloudflareWorkerRuleTarget{
				URL:          t.Url.ValueString(),
				Headers:      GetHeaders(t.Headers),
				SigningKeyID: optionalStringPtr(t.SigningKeyId),
			},
		}, diags
	case *AblyRuleTargetPulsar:
		return control.PulsarRulePost{
			Status:      status,
			RuleType:    "pulsar",
			RequestMode: requestMode,
			Source:      source,
			Target: control.PulsarRuleTarget{
				RoutingKey:    t.RoutingKey.ValueString(),
				Topic:         t.Topic.ValueString(),
				ServiceURL:    t.ServiceURL.ValueString(),
				TLSTrustCerts: sliceString(t.TlsTrustCerts),
				Authentication: &control.PulsarAuth{
					AuthenticationMode: t.Authentication.Mode.ValueString(),
					Token:              t.Authentication.Token.ValueString(),
				},
				Enveloped: ptr(t.Enveloped.ValueBool()),
				Format:    t.Format.ValueString(),
			},
		}, diags
	case *AblyRuleTargetHTTP:
		return control.HTTPRulePost{
			Status:      status,
			RuleType:    "http",
			RequestMode: requestMode,
			Source:      source,
			Target: control.HTTPRuleTarget{
				URL:          t.Url.ValueString(),
				Headers:      GetHeaders(t.Headers),
				SigningKeyID: optionalStringPtr(t.SigningKeyId),
				Format:       t.Format.ValueString(),
				Enveloped:    webhookEnveloped(t.Enveloped, requestMode),
			},
		}, diags
	case *AblyRuleTargetIFTTT:
		return control.IFTTTRulePost{
			Status:      status,
			RuleType:    "http/ifttt",
			RequestMode: requestMode,
			Source:      source,
			Target: control.IFTTTRuleTarget{
				WebhookKey: t.WebhookKey.ValueString(),
				EventName:  t.EventName.ValueString(),
			},
		}, diags
	case *AblyRuleTargetAzureFunction:
		return control.AzureFunctionRulePost{
			Status:      status,
			RuleType:    "http/azure-function",
			RequestMode: requestMode,
			Source:      source,
			Target: control.AzureFunctionRuleTarget{
				AzureAppID:        t.AzureAppID.ValueString(),
				AzureFunctionName: t.AzureFunctionName.ValueString(),
				Headers:           GetHeaders(t.Headers),
				SigningKeyID:      optionalStringPtr(t.SigningKeyID),
				Enveloped:         webhookEnveloped(t.Enveloped, requestMode),
				Format:            t.Format.ValueString(),
			},
		}, diags
	case *AblyRuleTargetGoogleFunction:
		return control.GoogleCloudFunctionRulePost{
			Status:      status,
			RuleType:    "http/google-cloud-function",
			RequestMode: requestMode,
			Source:      source,
			Target: control.GoogleCloudFunctionRuleTarget{
				Region:       t.Region.ValueString(),
				ProjectID:    t.ProjectID.ValueString(),
				FunctionName: t.FunctionName.ValueString(),
				Headers:      GetHeaders(t.Headers),
				SigningKeyID: optionalStringPtr(t.SigningKeyId),
				Enveloped:    webhookEnveloped(t.Enveloped, requestMode),
				Format:       t.Format.ValueString(),
			},
		}, diags
	case *AblyRuleTargetKafka:
		return control.KafkaRulePost{
			Status:      status,
			RuleType:    "kafka",
			RequestMode: requestMode,
			Source:      source,
			Target: control.KafkaRuleTarget{
				RoutingKey: t.RoutingKey.ValueString(),
				Brokers:    sliceString(t.Brokers),
				Auth: &control.KafkaAuth{
					SASL: &control.KafkaSASL{
						Mechanism: t.KafkaAuthentication.Sasl.Mechanism.ValueString(),
						Username:  t.KafkaAuthentication.Sasl.Username.ValueString(),
						Password:  t.KafkaAuthentication.Sasl.Password.ValueString(),
					},
				},
				Enveloped: ptr(t.Enveloped.ValueBool()),
				Format:    t.Format.ValueString(),
			},
		}, diags
	case *AblyRuleTargetAMQP:
		return control.AMQPRulePost{
			Status:      status,
			RuleType:    "amqp",
			RequestMode: requestMode,
			Source:      source,
			Target: control.AMQPRuleTarget{
				QueueID:   t.QueueID.ValueString(),
				Headers:   GetHeaders(t.Headers),
				Enveloped: ptr(t.Enveloped.ValueBool()),
				Format:    t.Format.ValueString(),
			},
		}, diags
	case *AblyRuleTargetAMQPExternal:
		msgTTL := (*int)(nil)
		if !t.MessageTtl.IsNull() && !t.MessageTtl.IsUnknown() {
			v := int(t.MessageTtl.ValueInt64())
			msgTTL = &v
		}
		exchange := t.Exchange.ValueString()
		return control.AMQPExternalRulePost{
			Status:      status,
			RuleType:    "amqp/external",
			RequestMode: requestMode,
			Source:      source,
			Target: control.AMQPExternalRuleTarget{
				URL:                t.Url.ValueString(),
				RoutingKey:         t.RoutingKey.ValueString(),
				Exchange:           exchange,
				MandatoryRoute:     ptr(t.MandatoryRoute.ValueBool()),
				PersistentMessages: ptr(t.PersistentMessages.ValueBool()),
				MessageTTL:         msgTTL,
				Headers:            GetHeaders(t.Headers),
				Enveloped:          ptr(t.Enveloped.ValueBool()),
				Format:             t.Format.ValueString(),
			},
		}, diags
	}

	diags.AddError(
		"Unrecognized rule target type",
		fmt.Sprintf("The plan contains an unrecognized rule target type: %T", plan.Target),
	)
	return nil, diags
}

func GetHeaders(headers []AblyRuleHeaders) []control.RuleHeader {
	var retHeaders []control.RuleHeader
	for _, h := range headers {
		retHeaders = append(retHeaders, control.RuleHeader{
			Name:  h.Name.ValueString(),
			Value: h.Value.ValueString(),
		})
	}

	return retHeaders
}

func GetRequestMode(plan AblyRule) string {
	if plan.RequestMode.ValueString() == "batch" {
		return "batch"
	}
	return "single"
}

// GetAwsAuth converts AWS authentication from control SDK format to terraform format.
// Using plan to fill in values that the api does not return.
func GetAwsAuth(auth control.AWSAuthentication, plan *AblyRule) AwsAuth {
	var planAuth AwsAuth

	switch p := plan.Target.(type) {
	case *AblyRuleTargetKinesis:
		if p != nil {
			planAuth = p.AwsAuth
		}
	case *AblyRuleTargetSqs:
		if p != nil {
			planAuth = p.AwsAuth
		}
	case *AblyRuleTargetLambda:
		if p != nil {
			planAuth = p.AwsAuth
		}
	}

	var respAwsAuth AwsAuth
	switch control.AWSAuthMode(auth.AuthenticationMode) {
	case control.AWSAuthModeCredentials:
		respAwsAuth = AwsAuth{
			AuthenticationMode: types.StringValue(auth.AuthenticationMode),
			AccessKeyId:        types.StringValue(auth.AccessKeyID),
			SecretAccessKey:    planAuth.SecretAccessKey,
			RoleArn:            types.StringNull(),
		}
	case control.AWSAuthModeAssumeRole:
		respAwsAuth = AwsAuth{
			AuthenticationMode: types.StringValue(auth.AuthenticationMode),
			RoleArn:            types.StringValue(auth.AssumeRoleArn),
			AccessKeyId:        types.StringNull(),
			SecretAccessKey:    types.StringNull(),
		}
	}

	return respAwsAuth
}

// unmarshalTarget JSON-marshals the generic target from RuleResponse and unmarshals into a typed struct.
func unmarshalTarget[T any](target interface{}) (T, error) {
	var result T
	b, err := json.Marshal(target)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(b, &result)
	return result, err
}

// ToHeaders converts a slice of control.RuleHeader to terraform AblyRuleHeaders.
func ToHeaders(headers []control.RuleHeader) []AblyRuleHeaders {
	var respHeaders []AblyRuleHeaders
	for _, b := range headers {
		item := AblyRuleHeaders{
			Name:  types.StringValue(b.Name),
			Value: types.StringValue(b.Value),
		}
		respHeaders = append(respHeaders, item)
	}
	return respHeaders
}

// GetRuleResponse maps response body to resource schema attributes.
// Using plan to fill in values that the api does not return.
// Returns (AblyRule, diag.Diagnostics) so callers can check for unmarshal errors.
func GetRuleResponse(ablyRule *control.RuleResponse, plan *AblyRule) (AblyRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	var respTarget any

	switch ablyRule.RuleType {
	case "aws/kinesis":
		target, err := unmarshalTarget[control.AWSKinesisTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal aws/kinesis target: %s", err.Error()))
			return AblyRule{}, diags
		}
		respTarget = &AblyRuleTargetKinesis{
			Region:       types.StringValue(target.Region),
			StreamName:   types.StringValue(target.StreamName),
			PartitionKey: types.StringValue(target.PartitionKey),
			AwsAuth:      GetAwsAuth(target.Authentication, plan),
			Enveloped:    types.BoolValue(deref(target.Enveloped)),
			Format:       types.StringValue(target.Format),
		}
	case "aws/sqs":
		target, err := unmarshalTarget[control.AWSSQSTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal aws/sqs target: %s", err.Error()))
			return AblyRule{}, diags
		}
		respTarget = &AblyRuleTargetSqs{
			Region:       types.StringValue(target.Region),
			AwsAccountID: types.StringValue(target.AWSAccountID),
			QueueName:    types.StringValue(target.QueueName),
			AwsAuth:      GetAwsAuth(target.Authentication, plan),
			Enveloped:    types.BoolValue(deref(target.Enveloped)),
			Format:       types.StringValue(target.Format),
		}
	case "aws/lambda":
		target, err := unmarshalTarget[control.AWSLambdaTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal aws/lambda target: %s", err.Error()))
			return AblyRule{}, diags
		}
		respTarget = &AblyRuleTargetLambda{
			Region:       types.StringValue(target.Region),
			FunctionName: types.StringValue(target.FunctionName),
			AwsAuth:      GetAwsAuth(target.Authentication, plan),
			Enveloped:    types.BoolValue(deref(target.Enveloped)),
		}
	case "http/zapier":
		target, err := unmarshalTarget[control.ZapierRuleTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal http/zapier target: %s", err.Error()))
			return AblyRule{}, diags
		}
		headers := ToHeaders(target.Headers)
		respTarget = &AblyRuleTargetZapier{
			Url:          types.StringValue(target.URL),
			SigningKeyId: optStringValue(target.SigningKeyID),
			Headers:      headers,
		}
	case "http/cloudflare-worker":
		target, err := unmarshalTarget[control.CloudflareWorkerRuleTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal http/cloudflare-worker target: %s", err.Error()))
			return AblyRule{}, diags
		}
		headers := ToHeaders(target.Headers)
		respTarget = &AblyRuleTargetCloudflareWorker{
			Url:          types.StringValue(target.URL),
			SigningKeyId: optStringValue(target.SigningKeyID),
			Headers:      headers,
		}
	case "pulsar":
		target, err := unmarshalTarget[control.PulsarRuleTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal pulsar target: %s", err.Error()))
			return AblyRule{}, diags
		}
		// TlsTrustCerts is write-only in the API (accepted on create/update but
		// never returned on read), so preserve whatever the user configured in
		// state rather than overwriting it with nil from the API response.
		var tlsTrustCerts []types.String
		if p, ok := plan.Target.(*AblyRuleTargetPulsar); ok && p != nil {
			tlsTrustCerts = p.TlsTrustCerts
		}
		authMode := ""
		authToken := ""
		if target.Authentication != nil {
			authMode = target.Authentication.AuthenticationMode
			authToken = target.Authentication.Token
		}
		respTarget = &AblyRuleTargetPulsar{
			RoutingKey:    types.StringValue(target.RoutingKey),
			Topic:         types.StringValue(target.Topic),
			ServiceURL:    types.StringValue(target.ServiceURL),
			TlsTrustCerts: tlsTrustCerts,
			Authentication: PulsarAuthentication{
				Mode:  types.StringValue(authMode),
				Token: types.StringValue(authToken),
			},
			Enveloped: types.BoolValue(deref(target.Enveloped)),
			Format:    types.StringValue(target.Format),
		}
	case "http/ifttt":
		target, err := unmarshalTarget[control.IFTTTRuleTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal http/ifttt target: %s", err.Error()))
			return AblyRule{}, diags
		}
		respTarget = &AblyRuleTargetIFTTT{
			EventName:  types.StringValue(target.EventName),
			WebhookKey: types.StringValue(target.WebhookKey),
		}
	case "http/google-cloud-function":
		target, err := unmarshalTarget[control.GoogleCloudFunctionRuleTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal http/google-cloud-function target: %s", err.Error()))
			return AblyRule{}, diags
		}
		headers := ToHeaders(target.Headers)
		respTarget = &AblyRuleTargetGoogleFunction{
			Region:       types.StringValue(target.Region),
			ProjectID:    types.StringValue(target.ProjectID),
			FunctionName: types.StringValue(target.FunctionName),
			Headers:      headers,
			SigningKeyId: optStringValue(target.SigningKeyID),
			Enveloped:    types.BoolValue(deref(target.Enveloped)),
			Format:       types.StringValue(target.Format),
		}
	case "http/azure-function":
		target, err := unmarshalTarget[control.AzureFunctionRuleTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal http/azure-function target: %s", err.Error()))
			return AblyRule{}, diags
		}
		headers := ToHeaders(target.Headers)
		respTarget = &AblyRuleTargetAzureFunction{
			AzureAppID:        types.StringValue(target.AzureAppID),
			AzureFunctionName: types.StringValue(target.AzureFunctionName),
			Headers:           headers,
			SigningKeyID:      optStringValue(target.SigningKeyID),
			Enveloped:         types.BoolValue(deref(target.Enveloped)),
			Format:            types.StringValue(target.Format),
		}
	case "http":
		target, err := unmarshalTarget[control.HTTPRuleTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal http target: %s", err.Error()))
			return AblyRule{}, diags
		}
		headers := ToHeaders(target.Headers)
		respTarget = &AblyRuleTargetHTTP{
			Url:          types.StringValue(target.URL),
			Headers:      headers,
			SigningKeyId: optStringValue(target.SigningKeyID),
			Format:       types.StringValue(target.Format),
			Enveloped:    types.BoolValue(deref(target.Enveloped)),
		}
	case "kafka":
		target, err := unmarshalTarget[control.KafkaRuleTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal kafka target: %s", err.Error()))
			return AblyRule{}, diags
		}
		saslMechanism := ""
		saslUsername := ""
		saslPassword := ""
		if target.Auth != nil && target.Auth.SASL != nil {
			saslMechanism = target.Auth.SASL.Mechanism
			saslUsername = target.Auth.SASL.Username
			saslPassword = target.Auth.SASL.Password
		}
		respTarget = &AblyRuleTargetKafka{
			RoutingKey: types.StringValue(target.RoutingKey),
			Brokers:    toTypedStringSlice(target.Brokers),
			KafkaAuthentication: KafkaAuthentication{
				Sasl{
					Mechanism: types.StringValue(saslMechanism),
					Username:  types.StringValue(saslUsername),
					Password:  types.StringValue(saslPassword),
				},
			},
			Enveloped: types.BoolValue(deref(target.Enveloped)),
			Format:    types.StringValue(target.Format),
		}
	case "amqp":
		target, err := unmarshalTarget[control.AMQPRuleTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal amqp target: %s", err.Error()))
			return AblyRule{}, diags
		}
		headers := ToHeaders(target.Headers)
		respTarget = &AblyRuleTargetAMQP{
			QueueID:   types.StringValue(target.QueueID),
			Headers:   headers,
			Enveloped: types.BoolValue(deref(target.Enveloped)),
			Format:    types.StringValue(target.Format),
		}
	case "amqp/external":
		target, err := unmarshalTarget[control.AMQPExternalRuleTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal amqp/external target: %s", err.Error()))
			return AblyRule{}, diags
		}
		headers := ToHeaders(target.Headers)

		// Several target fields are not required in the API response and may
		// be omitted. When the plan provided values for these fields, preserve
		// them so Terraform doesn't see a diff (the target block contains the
		// sensitive "url" field, so ANY field mismatch triggers the opaque
		// "inconsistent values for sensitive attribute" error).
		url := types.StringValue(target.URL)
		exchange := types.StringNull()
		if target.Exchange != "" {
			exchange = types.StringValue(target.Exchange)
		}
		ttl := types.Int64Null()
		if target.MessageTTL != nil && *target.MessageTTL != 0 {
			ttl = types.Int64Value(int64(*target.MessageTTL))
		}
		if p, ok := plan.Target.(*AblyRuleTargetAMQPExternal); ok && p != nil {
			if !p.Url.IsNull() {
				url = p.Url
			}
			if target.Exchange == "" {
				exchange = p.Exchange
			}
			if ttl.IsNull() && !p.MessageTtl.IsNull() {
				ttl = p.MessageTtl
			}
		}
		respTarget = &AblyRuleTargetAMQPExternal{
			Url:                url,
			RoutingKey:         types.StringValue(target.RoutingKey),
			Exchange:           exchange,
			MandatoryRoute:     types.BoolValue(deref(target.MandatoryRoute)),
			PersistentMessages: types.BoolValue(deref(target.PersistentMessages)),
			MessageTtl:         ttl,
			Headers:            headers,
			Enveloped:          types.BoolValue(deref(target.Enveloped)),
			Format:             types.StringValue(target.Format),
		}
	default:
		diags.AddError(
			"Unknown rule type in response",
			fmt.Sprintf("Received unrecognized rule type from API: %q", ablyRule.RuleType),
		)
		return AblyRule{}, diags
	}

	channelFilter := types.StringNull()
	if ablyRule.Source != nil && ablyRule.Source.ChannelFilter != "" {
		channelFilter = types.StringValue(ablyRule.Source.ChannelFilter)
	}

	sourceType := ""
	if ablyRule.Source != nil {
		sourceType = ablyRule.Source.Type
	}

	respSource := AblyRuleSource{
		ChannelFilter: channelFilter,
		Type:          types.StringValue(sourceType),
	}

	respRule := AblyRule{
		ID:          types.StringValue(ablyRule.ID),
		AppID:       types.StringValue(ablyRule.AppID),
		Status:      types.StringValue(ablyRule.Status),
		Source:      &respSource,
		Target:      respTarget,
		RequestMode: types.StringValue(ablyRule.RequestMode),
	}

	return respRule, diags
}

// GetRuleSchema returns the schema for a rule resource.
func GetRuleSchema(target map[string]schema.Attribute, markdownDescription string) schema.Schema {
	return schema.Schema{
		MarkdownDescription: markdownDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The rule ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_id": schema.StringAttribute{
				Required:    true,
				Description: "The Ably application ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The status of the rule. Rules can be enabled or disabled.",
				Default:     stringdefault.StaticString("enabled"),
				Validators: []validator.String{
					stringvalidator.OneOf("enabled", "disabled"),
				},
			},
			"request_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "This is Single Request mode or Batch Request mode. Single Request mode sends each event separately to the endpoint specified by the rule",
				PlanModifiers: []planmodifier.String{
					DefaultStringAttribute(types.StringValue("single")),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("single", "batch"),
				},
			},
			"source": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The source for the rule",
				Attributes: map[string]schema.Attribute{
					"channel_filter": schema.StringAttribute{
						Optional: true,
					},
					"type": schema.StringAttribute{
						Required: true,
					},
				},
			},
			"target": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The target for the rule",
				Attributes:  target,
			},
		},
	}
}

func GetAwsAuthSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Required:    true,
		Description: "AWS authentication configuration",
		Attributes: map[string]schema.Attribute{
			"mode": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Authentication method. Use 'credentials' or 'assumeRole'",
				Validators: []validator.String{
					stringvalidator.OneOf("credentials", "assumeRole"),
				},
			},
			"role_arn": schema.StringAttribute{
				Optional:    true,
				Description: "If you are using the 'ARN of an assumable role' authentication method, this is your Assume Role ARN",
			},
			"access_key_id": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The AWS key ID for the AWS IAM user",
			},
			"secret_access_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The AWS secret key for the AWS IAM user",
			},
		},
	}
}

func GetHeaderSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		Optional:    true,
		Description: "If you have additional information to send, you'll need to include the relevant headers",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:    true,
					Description: "The name of the header",
				},
				"value": schema.StringAttribute{
					Required:    true,
					Description: "The value of the header",
				},
			},
		},
	}
}

func GetEnvelopedSchema() schema.Attribute {
	return schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Delivered messages are wrapped in an Ably envelope by default that contains metadata about the message and its payload. The form of the envelope depends on whether it is part of a Webhook/Function or a Queue/Firehose rule. For everything besides Webhooks, you can ensure you only get the raw payload by unchecking \"Enveloped\" when setting up the rule.",
		PlanModifiers: []planmodifier.Bool{
			DefaultBoolAttribute(types.BoolValue(false)),
		},
	}
}

func GetFormatSchema() schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "JSON provides a text-based encoding, whereas MsgPack provides a more efficient binary encoding",
		PlanModifiers: []planmodifier.String{
			DefaultStringAttribute(types.StringValue("json")),
		},
		Validators: []validator.String{
			stringvalidator.OneOf("json", "msgpack", "json/ably-compact"),
		},
	}
}

type Rule interface {
	Provider() *AblyProvider
	Name() string
}

// CreateRule creates a new rule resource.
func CreateRule[T any](r Rule, ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.Provider().ensureConfigured(&resp.Diagnostics) {
		return
	}

	// Gets plan values
	var p AblyRuleDecoder[*T]
	diags := req.Plan.Get(ctx, &p)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan := p.Rule()
	planValues, planDiags := GetPlanRule(plan)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Creates a new Ably Rule by invoking the CreateRule function from the Client Library
	rule, err := r.Provider().client.CreateRule(ctx, plan.AppID.ValueString(), planValues)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating resource %s", r.Name()),
			fmt.Sprintf("Could not create resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	responseValues, respDiags := GetRuleResponse(&rule, &plan)
	resp.Diagnostics.Append(respDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, responseValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ReadRule reads an existing rule resource.
func ReadRule[T any](r Rule, ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var s AblyRuleDecoder[*T]
	diags := req.State.Get(ctx, &s)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	state := s.Rule()

	// Gets the Ably App ID and Ably Rule ID value for the resource
	appID := s.AppID.ValueString()
	ruleID := s.ID.ValueString()

	// Get Rule data
	rule, err := r.Provider().client.GetRule(ctx, appID, ruleID)

	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading resource %s", r.Name()),
			fmt.Sprintf("Could not read resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	responseValues, respDiags := GetRuleResponse(&rule, &state)
	resp.Diagnostics.Append(respDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Sets state to app values.
	diags = resp.State.Set(ctx, &responseValues)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// UpdateRule updates an existing rule resource.
func UpdateRule[T any](r Rule, ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Gets plan values
	var p AblyRuleDecoder[*T]
	diags := req.Plan.Get(ctx, &p)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	plan := p.Rule()

	ruleValues, planDiags := GetPlanRule(plan)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID and Ably Rule ID value for the resource
	appID := plan.AppID.ValueString()
	ruleID := plan.ID.ValueString()

	// Update Ably Rule
	rule, err := r.Provider().client.UpdateRule(ctx, appID, ruleID, ruleValues)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating resource %s", r.Name()),
			fmt.Sprintf("Could not update resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	responseValues, respDiags := GetRuleResponse(&rule, &plan)
	resp.Diagnostics.Append(respDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Sets state to app values.
	diags = resp.State.Set(ctx, &responseValues)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// DeleteRule deletes a rule resource.
func DeleteRule[T any](r Rule, ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var s AblyRuleDecoder[*T]
	diags := req.State.Get(ctx, &s)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	state := s.Rule()

	// Gets the Ably App ID and Ably Rule ID value for the resource
	appID := state.AppID.ValueString()
	ruleID := state.ID.ValueString()

	err := r.Provider().client.DeleteRule(ctx, appID, ruleID)
	if err != nil {
		if is404(err) {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Resource %s does not exist", r.Name()),
				fmt.Sprintf("Resource %s does not exist, it may have already been deleted: %s", r.Name(), err.Error()),
			)
		} else {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error deleting resource %s", r.Name()),
				fmt.Sprintf("Could not delete resource %s, unexpected error: %s", r.Name(), err.Error()),
			)
			return
		}
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

// ImportResource handles importing a resource.
func ImportResource(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse, fields ...string) {
	// Save the import identifier in the id attribute
	// identifier should be in the format appID,key_id
	idParts := strings.Split(req.ID, ",")
	anyEmpty := false

	for _, v := range idParts {
		if v == "" {
			anyEmpty = true
		}
	}

	if len(idParts) != len(fields) || anyEmpty {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: '%s'. Got: %q", strings.Join(fields, ","), req.ID),
		)
		return
	}
	// Recent PR in TF Plugin Framework for paths but Hashicorp examples not updated - https://github.com/hashicorp/terraform-plugin-framework/pull/390
	for i, v := range fields {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(v), idParts[i])...)
	}
}
