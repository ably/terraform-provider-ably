package ably_control

import (
	ably_control_go "github.com/ably/ably-control-go"
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
		var headers []ably_control_go.Header
		for _, h := range t.Headers {
			headers = append(headers, ably_control_go.Header{
				Name:  h.Name.Value,
				Value: h.Value.Value,
			})
		}

		target = &ably_control_go.HttpZapierTarget{
			Url:          t.Url,
			Headers:      headers,
			SigningKeyID: t.SigningKeyId,
		}

	case *AblyRuleTargetIFTTT:
		target = &ably_control_go.HttpIftttTarget{
			WebhookKey: t.WebhookKey,
			EventName:  t.EventName,
		}
	case *AblyRuleTargetGoogleFunction:
		var headers []ably_control_go.Header
		for _, h := range t.Headers {
			headers = append(headers, ably_control_go.Header{
				Name:  h.Name.Value,
				Value: h.Value.Value,
			})
		}

		target = &ably_control_go.HttpGoogleCloudFunctionTarget{
			Region:       t.Region,
			ProjectID:    t.ProjectID,
			FunctionName: t.FunctionName,
			Headers:      headers,
			SigningKeyID: t.SigningKeyId,
			Enveloped:    t.Enveloped,
			Format:       t.Format,
		}
	}

	rule_values := ably_control_go.NewRule{
		Status:      plan.Status.Value,
		RequestMode: GetRequestMode(plan),
		Source: ably_control_go.Source{
			ChannelFilter: plan.Source.ChannelFilter.Value,
			Type:          GetSourceType(plan.Source.Type),
		},
		Target: target,
	}

	return rule_values
}

func GetRequestMode(plan AblyRule) ably_control_go.RequestMode {
	switch plan.RequestMode.Value {
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
		headers := GetHeaders(v)

		resp_target = &AblyRuleTargetZapier{
			Url:          v.Url,
			SigningKeyId: v.SigningKeyID,
			Headers:      headers,
		}
	case *ably_control_go.HttpIftttTarget:
		resp_target = &AblyRuleTargetIFTTT{
			EventName:  v.EventName,
			WebhookKey: v.WebhookKey,
		}
	case *ably_control_go.HttpGoogleCloudFunctionTarget:
		headers := GetHeaders(v)

		resp_target = &AblyRuleTargetGoogleFunction{
			Region:       v.Region,
			ProjectID:    v.ProjectID,
			FunctionName: v.FunctionName,
			Headers:      headers,
			SigningKeyId: v.SigningKeyID,
			Enveloped:    v.Enveloped,
			Format:       v.Format,
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
		ID:          types.String{Value: ably_rule.ID},
		AppID:       types.String{Value: ably_rule.AppID},
		Status:      types.String{Value: ably_rule.Status},
		Source:      resp_source,
		Target:      resp_target,
		RequestMode: types.String{Value: string(ably_rule.RequestMode)},
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
					DefaultAttribute(types.String{Value: "single"}),
					tfsdk_resource.UseStateForUnknown(),
				},
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

func GetHeaders(plan ably_control_go.Target) []AblyRuleHeaders {
	var resp_headers []AblyRuleHeaders
	var headers []ably_control_go.Header

	switch t := plan.(type) {
	case *ably_control_go.HttpZapierTarget:
		headers = t.Headers
	case *ably_control_go.HttpGoogleCloudFunctionTarget:
		headers = t.Headers
	}

	for _, b := range headers {
		item := AblyRuleHeaders{
			Name:  types.String{Value: b.Name},
			Value: types.String{Value: b.Value},
		}
		resp_headers = append(resp_headers, item)
	}

	return resp_headers
}
