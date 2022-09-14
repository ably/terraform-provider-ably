package ably_control

import (
	"context"
	"fmt"
	"strings"

	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleSqsType struct{}

// Get Rule Resource schema
func (r resourceRuleSqsType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
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
			"source": {
				Required:    true,
				Description: "object (rule_source)",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"channel_filter": {
						Type:     types.StringType,
						Required: true,
					},
					"type": {
						Type:     types.StringType,
						Required: true,
					},
				}),
			},
			"aws_authentication": {
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
			},
			"target": {
				Required:    true,
				Description: "object (rule_source)",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"region": {
						Type:        types.StringType,
						Optional:    true,
						Description: "The region is which AWS SQS is hosted",
					},
					"aws_account_id": {
						Type:        types.StringType,
						Optional:    true,
						Description: "Your AWS account ID",
					},
					"queue_name": {
						Type:        types.StringType,
						Optional:    true,
						Description: "The AWS SQS queue name",
					},
					"enveloped": {
						Type:        types.BoolType,
						Optional:    true,
						Description: "Delivered messages are wrapped in an Ably envelope by default that contains metadata about the message and its payload. The form of the envelope depends on whether it is part of a Webhook/Function or a Queue/Firehose rule. For everything besides Webhooks, you can ensure you only get the raw payload by unchecking 'Enveloped' when setting up the rule.",
					},
					"format": {
						Type:        types.StringType,
						Optional:    true,
						Description: "JSON provides a text-based encoding",
					},
				}),
			},
		},
	}, nil
}

func gen_plan_sqs_target_config(plan AblyRuleSqs, req_aws_auth ably_control_go.AwsAuthentication) ably_control_go.Target {
	target_config := &ably_control_go.AwsSqsTarget{
		Region:         plan.Target.Region,
		AwsAccountID:   plan.Target.AwsAccountID,
		QueueName:      plan.Target.QueueName,
		Enveloped:      plan.Target.Enveloped,
		Format:         format(plan.Target.Format),
		Authentication: req_aws_auth,
	}

	return target_config
}

// Get Plan Values
func get_plan_sqs_values(plan AblyRuleSqs) ably_control_go.NewRule {
	var req_aws_auth ably_control_go.AwsAuthentication

	assume_role_type := types.String{
		Value: "assumeRole",
	}
	credentials_type := types.String{
		Value: "credentials",
	}

	if plan.AwsAuth.AuthenticationMode.Value == assume_role_type.Value {
		req_aws_auth = ably_control_go.AwsAuthentication{
			Authentication: &ably_control_go.AuthenticationModeAssumeRole{
				AssumeRoleArn: plan.AwsAuth.RoleArn.Value,
			},
		}
	} else if plan.AwsAuth.AuthenticationMode.Value == credentials_type.Value {
		req_aws_auth = ably_control_go.AwsAuthentication{
			Authentication: &ably_control_go.AuthenticationModeCredentials{
				AccessKeyId:     plan.AwsAuth.AccessKeyId.Value,
				SecretAccessKey: plan.AwsAuth.SecretAccessKey.Value,
			},
		}
	}

	rule_values := ably_control_go.NewRule{
		Status:      plan.Status.Value,
		RequestMode: ably_control_go.Single, // This will always be single for Sqs rule type.
		Source: ably_control_go.Source{
			ChannelFilter: plan.Source.ChannelFilter.Value,
			Type:          source_type(plan.Source.Type),
		},
		Target: gen_plan_sqs_target_config(plan, req_aws_auth),
	}

	return rule_values
}

