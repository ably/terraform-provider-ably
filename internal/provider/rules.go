package ably_control

import (
	ably_control_go "github.com/ably/ably-control-go"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func get_plan_aws_auth(plan AblyRule) ably_control_go.AwsAuthentication {
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

	if auth.AuthenticationMode.Value == "assumeRole" {
		control_auth = ably_control_go.AwsAuthentication{
			Authentication: &ably_control_go.AuthenticationModeAssumeRole{
				AssumeRoleArn: auth.RoleArn.Value,
			},
		}
	} else if auth.AuthenticationMode.Value == "credentials" {
		control_auth = ably_control_go.AwsAuthentication{
			Authentication: &ably_control_go.AuthenticationModeCredentials{
				AccessKeyId:     auth.AccessKeyId.Value,
				SecretAccessKey: auth.SecretAccessKey.Value,
			},
		}
	}

	return control_auth
}

// converts rule from terraform format to control sdk format
func get_plan_rule(plan AblyRule) ably_control_go.NewRule {
	var target ably_control_go.Target

	switch t := plan.Target.(type) {
	case *AblyRuleTargetKinesis:
		target = &ably_control_go.AwsKinesisTarget{
			Region:         t.Region,
			StreamName:     t.StreamName,
			PartitionKey:   t.PartitionKey,
			Authentication: get_plan_aws_auth(plan),
			Enveloped:      t.Enveloped,
			Format:         t.Format,
		}
	case *AblyRuleTargetSqs:
		target = &ably_control_go.AwsSqsTarget{
			Region:         t.Region,
			AwsAccountID:   t.AwsAccountID,
			QueueName:      t.QueueName,
			Authentication: get_plan_aws_auth(plan),
			Enveloped:      t.Enveloped,
			Format:         t.Format,
		}
	case *AblyRuleTargetLambda:
		target = &ably_control_go.AwsLambdaTarget{
			Region:         t.Region,
			FunctionName:   t.FunctionName,
			Authentication: get_plan_aws_auth(plan),
			Enveloped:      t.Enveloped,
		}
	}

	rule_values := ably_control_go.NewRule{
		Status:      plan.Status.Value,
		RequestMode: ably_control_go.Single, // This will always be single for Kinesis rule type.
		Source: ably_control_go.Source{
			ChannelFilter: plan.Source.ChannelFilter.Value,
			Type:          source_type(plan.Source.Type),
		},
		Target: target,
	}

	return rule_values
}

// Maps response body to resource schema attributes.
// Using plan to fill in values that the api does not return.
func get_aws_auth(auth *ably_control_go.AwsAuthentication, plan *AblyRule) AwsAuth {
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
			AccessKeyId:        types.String{Value: a.AccessKeyId},
			SecretAccessKey:    plan_auth.SecretAccessKey,
			RoleArn:            types.String{Null: true},
		}
	case *ably_control_go.AuthenticationModeAssumeRole:
		resp_aws_auth = AwsAuth{
			AuthenticationMode: plan_auth.AuthenticationMode,
			RoleArn:            types.String{Value: a.AssumeRoleArn},
			AccessKeyId:        types.String{Null: true},
			SecretAccessKey:    types.String{Null: true},
		}
	}

	return resp_aws_auth
}

// Maps response body to resource schema attributes.
// Using plan to fill in values that the api does not return.
func get_rule_response(ably_rule *ably_control_go.Rule, plan *AblyRule) AblyRule {
	var resp_target interface{}

	switch v := ably_rule.Target.(type) {
	case *ably_control_go.AwsKinesisTarget:
		resp_target = &AblyRuleTargetKinesis{
			Region:       v.Region,
			StreamName:   v.StreamName,
			PartitionKey: v.PartitionKey,
			AwsAuth:      get_aws_auth(&v.Authentication, plan),
			Enveloped:    v.Enveloped,
			Format:       v.Format,
		}
	case *ably_control_go.AwsSqsTarget:
		resp_target = &AblyRuleTargetSqs{
			Region:       v.Region,
			AwsAccountID: v.AwsAccountID,
			QueueName:    v.QueueName,
			AwsAuth:      get_aws_auth(&v.Authentication, plan),
			Enveloped:    v.Enveloped,
			Format:       v.Format,
		}
	case *ably_control_go.AwsLambdaTarget:
		resp_target = &AblyRuleTargetLambda{
			Region:       v.Region,
			FunctionName: v.FunctionName,
			AwsAuth:      get_aws_auth(&v.Authentication, plan),
			Enveloped:    v.Enveloped,
		}
	}

	channel_filter := types.String{
		Value: ably_rule.Source.ChannelFilter,
	}

	resp_source := AblyRuleSource{
		ChannelFilter: channel_filter,
		Type:          ably_rule.Source.Type,
	}

	resp_rule := AblyRule{
		ID:     types.String{Value: ably_rule.ID},
		AppID:  types.String{Value: ably_rule.AppID},
		Status: types.String{Value: ably_rule.Status},
		Source: resp_source,
		Target: resp_target,
	}

	return resp_rule
}

func GetRuleSchema(target map[string]tfsdk.Attribute) tfsdk.Schema {
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
