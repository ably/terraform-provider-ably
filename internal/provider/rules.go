package provider

import (
	"context"
	"fmt"
	"strings"

	control "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func GetPlanAwsAuth(plan AblyRule) control.AwsAuthentication {
	var auth AwsAuth
	var controlAuth control.AwsAuthentication

	switch t := plan.Target.(type) {
	case *AblyRuleTargetKinesis:
		auth = t.AwsAuth
	case *AblyRuleTargetSqs:
		auth = t.AwsAuth
	case *AblyRuleTargetLambda:
		auth = t.AwsAuth
	}

	if auth.AuthenticationMode.ValueString() == "assumeRole" {
		controlAuth = control.AwsAuthentication{
			Authentication: &control.AuthenticationModeAssumeRole{
				AssumeRoleArn: auth.RoleArn.ValueString(),
			},
		}
	} else if auth.AuthenticationMode.ValueString() == "credentials" {
		controlAuth = control.AwsAuthentication{
			Authentication: &control.AuthenticationModeCredentials{
				AccessKeyId:     auth.AccessKeyId.ValueString(),
				SecretAccessKey: auth.SecretAccessKey.ValueString(),
			},
		}
	}

	return controlAuth
}

// converts rule from terraform format to control sdk format
func GetPlanRule(plan AblyRule) control.NewRule {
	var target control.Target

	switch t := plan.Target.(type) {
	case *AblyRuleTargetKinesis:
		target = &control.AwsKinesisTarget{
			Region:         t.Region,
			StreamName:     t.StreamName,
			PartitionKey:   t.PartitionKey,
			Authentication: GetPlanAwsAuth(plan),
			Enveloped:      t.Enveloped,
			Format:         t.Format,
		}
	case *AblyRuleTargetSqs:
		target = &control.AwsSqsTarget{
			Region:         t.Region,
			AwsAccountID:   t.AwsAccountID,
			QueueName:      t.QueueName,
			Authentication: GetPlanAwsAuth(plan),
			Enveloped:      t.Enveloped,
			Format:         t.Format,
		}
	case *AblyRuleTargetLambda:
		target = &control.AwsLambdaTarget{
			Region:         t.Region,
			FunctionName:   t.FunctionName,
			Authentication: GetPlanAwsAuth(plan),
			Enveloped:      t.Enveloped,
		}
	case *AblyRuleTargetZapier:
		target = &control.HttpZapierTarget{
			Url:          t.Url,
			Headers:      GetHeaders(t.Headers),
			SigningKeyID: t.SigningKeyId,
		}
	case *AblyRuleTargetCloudflareWorker:
		target = &control.HttpCloudfareWorkerTarget{
			Url:          t.Url,
			Headers:      GetHeaders(t.Headers),
			SigningKeyID: t.SigningKeyId,
		}
	case *AblyRuleTargetPulsar:
		target = &control.PulsarTarget{
			RoutingKey:    t.RoutingKey,
			Topic:         t.Topic,
			ServiceURL:    t.ServiceURL,
			TlsTrustCerts: t.TlsTrustCerts,
			Authentication: control.PulsarAuthentication{
				AuthenticationMode: control.PularAuthenticationMode(t.Authentication.Mode),
				Token:              t.Authentication.Token,
			},
			Enveloped: t.Enveloped,
			Format:    t.Format,
		}
	case *AblyRuleTargetHTTP:
		var headers []control.Header
		for _, h := range t.Headers {
			headers = append(headers, control.Header{
				Name:  h.Name.ValueString(),
				Value: h.Value.ValueString(),
			})
		}

		target = &control.HttpTarget{
			Url:          t.Url,
			Headers:      headers,
			SigningKeyID: t.SigningKeyId,
			Format:       t.Format,
			Enveloped:    t.Enveloped,
		}
	case *AblyRuleTargetIFTTT:
		target = &control.HttpIftttTarget{
			WebhookKey: t.WebhookKey,
			EventName:  t.EventName,
		}
	case *AblyRuleTargetAzureFunction:
		target = &control.HttpAzureFunctionTarget{
			AzureAppID:        t.AzureAppID,
			AzureFunctionName: t.AzureFunctionName,
			Headers:           GetHeaders(t.Headers),
			SigningKeyID:      t.SigningKeyID,
			Format:            t.Format,
		}
	case *AblyRuleTargetGoogleFunction:
		target = &control.HttpGoogleCloudFunctionTarget{
			Region:       t.Region,
			ProjectID:    t.ProjectID,
			FunctionName: t.FunctionName,
			Headers:      GetHeaders(t.Headers),
			SigningKeyID: t.SigningKeyId,
			Enveloped:    t.Enveloped,
			Format:       t.Format,
		}

	case *AblyRuleTargetKafka:
		target = &control.KafkaTarget{
			RoutingKey: t.RoutingKey,
			Brokers:    t.Brokers,
			Authentication: control.KafkaAuthentication{
				Sasl: control.Sasl{
					Mechanism: control.SaslMechanism(t.KafkaAuthentication.Sasl.Mechanism),
					Username:  t.KafkaAuthentication.Sasl.Username,
					Password:  t.KafkaAuthentication.Sasl.Password,
				},
			},
			Enveloped: t.Enveloped,
			Format:    t.Format,
		}
	case *AblyRuleTargetAmqp:
		target = &control.AmqpTarget{
			QueueID:   t.QueueID,
			Headers:   GetHeaders(t.Headers),
			Enveloped: t.Enveloped,
			Format:    t.Format,
		}
	case *AblyRuleTargetAmqpExternal:
		target = &control.AmqpExternalTarget{
			Url:                t.Url,
			RoutingKey:         t.RoutingKey,
			Exchange:           t.Exchange,
			MandatoryRoute:     t.MandatoryRoute,
			PersistentMessages: t.PersistentMessages,
			MessageTTL:         int(t.MessageTtl.ValueInt64()),
			Headers:            GetHeaders(t.Headers),
			Enveloped:          t.Enveloped,
			Format:             t.Format,
		}
	}

	ruleValues := control.NewRule{
		Status:      plan.Status.ValueString(),
		RequestMode: GetRequestMode(plan),
		Source: control.Source{
			ChannelFilter: plan.Source.ChannelFilter.ValueString(),
			Type:          GetSourceType(plan.Source.Type),
		},
		Target: target,
	}

	return ruleValues
}

