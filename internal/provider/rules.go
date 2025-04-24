// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"
	"strings"

	control "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

// GetPlanRule converts rule from terraform format to control sdk format.
func GetPlanRule(plan AblyRule) control.NewRule {
	var target control.Target

	switch t := plan.Target.(type) {
	case *AblyRuleTargetKinesis:
		target = &control.AwsKinesisTarget{
			Region:         t.Region.ValueString(),
			StreamName:     t.StreamName.ValueString(),
			PartitionKey:   t.PartitionKey.ValueString(),
			Authentication: GetPlanAwsAuth(plan),
			Enveloped:      t.Enveloped.ValueBool(),
			Format:         control.Format(t.Format.ValueString()),
		}
	case *AblyRuleTargetSqs:
		target = &control.AwsSqsTarget{
			Region:         t.Region.ValueString(),
			AwsAccountID:   t.AwsAccountID.ValueString(),
			QueueName:      t.QueueName.ValueString(),
			Authentication: GetPlanAwsAuth(plan),
			Enveloped:      t.Enveloped.ValueBool(),
			Format:         control.Format(t.Format.ValueString()),
		}
	case *AblyRuleTargetLambda:
		target = &control.AwsLambdaTarget{
			Region:         t.Region.ValueString(),
			FunctionName:   t.FunctionName.ValueString(),
			Authentication: GetPlanAwsAuth(plan),
			Enveloped:      t.Enveloped.ValueBool(),
		}
	case *AblyRuleTargetZapier:
		target = &control.HttpZapierTarget{
			Url:          t.Url.ValueString(),
			Headers:      GetHeaders(t.Headers),
			SigningKeyID: t.SigningKeyId.ValueString(),
		}
	case *AblyRuleTargetCloudflareWorker:
		target = &control.HttpCloudfareWorkerTarget{
			Url:          t.Url.ValueString(),
			Headers:      GetHeaders(t.Headers),
			SigningKeyID: t.SigningKeyId.ValueString(),
		}
	case *AblyRuleTargetPulsar:
		target = &control.PulsarTarget{
			RoutingKey:    t.RoutingKey.ValueString(),
			Topic:         t.Topic.ValueString(),
			ServiceURL:    t.ServiceURL.ValueString(),
			TlsTrustCerts: sliceString(t.TlsTrustCerts),
			Authentication: control.PulsarAuthentication{
				AuthenticationMode: control.PularAuthenticationMode(t.Authentication.Mode.ValueString()),
				Token:              t.Authentication.Token.ValueString(),
			},
			Enveloped: t.Enveloped.ValueBool(),
			Format:    control.Format(t.Format.ValueString()),
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
			Url:          t.Url.ValueString(),
			Headers:      headers,
			SigningKeyID: t.SigningKeyId.ValueString(),
			Format:       control.Format(t.Format.ValueString()),
			Enveloped:    t.Enveloped.ValueBool(),
		}
	case *AblyRuleTargetIFTTT:
		target = &control.HttpIftttTarget{
			WebhookKey: t.WebhookKey.ValueString(),
			EventName:  t.EventName.ValueString(),
		}
	case *AblyRuleTargetAzureFunction:
		target = &control.HttpAzureFunctionTarget{
			AzureAppID:        t.AzureAppID.ValueString(),
			AzureFunctionName: t.AzureFunctionName.ValueString(),
			Headers:           GetHeaders(t.Headers),
			SigningKeyID:      t.SigningKeyID.ValueString(),
			Format:            control.Format(t.Format.ValueString()),
		}
	case *AblyRuleTargetGoogleFunction:
		target = &control.HttpGoogleCloudFunctionTarget{
			Region:       t.Region.ValueString(),
			ProjectID:    t.ProjectID.ValueString(),
			FunctionName: t.FunctionName.ValueString(),
			Headers:      GetHeaders(t.Headers),
			SigningKeyID: t.SigningKeyId.ValueString(),
			Enveloped:    t.Enveloped.ValueBool(),
			Format:       control.Format(t.Format.ValueString()),
		}

	case *AblyRuleTargetKafka:
		target = &control.KafkaTarget{
			RoutingKey: t.RoutingKey.ValueString(),
			Brokers:    sliceString(t.Brokers),
			Authentication: control.KafkaAuthentication{
				Sasl: control.Sasl{
					Mechanism: control.SaslMechanism(t.KafkaAuthentication.Sasl.Mechanism.ValueString()),
					Username:  t.KafkaAuthentication.Sasl.Username.ValueString(),
					Password:  t.KafkaAuthentication.Sasl.Password.ValueString(),
				},
			},
			Enveloped: t.Enveloped.ValueBool(),
			Format:    control.Format(t.Format.ValueString()),
		}
	case *AblyRuleTargetAMQP:
		target = &control.AmqpTarget{
			QueueID:   t.QueueID.ValueString(),
			Headers:   GetHeaders(t.Headers),
			Enveloped: t.Enveloped.ValueBool(),
			Format:    control.Format(t.Format.ValueString()),
		}
	case *AblyRuleTargetAMQPExternal:
		target = &control.AmqpExternalTarget{
			Url:                t.Url.ValueString(),
			RoutingKey:         t.RoutingKey.ValueString(),
			Exchange:           t.Exchange.ValueString(),
			MandatoryRoute:     t.MandatoryRoute.ValueBool(),
			PersistentMessages: t.PersistentMessages.ValueBool(),
			MessageTTL:         int(t.MessageTtl.ValueInt64()),
			Headers:            GetHeaders(t.Headers),
			Enveloped:          t.Enveloped.ValueBool(),
			Format:             control.Format(t.Format.ValueString()),
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

// GetAwsAuth converts AWS authentication from control SDK format to terraform format.
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

// GetRuleResponse maps response body to resource schema attributes..
// Using plan to fill in values that the api does not return.
func GetRuleResponse(ablyRule *control.Rule, plan *AblyRule) AblyRule {
	var respTarget any

	switch v := ablyRule.Target.(type) {
	case *control.AwsKinesisTarget:
		respTarget = &AblyRuleTargetKinesis{
			Region:       types.StringValue(v.Region),
			StreamName:   types.StringValue(v.StreamName),
			PartitionKey: types.StringValue(v.PartitionKey),
			AwsAuth:      GetAwsAuth(&v.Authentication, plan),
			Enveloped:    types.BoolValue(v.Enveloped),
			Format:       types.StringValue(string(v.Format)),
		}
	case *control.AwsSqsTarget:
		respTarget = &AblyRuleTargetSqs{
			Region:       types.StringValue(v.Region),
			AwsAccountID: types.StringValue(v.AwsAccountID),
			QueueName:    types.StringValue(v.QueueName),
			AwsAuth:      GetAwsAuth(&v.Authentication, plan),
			Enveloped:    types.BoolValue(v.Enveloped),
			Format:       types.StringValue(string(v.Format)),
		}
	case *control.AwsLambdaTarget:
		respTarget = &AblyRuleTargetLambda{
			Region:       types.StringValue(v.Region),
			FunctionName: types.StringValue(v.FunctionName),
			AwsAuth:      GetAwsAuth(&v.Authentication, plan),
			Enveloped:    types.BoolValue(v.Enveloped),
		}
	case *control.HttpZapierTarget:
		headers := ToHeaders(v)

		respTarget = &AblyRuleTargetZapier{
			Url:          types.StringValue(v.Url),
			SigningKeyId: types.StringValue(v.SigningKeyID),
			Headers:      headers,
		}
	case *control.HttpCloudfareWorkerTarget:
		headers := ToHeaders(v)

		respTarget = &AblyRuleTargetCloudflareWorker{
			Url:          types.StringValue(v.Url),
			SigningKeyId: types.StringValue(v.SigningKeyID),
			Headers:      headers,
		}
	case *control.PulsarTarget:
		respTarget = &AblyRuleTargetPulsar{
			RoutingKey:    types.StringValue(v.RoutingKey),
			Topic:         types.StringValue(v.Topic),
			ServiceURL:    types.StringValue(v.ServiceURL),
			TlsTrustCerts: toTypedStringSlice(v.TlsTrustCerts),
			Authentication: PulsarAuthentication{
				Mode:  types.StringValue(string(v.Authentication.AuthenticationMode)),
				Token: types.StringValue(v.Authentication.Token),
			},
			Enveloped: types.BoolValue(v.Enveloped),
			Format:    types.StringValue(string(v.Format)),
		}
	case *control.HttpIftttTarget:
		respTarget = &AblyRuleTargetIFTTT{
			EventName:  types.StringValue(v.EventName),
			WebhookKey: types.StringValue(v.WebhookKey),
		}
	case *control.HttpGoogleCloudFunctionTarget:
		headers := ToHeaders(v)

		respTarget = &AblyRuleTargetGoogleFunction{
			Region:       types.StringValue(v.Region),
			ProjectID:    types.StringValue(v.ProjectID),
			FunctionName: types.StringValue(v.FunctionName),
			Headers:      headers,
			SigningKeyId: types.StringValue(v.SigningKeyID),
			Enveloped:    types.BoolValue(v.Enveloped),
			Format:       types.StringValue(string(v.Format)),
		}
	case *control.HttpAzureFunctionTarget:
		headers := ToHeaders(v)

		respTarget = &AblyRuleTargetAzureFunction{
			AzureAppID:        types.StringValue(v.AzureAppID),
			AzureFunctionName: types.StringValue(v.AzureFunctionName),
			Headers:           headers,
			SigningKeyID:      types.StringValue(v.SigningKeyID),
			Format:            types.StringValue(string(v.Format)),
		}
	case *control.HttpTarget:
		headers := ToHeaders(v)

		respTarget = &AblyRuleTargetHTTP{
			Url:          types.StringValue(v.Url),
			Headers:      headers,
			SigningKeyId: types.StringValue(v.SigningKeyID),
			Format:       types.StringValue(string(v.Format)),
			Enveloped:    types.BoolValue(v.Enveloped),
		}
	case *control.KafkaTarget:
		respTarget = &AblyRuleTargetKafka{
			RoutingKey: types.StringValue(v.RoutingKey),
			Brokers:    toTypedStringSlice(v.Brokers),
			KafkaAuthentication: KafkaAuthentication{
				Sasl{
					Mechanism: types.StringValue(string(v.Authentication.Sasl.Mechanism)),
					Username:  types.StringValue(v.Authentication.Sasl.Username),
					Password:  types.StringValue(v.Authentication.Sasl.Password),
				},
			},
			Enveloped: types.BoolValue(v.Enveloped),
			Format:    types.StringValue(string(v.Format)),
		}
	case *control.AmqpTarget:
		headers := ToHeaders(v)

		respTarget = &AblyRuleTargetAMQP{
			QueueID:   types.StringValue(v.QueueID),
			Headers:   headers,
			Enveloped: types.BoolValue(v.Enveloped),
			Format:    types.StringValue(string(v.Format)),
		}
	case *control.AmqpExternalTarget:
		headers := ToHeaders(v)
		ttl := types.Int64Null()
		if v.MessageTTL != 0 {
			ttl = types.Int64Value(int64(v.MessageTTL))
		}

		respTarget = &AblyRuleTargetAMQPExternal{
			Url:                types.StringValue(v.Url),
			RoutingKey:         types.StringValue(v.RoutingKey),
			Exchange:           types.StringValue(v.Exchange),
			MandatoryRoute:     types.BoolValue(v.MandatoryRoute),
			PersistentMessages: types.BoolValue(v.PersistentMessages),
			MessageTtl:         ttl,
			Headers:            headers,
			Enveloped:          types.BoolValue(v.Enveloped),
			Format:             types.StringValue(string(v.Format)),
		}
	}

	channelFilter := types.StringNull()
	if ablyRule.Source.ChannelFilter != "" {
		channelFilter = types.StringValue(
			ablyRule.Source.ChannelFilter,
		)
	}

	respSource := AblyRuleSource{
		ChannelFilter: channelFilter,
		Type:          types.StringValue(string(ablyRule.Source.Type)),
	}

	respRule := AblyRule{
		ID:          types.StringValue(ablyRule.ID),
		AppID:       types.StringValue(ablyRule.AppID),
		Status:      types.StringValue(ablyRule.Status),
		Source:      &respSource,
		Target:      respTarget,
		RequestMode: types.StringValue(string(ablyRule.RequestMode)),
	}

	return respRule
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
				Description: "The status of the rule. Rules can be enabled or disabled.",
			},
			"request_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "This is Single Request mode or Batch Request mode. Single Request mode sends each event separately to the endpoint specified by the rule",
				PlanModifiers: []planmodifier.String{
					DefaultStringAttribute(types.StringValue("single")),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source": schema.SingleNestedAttribute{
				Required:    true,
				Description: "object (rule_source)",
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
				Description: "object (rule_source)",
				Attributes:  target,
			},
		},
	}
}

func GetAwsAuthSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Required:    true,
		Description: "object (rule_source)",
		Attributes: map[string]schema.Attribute{
			"mode": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Authentication method. Use 'credentials' or 'assumeRole'",
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
	}
}

func GetSourceType(mode types.String) control.SourceType {
	switch mode {
	case types.StringValue("channel.message"):
		return control.ChannelMessage
	case types.StringValue("channel.presence"):
		return control.ChannelPresence
	case types.StringValue("channel.lifecycle"):
		return control.ChannelLifeCycle
	case types.StringValue("channel.occupancy"):
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

type Rule interface {
	Provider() *AblyProvider
	Name() string
}

// CreateRule creates a new rule resource.
func CreateRule[T any](r Rule, ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
