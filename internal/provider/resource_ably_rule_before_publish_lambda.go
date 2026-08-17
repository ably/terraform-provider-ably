// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/ably/terraform-provider-ably/internal/provider/codegen/resource_rule_before_publish_lambda"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// beforePublishLambdaRuleType is the Control API discriminator for
// before-publish AWS Lambda rules.
const beforePublishLambdaRuleType = "aws/lambda/before-publish"

// AblyRuleBeforePublishLambdaAuth mirrors control.AWSAuthentication.
//
// Note the attribute names differ from the older hand-written AWS rules
// (ably_rule_lambda et al), which use `mode` and `role_arn`. These come from the
// generated schema, which follows the Control API spec
// (`authentication_mode`, `assume_role_arn`).
type AblyRuleBeforePublishLambdaAuth struct {
	AuthenticationMode types.String `tfsdk:"authentication_mode"`
	AccessKeyID        types.String `tfsdk:"access_key_id"`
	SecretAccessKey    types.String `tfsdk:"secret_access_key"`
	AssumeRoleArn      types.String `tfsdk:"assume_role_arn"`
}

// AblyRuleBeforePublishLambdaTarget mirrors
// control.BeforePublishAWSLambdaTarget.
type AblyRuleBeforePublishLambdaTarget struct {
	Region         types.String                     `tfsdk:"region"`
	FunctionName   types.String                     `tfsdk:"function_name"`
	Authentication *AblyRuleBeforePublishLambdaAuth `tfsdk:"authentication"`
}

// AblyRuleBeforePublishLambda is the tfsdk model for the before-publish AWS
// Lambda rule.
//
// This is the one before-publish family that takes a `source`, and it is
// optional here rather than required as it is for webhook/firehose rules. It
// still has no `request_mode`, so it does not use the AblyRule plumbing; see
// before_publish_rules.go.
type AblyRuleBeforePublishLambda struct {
	ID                  types.String                       `tfsdk:"id"`
	AppID               types.String                       `tfsdk:"app_id"`
	Status              types.String                       `tfsdk:"status"`
	InvocationMode      types.String                       `tfsdk:"invocation_mode"`
	ChatRoomFilter      types.String                       `tfsdk:"chat_room_filter"`
	BeforePublishConfig *AblyBeforePublishConfig           `tfsdk:"before_publish_config"`
	Source              *AblyRuleSource                    `tfsdk:"source"`
	Target              *AblyRuleBeforePublishLambdaTarget `tfsdk:"target"`
}

type ResourceRuleBeforePublishLambda struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleBeforePublishLambda{}
var _ resource.ResourceWithImportState = &ResourceRuleBeforePublishLambda{}

// Schema defines the schema for the resource.
//
// PORTED ONTO GENERATED CODE (see DEVELOPMENT.md "Porting a resource onto
// generated code").
func (r ResourceRuleBeforePublishLambda) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_rule_before_publish_lambda.RuleBeforePublishLambdaResourceSchema(ctx)
	stripNestedCustomTypes(&s, "before_publish_config", "source", "target")
	stripSingleNestedCustomType(&s, "target", "authentication")

	s.MarkdownDescription = "The `ably_rule_before_publish_lambda` resource allows you to create and manage an Ably integration rule that invokes an AWS Lambda function before a message is published, so the function can allow, reject or amend it. Read more at https://ably.com/docs/chat/moderation"

	resp.Schema = s
}

func (r ResourceRuleBeforePublishLambda) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_before_publish_lambda"
}

func (r *ResourceRuleBeforePublishLambda) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleBeforePublishLambda) Name() string {
	return "Before-publish AWS Lambda"
}

func (r ResourceRuleBeforePublishLambda) crud() beforePublishCRUD[AblyRuleBeforePublishLambda] {
	return beforePublishCRUD[AblyRuleBeforePublishLambda]{
		p:        r.p,
		name:     r.Name(),
		appID:    func(m AblyRuleBeforePublishLambda) string { return m.AppID.ValueString() },
		id:       func(m AblyRuleBeforePublishLambda) string { return m.ID.ValueString() },
		post:     getPlanBeforePublishLambdaPost,
		response: getBeforePublishLambdaResponse,
	}
}

// beforePublishLambdaAuthPost converts the plan's authentication block into the
// control type. Only the fields the chosen mode uses are sent: the API rejects
// credentials alongside an assume-role ARN and vice versa. This mirrors
// GetPlanAwsAuth for the webhook/firehose AWS rules.
func beforePublishLambdaAuthPost(auth *AblyRuleBeforePublishLambdaAuth) control.AWSAuthentication {
	if auth == nil {
		return control.AWSAuthentication{}
	}

	switch control.AWSAuthMode(auth.AuthenticationMode.ValueString()) {
	case control.AWSAuthModeAssumeRole:
		return control.AWSAuthentication{
			AuthenticationMode: string(control.AWSAuthModeAssumeRole),
			AssumeRoleArn:      auth.AssumeRoleArn.ValueString(),
		}
	case control.AWSAuthModeCredentials:
		return control.AWSAuthentication{
			AuthenticationMode: string(control.AWSAuthModeCredentials),
			AccessKeyID:        auth.AccessKeyID.ValueString(),
			SecretAccessKey:    auth.SecretAccessKey.ValueString(),
		}
	}

	return control.AWSAuthentication{}
}