func GetHeaders(headers []AblyRuleHeaders) []control.Header {
	var retHeaders []control.Header
	for _, h := range headers {
		retHeaders = append(retHeaders, control.Header{
			Name:  h.Name.ValueString(),
			Value: h.Value.ValueString(),
		})
	}

	return retHeaders
}

func GetRequestMode(plan AblyRule) control.RequestMode {
	switch plan.RequestMode.ValueString() {
	case "single":
		return control.Single
	case "batch":
		return control.Batch
	default:
		return control.Single
	}
}

// Maps response body to resource schema attributes.
// Using plan to fill in values that the api does not return.
func GetAwsAuth(auth *control.AwsAuthentication, plan *AblyRule) AwsAuth {
	var respAwsAuth AwsAuth
	var planAuth AwsAuth

	switch p := plan.Target.(type) {
	case *AblyRuleTargetKinesis:
		planAuth = p.AwsAuth
	case *AblyRuleTargetSqs:
		planAuth = p.AwsAuth
	case *AblyRuleTargetLambda:
		planAuth = p.AwsAuth
	}

	switch a := auth.Authentication.(type) {
	case *control.AuthenticationModeCredentials:
		respAwsAuth = AwsAuth{
			AuthenticationMode: planAuth.AuthenticationMode,
			AccessKeyId:        types.StringValue(a.AccessKeyId),
			SecretAccessKey:    planAuth.SecretAccessKey,
			RoleArn:            types.StringNull(),
		}
	case *control.AuthenticationModeAssumeRole:
		respAwsAuth = AwsAuth{
			AuthenticationMode: planAuth.AuthenticationMode,
			RoleArn:            types.StringValue(a.AssumeRoleArn),
			AccessKeyId:        types.StringNull(),
			SecretAccessKey:    types.StringNull(),
		}
	}

	return respAwsAuth
}

