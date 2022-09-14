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

type resourceRuleKinesisType struct{}

// Get Rule Resource schema
func (r resourceRuleKinesisType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetRuleSchema(
		map[string]tfsdk.Attribute{
			"region": {
				Type:     types.StringType,
				Optional: true,
			},
			"stream_name": {
				Type:     types.StringType,
				Optional: true,
			},
			"partition_key": {
				Type:     types.StringType,
				Optional: true,
			},
			"enveloped": {
				Type:     types.BoolType,
				Optional: true,
			},
			"format": {
				Type:     types.StringType,
				Optional: true,
			},
			"authentication": GetAwsAuthSchema(),
		},
	), nil
}

func gen_plan_kinesis_target_config(plan AblyRule, req_aws_auth ably_control_go.AwsAuthentication) ably_control_go.Target {
	var target_config ably_control_go.Target

	switch target := plan.Target.(type) {
	case *AblyRuleTargetKinesis:
		target_config = &ably_control_go.AwsKinesisTarget{
			Region:         target.Region,
			StreamName:     target.StreamName,
			PartitionKey:   target.PartitionKey,
			Enveloped:      target.Enveloped,
			Format:         format(target.Format),
			Authentication: req_aws_auth,
		}
	}

	return target_config
}

func source_type(mode ably_control_go.SourceType) ably_control_go.SourceType {
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

func format(format ably_control_go.Format) ably_control_go.Format {
	switch format {
	case "json":
		return ably_control_go.Json
	case "msgpack":
		return ably_control_go.MsgPack
	default:
		return ably_control_go.Json
	}
}

// New resource instance
func (r resourceRuleKinesisType) NewResource(_ context.Context, p tfsdk_provider.Provider) (tfsdk_resource.Resource, diag.Diagnostics) {
	return resourceRule{
		p: *(p.(*provider)),
	}, nil
}

type resourceRule struct {
	p provider
}

// Create a new resource
func (r resourceRule) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	// Checks whether the provider and API Client are configured. If they are not, the provider responds with an error.
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply",
		)
		return
	}

	// Gets plan values
	var p AblyRuleDecoder[*AblyRuleTargetKinesis]
	diags := req.Plan.Get(ctx, &p)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan := p.Rule()
	plan_values := get_plan_rule(plan)

	// Creates a new Ably Rule by invoking the CreateRule function from the Client Library
	rule, err := r.p.client.CreateRule(plan.AppID.Value, &plan_values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	response_values := get_rule_response(&rule, &plan)

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, response_values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource
func (r resourceRule) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var s AblyRuleDecoder[*AblyRuleTargetKinesis]
	diags := req.State.Get(ctx, &s)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	state := s.Rule()

	// Gets the Ably App ID and Ably Rule ID value for the resource
	app_id := state.AppID.Value
	rule_id := state.ID.Value

	// Get Rule data
	rule, _ := r.p.client.Rule(app_id, rule_id)

	response_values := get_rule_response(&rule, &state)

	// Sets state to app values.
	diags = resp.State.Set(ctx, &response_values)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update resource
func (r resourceRule) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	// Gets plan values
	var p AblyRuleDecoder[*AblyRuleTargetKinesis]
	diags := req.Plan.Get(ctx, &p)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var s AblyRuleDecoder[*AblyRuleTargetKinesis]
	diags = req.State.Get(ctx, &s)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	state := s.Rule()
	plan := p.Rule()

	rule_values := get_plan_rule(plan)

	// Gets the Ably App ID and Ably Rule ID value for the resource
	app_id := state.AppID.Value
	rule_id := state.ID.Value

	// Update Ably Rule
	rule, _ := r.p.client.UpdateRule(app_id, rule_id, &rule_values)

	response_values := get_rule_response(&rule, &plan)

	// Sets state to app values.
	diags = resp.State.Set(ctx, &response_values)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceRule) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var s AblyRuleDecoder[*AblyRuleTargetKinesis]
	diags := req.State.Get(ctx, &s)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	state := s.Rule()

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

// // Import resource
// func (r resourceRule) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
// 	tfsdk_resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
// }

// // Import resource
func (r resourceRule) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
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