// getPlanBeforePublishLambdaPost converts the plan model into the Control API
// create body.
func getPlanBeforePublishLambdaPost(_ context.Context, plan AblyRuleBeforePublishLambda) (any, diag.Diagnostics) {
	var source *control.RuleSource
	if plan.Source != nil {
		source = &control.RuleSource{
			ChannelFilter: plan.Source.ChannelFilter.ValueString(),
			Type:          plan.Source.Type.ValueString(),
		}
	}

	return control.BeforePublishAWSLambdaRulePost{
		Status:              plan.Status.ValueString(),
		RuleType:            beforePublishLambdaRuleType,
		InvocationMode:      plan.InvocationMode.ValueString(),
		ChatRoomFilter:      plan.ChatRoomFilter.ValueString(),
		BeforePublishConfig: beforePublishConfigPost(plan.BeforePublishConfig),
		Source:              source,
		Target: control.BeforePublishAWSLambdaTarget{
			Region:         plan.Target.Region.ValueString(),
			FunctionName:   plan.Target.FunctionName.ValueString(),
			Authentication: beforePublishLambdaAuthPost(plan.Target.Authentication),
		},
	}, nil
}

// beforePublishLambdaAuthResponse maps the API's authentication object back onto
// the tfsdk model.
//
// The Control API never returns secretAccessKey, so it has to be preserved from
// the plan (create/update) or prior state (read); reading it back as null would
// abort the apply with "inconsistent result after apply" on a sensitive
// attribute. The unused fields for the mode in force are explicitly null so a
// mode switch does not leave stale values in state.
func beforePublishLambdaAuthResponse(auth control.AWSAuthentication, prior *AblyRuleBeforePublishLambdaAuth) *AblyRuleBeforePublishLambdaAuth {
	priorSecret := types.StringNull()
	if prior != nil {
		priorSecret = prior.SecretAccessKey
	}

	switch control.AWSAuthMode(auth.AuthenticationMode) {
	case control.AWSAuthModeCredentials:
		return &AblyRuleBeforePublishLambdaAuth{
			AuthenticationMode: types.StringValue(auth.AuthenticationMode),
			AccessKeyID:        stringOrNull(auth.AccessKeyID),
			SecretAccessKey:    priorSecret,
			AssumeRoleArn:      types.StringNull(),
		}
	case control.AWSAuthModeAssumeRole:
		return &AblyRuleBeforePublishLambdaAuth{
			AuthenticationMode: types.StringValue(auth.AuthenticationMode),
			AccessKeyID:        types.StringNull(),
			SecretAccessKey:    types.StringNull(),
			AssumeRoleArn:      stringOrNull(auth.AssumeRoleArn),
		}
	}

	return &AblyRuleBeforePublishLambdaAuth{
		AuthenticationMode: stringOrNull(auth.AuthenticationMode),
		AccessKeyID:        stringOrNull(auth.AccessKeyID),
		SecretAccessKey:    priorSecret,
		AssumeRoleArn:      stringOrNull(auth.AssumeRoleArn),
	}
}

// getBeforePublishLambdaResponse maps an API rule response back onto the tfsdk
// model, using the plan or prior state for the write-only secret_access_key.
func getBeforePublishLambdaResponse(_ context.Context, rule *control.RuleResponse, prior *AblyRuleBeforePublishLambda) (AblyRuleBeforePublishLambda, diag.Diagnostics) {
	diags := checkRuleType(rule, beforePublishLambdaRuleType)
	if diags.HasError() {
		return AblyRuleBeforePublishLambda{}, diags
	}

	target, err := unmarshalTarget[control.BeforePublishAWSLambdaTarget](rule.Target)
	if err != nil {
		diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal %s target: %s", beforePublishLambdaRuleType, err.Error()))
		return AblyRuleBeforePublishLambda{}, diags
	}

	var priorAuth *AblyRuleBeforePublishLambdaAuth
	if prior != nil && prior.Target != nil {
		priorAuth = prior.Target.Authentication
	}

	var source *AblyRuleSource
	if rule.Source != nil {
		source = &AblyRuleSource{
			ChannelFilter: types.StringValue(rule.Source.ChannelFilter),
			Type:          types.StringValue(rule.Source.Type),
		}
	}

	return AblyRuleBeforePublishLambda{
		ID:                  types.StringValue(rule.ID),
		AppID:               types.StringValue(rule.AppID),
		Status:              types.StringValue(rule.Status),
		InvocationMode:      stringOrNull(rule.InvocationMode),
		ChatRoomFilter:      stringOrNull(rule.ChatRoomFilter),
		BeforePublishConfig: beforePublishConfigResponse(rule.BeforePublishConfig),
		Source:              source,
		Target: &AblyRuleBeforePublishLambdaTarget{
			Region:         stringOrNull(target.Region),
			FunctionName:   stringOrNull(target.FunctionName),
			Authentication: beforePublishLambdaAuthResponse(target.Authentication, priorAuth),
		},
	}, diags
}

// Create creates a new resource.
func (r ResourceRuleBeforePublishLambda) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.crud().create(ctx, req, resp)
}

// Read reads the resource.
func (r ResourceRuleBeforePublishLambda) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.crud().read(ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceRuleBeforePublishLambda) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.crud().update(ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceRuleBeforePublishLambda) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.crud().delete(ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceRuleBeforePublishLambda) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