// Maps response body to resource schema attributes.
// Using plan to fill in values that the api does not return.
func GetRuleResponse(rule *control.Rule, plan *AblyRule) AblyRule {
	var respTarget interface{}

	switch v := rule.Target.(type) {
	case *control.AwsKinesisTarget:
		respTarget = &AblyRuleTargetKinesis{
			Region:       v.Region,
			StreamName:   v.StreamName,
			PartitionKey: v.PartitionKey,
			AwsAuth:      GetAwsAuth(&v.Authentication, plan),
			Enveloped:    v.Enveloped,
			Format:       v.Format,
		}
	case *control.AwsSqsTarget:
		respTarget = &AblyRuleTargetSqs{
			Region:       v.Region,
			AwsAccountID: v.AwsAccountID,
			QueueName:    v.QueueName,
			AwsAuth:      GetAwsAuth(&v.Authentication, plan),
			Enveloped:    v.Enveloped,
			Format:       v.Format,
		}
	case *control.AwsLambdaTarget:
		respTarget = &AblyRuleTargetLambda{
			Region:       v.Region,
			FunctionName: v.FunctionName,
			AwsAuth:      GetAwsAuth(&v.Authentication, plan),
			Enveloped:    v.Enveloped,
		}
	case *control.HttpZapierTarget:
		headers := ToHeaders(v)

		respTarget = &AblyRuleTargetZapier{
			Url:          v.Url,
			SigningKeyId: v.SigningKeyID,
			Headers:      headers,
		}
	case *control.HttpCloudfareWorkerTarget:
		headers := ToHeaders(v)

		respTarget = &AblyRuleTargetCloudflareWorker{
			Url:          v.Url,
			SigningKeyId: v.SigningKeyID,
			Headers:      headers,
		}
	case *control.PulsarTarget:
		respTarget = &AblyRuleTargetPulsar{
			RoutingKey:    v.RoutingKey,
			Topic:         v.Topic,
			ServiceURL:    v.ServiceURL,
			TlsTrustCerts: v.TlsTrustCerts,
			Authentication: PulsarAuthentication{
				Mode:  string(v.Authentication.AuthenticationMode),
				Token: v.Authentication.Token,
			},
			Enveloped: v.Enveloped,
			Format:    v.Format,
		}
	case *control.HttpIftttTarget:
		respTarget = &AblyRuleTargetIFTTT{
			EventName:  v.EventName,
			WebhookKey: v.WebhookKey,
		}
	case *control.HttpGoogleCloudFunctionTarget:
		headers := ToHeaders(v)

		respTarget = &AblyRuleTargetGoogleFunction{
			Region:       v.Region,
			ProjectID:    v.ProjectID,
			FunctionName: v.FunctionName,
			Headers:      headers,
			SigningKeyId: v.SigningKeyID,
			Enveloped:    v.Enveloped,
			Format:       v.Format,
		}
	case *control.HttpAzureFunctionTarget:
		headers := ToHeaders(v)

		respTarget = &AblyRuleTargetAzureFunction{
			AzureAppID:        v.AzureAppID,
			AzureFunctionName: v.AzureFunctionName,
			Headers:           headers,
			SigningKeyID:      v.SigningKeyID,
			Format:            v.Format,
		}
	case *control.HttpTarget:
		headers := ToHeaders(v)

		respTarget = &AblyRuleTargetHTTP{
			Url:          v.Url,
			Headers:      headers,
			SigningKeyId: v.SigningKeyID,
			Format:       v.Format,
			Enveloped:    v.Enveloped,
		}
	case *control.KafkaTarget:
		respTarget = &AblyRuleTargetKafka{
			RoutingKey: v.RoutingKey,
			Brokers:    v.Brokers,
			KafkaAuthentication: KafkaAuthentication{
				Sasl{
					Mechanism: string(v.Authentication.Sasl.Mechanism),
					Username:  v.Authentication.Sasl.Username,
					Password:  v.Authentication.Sasl.Password,
				},
			},
			Enveloped: v.Enveloped,
			Format:    v.Format,
		}
	case *control.AmqpTarget:
		headers := ToHeaders(v)

		respTarget = &AblyRuleTargetAmqp{
			QueueID:   v.QueueID,
			Headers:   headers,
			Enveloped: v.Enveloped,
			Format:    v.Format,
		}
	case *control.AmqpExternalTarget:
		headers := ToHeaders(v)
		ttl := types.Int64Null()
		if v.MessageTTL != 0 {
			ttl = types.Int64Value(int64(v.MessageTTL))
		}

		respTarget = &AblyRuleTargetAmqpExternal{
			Url:                v.Url,
			RoutingKey:         v.RoutingKey,
			Exchange:           v.Exchange,
			MandatoryRoute:     v.MandatoryRoute,
			PersistentMessages: v.PersistentMessages,
			MessageTtl:         ttl,
			Headers:            headers,
			Enveloped:          v.Enveloped,
			Format:             v.Format,
		}
	}

	channelFilter := types.StringNull()
	if rule.Source.ChannelFilter != "" {
		channelFilter = types.StringValue(
			rule.Source.ChannelFilter,
		)
	}

	respSource := AblyRuleSource{
		ChannelFilter: channelFilter,
		Type:          rule.Source.Type,
	}

	respRule := AblyRule{
		ID:          types.StringValue(rule.ID),
		AppID:       types.StringValue(rule.AppID),
		Status:      types.StringValue(rule.Status),
		Source:      &respSource,
		Target:      respTarget,
		RequestMode: types.StringValue(string(rule.RequestMode)),
	}

	return respRule
}

