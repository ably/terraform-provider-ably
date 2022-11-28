package ably_control

import (
	"context"
	"fmt"
	"strings"

	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func GetPlanAwsAuth(plan AblyRule) ably_control_go.AwsAuthentication {
	var auth AwsAuth
	var control_auth ably_control_go.AwsAuthentication

	switch t := plan.Target.(type) {
	case *AblyRuleTargetKinesis:
		auth = t.AwsAuth
	case *AblyRuleTargetSqs:
		auth = t.AwsAuth
	case *AblyRuleTargetLambda:
		auth = t.AwsAuth
	}

	if auth.AuthenticationMode.ValueString() == "assumeRole" {
		control_auth = ably_control_go.AwsAuthentication{
			Authentication: &ably_control_go.AuthenticationModeAssumeRole{
				AssumeRoleArn: auth.RoleArn.ValueString(),
			},
		}
	} else if auth.AuthenticationMode.ValueString() == "credentials" {
		control_auth = ably_control_go.AwsAuthentication{
			Authentication: &ably_control_go.AuthenticationModeCredentials{
				AccessKeyId:     auth.AccessKeyId.ValueString(),
				SecretAccessKey: auth.SecretAccessKey.ValueString(),
			},
		}
	}

	return control_auth
}

// converts rule from terraform format to control sdk format
func GetPlanRule(plan AblyRule) ably_control_go.NewRule {
	var target ably_control_go.Target

	switch t := plan.Target.(type) {
	case *AblyRuleTargetKinesis:
		target = &ably_control_go.AwsKinesisTarget{
			Region:         t.Region,
			StreamName:     t.StreamName,
			PartitionKey:   t.PartitionKey,
			Authentication: GetPlanAwsAuth(plan),
			Enveloped:      t.Enveloped,
			Format:         t.Format,
		}
	case *AblyRuleTargetSqs:
		target = &ably_control_go.AwsSqsTarget{
			Region:         t.Region,
			AwsAccountID:   t.AwsAccountID,
			QueueName:      t.QueueName,
			Authentication: GetPlanAwsAuth(plan),
			Enveloped:      t.Enveloped,
			Format:         t.Format,
		}
	case *AblyRuleTargetLambda:
		target = &ably_control_go.AwsLambdaTarget{
			Region:         t.Region,
			FunctionName:   t.FunctionName,
			Authentication: GetPlanAwsAuth(plan),
			Enveloped:      t.Enveloped,
		}
	case *AblyRuleTargetZapier:
		target = &ably_control_go.HttpZapierTarget{
			Url:          t.Url,
			Headers:      GetHeaders(t.Headers),
			SigningKeyID: t.SigningKeyId,
		}
	case *AblyRuleTargetCloudflareWorker:
		target = &ably_control_go.HttpCloudfareWorkerTarget{
			Url:          t.Url,
			Headers:      GetHeaders(t.Headers),
			SigningKeyID: t.SigningKeyId,
		}
	case *AblyRuleTargetPulsar:
		target = &ably_control_go.PulsarTarget{
			RoutingKey:    t.RoutingKey,
			Topic:         t.Topic,
			ServiceURL:    t.ServiceURL,
			TlsTrustCerts: t.TlsTrustCerts,
			Authentication: ably_control_go.PulsarAuthentication{
				AuthenticationMode: ably_control_go.PularAuthenticationMode(t.Authentication.Mode),
				Token:              t.Authentication.Token,
			},
			Enveloped: t.Enveloped,
			Format:    t.Format,
		}
	case *AblyRuleTargetHTTP:
		var headers []ably_control_go.Header
		for _, h := range t.Headers {
			headers = append(headers, ably_control_go.Header{
				Name:  h.Name.ValueString(),
				Value: h.Value.ValueString(),
			})
		}

		target = &ably_control_go.HttpTarget{
			Url:          t.Url,
			Headers:      headers,
			SigningKeyID: t.SigningKeyId,
			Format:       t.Format,
		}
	case *AblyRuleTargetIFTTT:
		target = &ably_control_go.HttpIftttTarget{
			WebhookKey: t.WebhookKey,
			EventName:  t.EventName,
		}
	case *AblyRuleTargetAzureFunction:
		target = &ably_control_go.HttpAzureFunctionTarget{
			AzureAppID:        t.AzureAppID,
			AzureFunctionName: t.AzureFunctionName,
			Headers:           GetHeaders(t.Headers),
			SigningKeyID:      t.SigningKeyID,
			Format:            t.Format,
		}
	case *AblyRuleTargetGoogleFunction:
		target = &ably_control_go.HttpGoogleCloudFunctionTarget{
			Region:       t.Region,
			ProjectID:    t.ProjectID,
			FunctionName: t.FunctionName,
			Headers:      GetHeaders(t.Headers),
			SigningKeyID: t.SigningKeyId,
			Enveloped:    t.Enveloped,
			Format:       t.Format,
		}

	case *AblyRuleTargetKafka:
		target = &ably_control_go.KafkaTarget{
			RoutingKey: t.RoutingKey,
			Brokers:    t.Brokers,
			Authentication: ably_control_go.KafkaAuthentication{
				Sasl: ably_control_go.Sasl{
					Mechanism: ably_control_go.SaslMechanism(t.KafkaAuthentication.Sasl.Mechanism),
					Username:  t.KafkaAuthentication.Sasl.Username,
					Password:  t.KafkaAuthentication.Sasl.Password,
				},
			},
			Enveloped: t.Enveloped,
			Format:    t.Format,
		}
	case *AblyRuleTargetAmqp:
		target = &ably_control_go.AmqpTarget{
			QueueID:   t.QueueID,
			Headers:   GetHeaders(t.Headers),
			Enveloped: t.Enveloped,
			Format:    t.Format,
		}
	case *AblyRuleTargetAmqpExternal:
		target = &ably_control_go.AmqpExternalTarget{
			Url:                t.Url,
			RoutingKey:         t.RoutingKey,
			MandatoryRoute:     t.MandatoryRoute,
			PersistentMessages: t.PersistentMessages,
			MessageTTL:         int(t.MessageTtl.ValueInt64()),
			Headers:            GetHeaders(t.Headers),
			Enveloped:          t.Enveloped,
			Format:             t.Format,
		}
	}

	rule_values := ably_control_go.NewRule{
		Status:      plan.Status.ValueString(),
		RequestMode: GetRequestMode(plan),
		Source: ably_control_go.Source{
			ChannelFilter: plan.Source.ChannelFilter.ValueString(),
			Type:          GetSourceType(plan.Source.Type),
		},
		Target: target,
	}

	return rule_values
}

func GetHeaders(headers []AblyRuleHeaders) []ably_control_go.Header {
	var ret_headers []ably_control_go.Header
	for _, h := range headers {
		ret_headers = append(ret_headers, ably_control_go.Header{
			Name:  h.Name.ValueString(),
			Value: h.Value.ValueString(),
		})
	}

	return ret_headers
}

func GetRequestMode(plan AblyRule) ably_control_go.RequestMode {
	switch plan.RequestMode.ValueString() {
	case "single":
		return ably_control_go.Single
	case "batch":
		return ably_control_go.Batch
	default:
		return ably_control_go.Single
	}
}

// Maps response body to resource schema attributes.
// Using plan to fill in values that the api does not return.
func GetAwsAuth(auth *ably_control_go.AwsAuthentication, plan *AblyRule) AwsAuth {
	var resp_aws_auth AwsAuth
	var plan_auth AwsAuth

	switch p := plan.Target.(type) {
	case *AblyRuleTargetKinesis:
		plan_auth = p.AwsAuth
	case *AblyRuleTargetSqs:
		plan_auth = p.AwsAuth
	case *AblyRuleTargetLambda:
		plan_auth = p.AwsAuth
	}

	switch a := auth.Authentication.(type) {
	case *ably_control_go.AuthenticationModeCredentials:
		resp_aws_auth = AwsAuth{
			AuthenticationMode: plan_auth.AuthenticationMode,
			AccessKeyId:        types.StringValue(a.AccessKeyId),
			SecretAccessKey:    plan_auth.SecretAccessKey,
			RoleArn:            types.StringNull(),
		}
	case *ably_control_go.AuthenticationModeAssumeRole:
		resp_aws_auth = AwsAuth{
			AuthenticationMode: plan_auth.AuthenticationMode,
			RoleArn:            types.StringValue(a.AssumeRoleArn),
			AccessKeyId:        types.StringNull(),
			SecretAccessKey:    types.StringNull(),
		}
	}

	return resp_aws_auth
}

// Maps response body to resource schema attributes.
// Using plan to fill in values that the api does not return.
func GetRuleResponse(ably_rule *ably_control_go.Rule, plan *AblyRule) AblyRule {
	var resp_target interface{}

	switch v := ably_rule.Target.(type) {
	case *ably_control_go.AwsKinesisTarget:
		resp_target = &AblyRuleTargetKinesis{
			Region:       v.Region,
			StreamName:   v.StreamName,
			PartitionKey: v.PartitionKey,
			AwsAuth:      GetAwsAuth(&v.Authentication, plan),
			Enveloped:    v.Enveloped,
			Format:       v.Format,
		}
	case *ably_control_go.AwsSqsTarget:
		resp_target = &AblyRuleTargetSqs{
			Region:       v.Region,
			AwsAccountID: v.AwsAccountID,
			QueueName:    v.QueueName,
			AwsAuth:      GetAwsAuth(&v.Authentication, plan),
			Enveloped:    v.Enveloped,
			Format:       v.Format,
		}
	case *ably_control_go.AwsLambdaTarget:
		resp_target = &AblyRuleTargetLambda{
			Region:       v.Region,
			FunctionName: v.FunctionName,
			AwsAuth:      GetAwsAuth(&v.Authentication, plan),
			Enveloped:    v.Enveloped,
		}
	case *ably_control_go.HttpZapierTarget:
		headers := ToHeaders(v)

		resp_target = &AblyRuleTargetZapier{
			Url:          v.Url,
			SigningKeyId: v.SigningKeyID,
			Headers:      headers,
		}
	case *ably_control_go.HttpCloudfareWorkerTarget:
		headers := ToHeaders(v)

		resp_target = &AblyRuleTargetCloudflareWorker{
			Url:          v.Url,
			SigningKeyId: v.SigningKeyID,
			Headers:      headers,
		}
	case *ably_control_go.PulsarTarget:
		resp_target = &AblyRuleTargetPulsar{
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
	case *ably_control_go.HttpIftttTarget:
		resp_target = &AblyRuleTargetIFTTT{
			EventName:  v.EventName,
			WebhookKey: v.WebhookKey,
		}
	case *ably_control_go.HttpGoogleCloudFunctionTarget:
		headers := ToHeaders(v)

		resp_target = &AblyRuleTargetGoogleFunction{
			Region:       v.Region,
			ProjectID:    v.ProjectID,
			FunctionName: v.FunctionName,
			Headers:      headers,
			SigningKeyId: v.SigningKeyID,
			Enveloped:    v.Enveloped,
			Format:       v.Format,
		}
	case *ably_control_go.HttpAzureFunctionTarget:
		headers := ToHeaders(v)

		resp_target = &AblyRuleTargetAzureFunction{
			AzureAppID:        v.AzureAppID,
			AzureFunctionName: v.AzureFunctionName,
			Headers:           headers,
			SigningKeyID:      v.SigningKeyID,
			Format:            v.Format,
		}
	case *ably_control_go.HttpTarget:
		headers := ToHeaders(v)

		resp_target = &AblyRuleTargetHTTP{
			Url:          v.Url,
			Headers:      headers,
			SigningKeyId: v.SigningKeyID,
			Format:       v.Format,
		}
	case *ably_control_go.KafkaTarget:
		resp_target = &AblyRuleTargetKafka{
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
	case *ably_control_go.AmqpTarget:
		headers := ToHeaders(v)

		resp_target = &AblyRuleTargetAmqp{
			QueueID:   v.QueueID,
			Headers:   headers,
			Enveloped: v.Enveloped,
			Format:    v.Format,
		}
	case *ably_control_go.AmqpExternalTarget:
		headers := ToHeaders(v)
		ttl := types.Int64Null()
		if v.MessageTTL != 0 {
			ttl = types.Int64Value(int64(v.MessageTTL))
		}

		resp_target = &AblyRuleTargetAmqpExternal{
			Url:                v.Url,
			RoutingKey:         v.RoutingKey,
			MandatoryRoute:     v.MandatoryRoute,
			PersistentMessages: v.PersistentMessages,
			MessageTtl:         ttl,
			Headers:            headers,
			Enveloped:          v.Enveloped,
			Format:             v.Format,
		}
	}

	channel_filter := types.StringNull()
	if ably_rule.Source.ChannelFilter != "" {
		channel_filter = types.StringValue(
			ably_rule.Source.ChannelFilter,
		)
	}

	resp_source := AblyRuleSource{
		ChannelFilter: channel_filter,
		Type:          ably_rule.Source.Type,
	}

	resp_rule := AblyRule{
		ID:          types.StringValue(ably_rule.ID),
		AppID:       types.StringValue(ably_rule.AppID),
		Status:      types.StringValue(ably_rule.Status),
		Source:      resp_source,
		Target:      resp_target,
		RequestMode: types.StringValue(string(ably_rule.RequestMode)),
	}

	return resp_rule
}

func GetRuleSchema(target map[string]tfsdk.Attribute, markdown_description string) tfsdk.Schema {
	return tfsdk.Schema{
		MarkdownDescription: markdown_description,
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

func GetEnvelopedchema() tfsdk.Attribute {
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

func GetSourceType(mode ably_control_go.SourceType) ably_control_go.SourceType {
	switch mode {
	case "channel.message":
		return ably_control_go.ChannelMessage
	case "channel.presence":
		return ably_control_go.ChannelPresence
	case "channel.lifecycle":
		return ably_control_go.ChannelLifeCycle
	case "channel.occupancy":
		return ably_control_go.ChannelOccupancy
	default:
		return ably_control_go.ChannelMessage
	}
}

func ToHeaders(plan ably_control_go.Target) []AblyRuleHeaders {
	var resp_headers []AblyRuleHeaders
	var headers []ably_control_go.Header

	switch t := plan.(type) {
	case *ably_control_go.HttpTarget:
		headers = t.Headers
	case *ably_control_go.HttpZapierTarget:
		headers = t.Headers
	case *ably_control_go.HttpCloudfareWorkerTarget:
		headers = t.Headers
	case *ably_control_go.HttpGoogleCloudFunctionTarget:
		headers = t.Headers
	case *ably_control_go.HttpAzureFunctionTarget:
		headers = t.Headers
	case *ably_control_go.AmqpTarget:
		headers = t.Headers
	case *ably_control_go.AmqpExternalTarget:
		headers = t.Headers
	}

	for _, b := range headers {
		item := AblyRuleHeaders{
			Name:  types.StringValue(b.Name),
			Value: types.StringValue(b.Value),
		}
		resp_headers = append(resp_headers, item)
	}

	return resp_headers
}

func GetKafkaAuthSchema(headers []AblyRuleHeaders) []ably_control_go.Header {
	var ret_headers []ably_control_go.Header
	for _, h := range headers {
		ret_headers = append(ret_headers, ably_control_go.Header{
			Name:  h.Name.ValueString(),
			Value: h.Value.ValueString(),
		})
	}

	return ret_headers
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
	plan_values := GetPlanRule(plan)

	// Creates a new Ably Rule by invoking the CreateRule function from the Client Library
	rule, err := r.Provider().client.CreateRule(plan.AppID.ValueString(), &plan_values)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating Resource '%s'", r.Name()),
			fmt.Sprintf("Could not create resource '%s', unexpected error: %s", r.Name(), err.Error()),
		)

		return
	}

	response_values := GetRuleResponse(&rule, &plan)

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, response_values)
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
	app_id := s.AppID.ValueString()
	rule_id := s.ID.ValueString()

	// Get Rule data
	rule, err := r.Provider().client.Rule(app_id, rule_id)

	if err != nil {
		if is_404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting Resource %s", r.Name()),
			fmt.Sprintf("Could not delete resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	response_values := GetRuleResponse(&rule, &state)

	// Sets state to app values.
	diags = resp.State.Set(ctx, &response_values)

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

	rule_values := GetPlanRule(plan)

	// Gets the Ably App ID and Ably Rule ID value for the resource
	app_id := plan.AppID.ValueString()
	rule_id := plan.ID.ValueString()

	// Update Ably Rule
	rule, err := r.Provider().client.UpdateRule(app_id, rule_id, &rule_values)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updading Resource %s", r.Name()),
			fmt.Sprintf("Could not update resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	response_values := GetRuleResponse(&rule, &plan)

	// Sets state to app values.
	diags = resp.State.Set(ctx, &response_values)

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
	app_id := state.AppID.ValueString()
	rule_id := state.ID.ValueString()

	err := r.Provider().client.DeleteRule(app_id, rule_id)
	if err != nil {
		if is_404(err) {
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