// Get Response Values
func get_sqs_response_values(ably_rule *ably_control_go.Rule, plan AblyRuleSqs) AblyRuleSqs {
	// Maps response body to resource schema attributes.
	channel_filter := types.String{
		Value: ably_rule.Source.ChannelFilter,
	}

	resp_source := AblyRuleSource{
		ChannelFilter: channel_filter,
		Type:          ably_rule.Source.Type,
	}

	var resp_target AblyRuleTargetSqs
	var resp_aws_auth AwsAuth
	var resp_access_key_id types.String
	var resp_secret_access_key types.String
	var resp_role_arn types.String

	if v, ok := ably_rule.Target.(*ably_control_go.AwsSqsTarget); ok {
		resp_target = AblyRuleTargetSqs{
			Region:       v.Region,
			AwsAccountID: v.AwsAccountID,
			QueueName:    v.QueueName,
			Enveloped:    v.Enveloped,
			Format:       v.Format,
		}
		if a, ok := v.Authentication.Authentication.(*ably_control_go.AuthenticationModeCredentials); ok {
			resp_access_key_id = types.String{
				Value: a.AccessKeyId,
			}

			resp_role_arn = types.String{
				Null: true,
			}

			resp_aws_auth = AwsAuth{

				AuthenticationMode: plan.AwsAuth.AuthenticationMode,
				AccessKeyId:        resp_access_key_id,
				SecretAccessKey:    plan.AwsAuth.SecretAccessKey,
				RoleArn:            resp_role_arn,
			}

		} else if a, ok := v.Authentication.Authentication.(*ably_control_go.AuthenticationModeAssumeRole); ok {
			resp_access_key_id = types.String{
				Null: true,
			}

			resp_secret_access_key = types.String{
				Null: true,
			}

			resp_role_arn = types.String{
				Value: a.AssumeRoleArn,
			}

			resp_aws_auth = AwsAuth{
				AuthenticationMode: plan.AwsAuth.AuthenticationMode,
				RoleArn:            resp_role_arn,
				AccessKeyId:        resp_access_key_id,
				SecretAccessKey:    resp_secret_access_key,
			}
		}
	}

	resp_rule := AblyRuleSqs{
		ID:      types.String{Value: ably_rule.ID},
		AppID:   types.String{Value: ably_rule.AppID},
		Status:  types.String{Value: ably_rule.Status},
		Source:  resp_source,
		Target:  resp_target,
		AwsAuth: resp_aws_auth,
	}

	return resp_rule
}

// New resource instance
func (r resourceRuleSqsType) NewResource(_ context.Context, p tfsdk_provider.Provider) (tfsdk_resource.Resource, diag.Diagnostics) {
	return resourceRuleSqs{
		p: *(p.(*provider)),
	}, nil
}

type resourceRuleSqs struct {
	p provider
}

// Create a new resource
func (r resourceRuleSqs) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	// Checks whether the provider and API Client are configured. If they are not, the provider responds with an error.
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply",
		)
		return
	}

	// Gets plan values
	var plan AblyRuleSqs
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan_values := get_plan_sqs_values(plan)

	// Creates a new Ably Rule by invoking the CreateRule function from the Client Library
	rule, err := r.p.client.CreateRule(plan.AppID.Value, &plan_values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	response_values := get_sqs_response_values(&rule, plan)

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, response_values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource
func (r resourceRuleSqs) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var state AblyRuleSqs
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID and Ably Rule ID value for the resource
	app_id := state.AppID.Value
	rule_id := state.ID.Value

	// Get Rule data
	rule, _ := r.p.client.Rule(app_id, rule_id)

	response_values := get_sqs_response_values(&rule, state)

	// Sets state to app values.
	diags = resp.State.Set(ctx, &response_values)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update resource
func (r resourceRuleSqs) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	// Gets plan values
	var plan AblyRuleSqs
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var state AblyRuleSqs
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	rule_values := get_plan_sqs_values(plan)

	// Gets the Ably App ID and Ably Rule ID value for the resource
	app_id := state.AppID.Value
	rule_id := state.ID.Value

	// Update Ably Rule
	rule, _ := r.p.client.UpdateRule(app_id, rule_id, &rule_values)

	response_values := get_sqs_response_values(&rule, plan)

	// Sets state to app values.
	diags = resp.State.Set(ctx, &response_values)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceRuleSqs) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var state AblyRuleSqs
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID and Ably Rule ID value for the resource
	app_id := state.AppID.Value
	rule_id := state.ID.Value

	err := r.p.client.DeleteRule(app_id, rule_id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Resource",
			"Could not delete resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

// Import resource
func (r resourceRuleSqs) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// identifier should be in the format app_id,key_id
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: 'app_id,rule_id'. Got: %q", req.ID),
		)
		return
	}
	// Recent PR in TF Plugin Framework for paths but Hashicorp examples not updated - https://github.com/hashicorp/terraform-plugin-framework/pull/390
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