func GetRuleSchema(target map[string]tfsdk.Attribute, markdownDescription string) tfsdk.Schema {
	return tfsdk.Schema{
		MarkdownDescription: markdownDescription,
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The rule ID.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk_resource.UseStateForUnknown(),
				},
			},
			"app_id": {
				Type:        types.StringType,
				Required:    true,
				Description: "The Ably application ID.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk_resource.RequiresReplace(),
				},
			},
			"status": {
				Type:        types.StringType,
				Optional:    true,
				Description: "The status of the rule. Rules can be enabled or disabled.",
			},
			"request_mode": {
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "This is Single Request mode or Batch Request mode. Single Request mode sends each event separately to the endpoint specified by the rule",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.StringValue("single")),
					tfsdk_resource.UseStateForUnknown(),
				},
			},
			"source": {
				Required:    true,
				Description: "object (rule_source)",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"channel_filter": {
						Type:     types.StringType,
						Optional: true,
					},
					"type": {
						Type:     types.StringType,
						Required: true,
					},
				}),
			},
			"target": {
				Required:    true,
				Description: "object (rule_source)",
				Attributes:  tfsdk.SingleNestedAttributes(target),
			},
		},
	}
}

func GetAwsAuthSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Required:    true,
		Description: "object (rule_source)",
		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			"mode": {
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk_resource.RequiresReplace(),
				},
				Description: "Authentication method. Use 'credentials' or 'assumeRole'",
			},
			"role_arn": {
				Type:        types.StringType,
				Optional:    true,
				Description: "If you are using the 'ARN of an assumable role' authentication method, this is your Assume Role ARN",
			},
			"access_key_id": {
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
				Description: "The AWS key ID for the AWS IAM user",
			},
			"secret_access_key": {
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
				Description: "The AWS secret key for the AWS IAM user",
			},
		}),
	}
}

func GetHeaderSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Optional:    true,
		Description: "If you have additional information to send, you'll need to include the relevant headers",
		Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
			"name": {
				Type:        types.StringType,
				Required:    true,
				Description: "The name of the header",
			},
			"value": {
				Type:        types.StringType,
				Required:    true,
				Description: "The value of the header",
			},
		}),
	}
}

func GetEnvelopedSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Type:        types.BoolType,
		Optional:    true,
		Computed:    true,
		Description: "Delivered messages are wrapped in an Ably envelope by default that contains metadata about the message and its payload. The form of the envelope depends on whether it is part of a Webhook/Function or a Queue/Firehose rule. For everything besides Webhooks, you can ensure you only get the raw payload by unchecking \"Enveloped\" when setting up the rule.",
		PlanModifiers: []tfsdk.AttributePlanModifier{
			DefaultAttribute(types.BoolValue(false)),
		},
	}
}

func GetFormatSchema() tfsdk.Attribute {
	return tfsdk.Attribute{
		Type:        types.StringType,
		Optional:    true,
		Computed:    true,
		Description: "JSON provides a text-based encoding, whereas MsgPack provides a more efficient binary encoding",
		PlanModifiers: []tfsdk.AttributePlanModifier{
			DefaultAttribute(types.StringValue("json")),
		},
	}
}

func GetSourceType(mode control.SourceType) control.SourceType {
	switch mode {
	case "channel.message":
		return control.ChannelMessage
	case "channel.presence":
		return control.ChannelPresence
	case "channel.lifecycle":
		return control.ChannelLifeCycle
	case "channel.occupancy":
		return control.ChannelOccupancy
	default:
		return control.ChannelMessage
	}
}

func ToHeaders(plan control.Target) []AblyRuleHeaders {
	var respHeaders []AblyRuleHeaders
	var headers []control.Header

	switch t := plan.(type) {
	case *control.HttpTarget:
		headers = t.Headers
	case *control.HttpZapierTarget:
		headers = t.Headers
	case *control.HttpCloudfareWorkerTarget:
		headers = t.Headers
	case *control.HttpGoogleCloudFunctionTarget:
		headers = t.Headers
	case *control.HttpAzureFunctionTarget:
		headers = t.Headers
	case *control.AmqpTarget:
		headers = t.Headers
	case *control.AmqpExternalTarget:
		headers = t.Headers
	}

	for _, b := range headers {
		item := AblyRuleHeaders{
			Name:  types.StringValue(b.Name),
			Value: types.StringValue(b.Value),
		}
		respHeaders = append(respHeaders, item)
	}

	return respHeaders
}

func GetKafkaAuthSchema(headers []AblyRuleHeaders) []control.Header {
	var retHeaders []control.Header
	for _, h := range headers {
		retHeaders = append(retHeaders, control.Header{
			Name:  h.Name.ValueString(),
			Value: h.Value.ValueString(),
		})
	}

	return retHeaders
}

type Rule interface {
	Provider() *provider
	Name() string
}

// Create a new resource
func CreateRule[T any](r Rule, ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	// Checks whether the provider and API Client are configured. If they are not, the provider responds with an error.
	if !r.Provider().configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply",
		)
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
	planValues := GetPlanRule(plan)

	// Creates a new Ably Rule by invoking the CreateRule function from the Client Library
	rule, err := r.Provider().client.CreateRule(plan.AppID.ValueString(), &planValues)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating Resource '%s'", r.Name()),
			fmt.Sprintf("Could not create resource '%s', unexpected error: %s", r.Name(), err.Error()),
		)

		return
	}

	responseValues := GetRuleResponse(&rule, &plan)

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, responseValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource
func ReadRule[T any](r Rule, ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
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
	rule, err := r.Provider().client.Rule(appID, ruleID)

	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting Resource %s", r.Name()),
			fmt.Sprintf("Could not delete resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	responseValues := GetRuleResponse(&rule, &state)

	// Sets state to app values.
	diags = resp.State.Set(ctx, &responseValues)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// // Update resource
func UpdateRule[T any](r Rule, ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	// Gets plan values
	var p AblyRuleDecoder[*T]
	diags := req.Plan.Get(ctx, &p)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	plan := p.Rule()

	ruleValues := GetPlanRule(plan)

	// Gets the Ably App ID and Ably Rule ID value for the resource
	appID := plan.AppID.ValueString()
	ruleID := plan.ID.ValueString()

	// Update Ably Rule
	rule, err := r.Provider().client.UpdateRule(appID, ruleID, &ruleValues)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updading Resource %s", r.Name()),
			fmt.Sprintf("Could not update resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	responseValues := GetRuleResponse(&rule, &plan)

	// Sets state to app values.
	diags = resp.State.Set(ctx, &responseValues)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func DeleteRule[T any](r Rule, ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
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

	err := r.Provider().client.DeleteRule(appID, ruleID)
	if err != nil {
		if is404(err) {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Resource does %s not exist", r.Name()),
				fmt.Sprintf("Resource does %s not exist, it may have already been deleted: %s", r.Name(), err.Error()),
			)
		} else {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error deleting Resource %s'", r.Name()),
				fmt.Sprintf("Could not delete resource '%s', unexpected error: %s", r.Name(), err.Error()),
			)
			return
		}
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

// // Import resource
func ImportResource(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse, fields ...string) {
	// Save the import identifier in the id attribute
	// identifier should be in the format app_id,key_id
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
