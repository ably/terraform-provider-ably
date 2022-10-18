package ably_control

import (
	"context"

	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceKeyType struct{}

// Get Resource schema
func (r resourceKeyType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The key ID.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk_resource.UseStateForUnknown(),
				},
			},
			"app_id": {
				Type:        types.StringType,
				Required:    true,
				Description: "The Ably application ID which this key is associated with.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk_resource.RequiresReplace(),
				},
			},
			"name": {
				Type:        types.StringType,
				Required:    true,
				Description: "The name for your API key. This is a friendly name for your reference.",
			},
			"capabilities": {
				Type: types.MapType{
					ElemType: types.ListType{
						ElemType: types.StringType,
					},
				},
				Required:    true,
				Description: "The capabilities that this key has. More information on capabilities can be found in the [Ably documentation](https://ably.com/docs/core-features/authentication#capabilities-explained)",
			},
			"status": {
				Type:        types.Int64Type,
				Computed:    true,
				Description: "The status of the key. 0 is enabled, 1 is revoked.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.Int64{Value: 0}),
				},
			},
			"created": {
				Type:        types.Int64Type,
				Computed:    true,
				Description: "Enforce TLS for all connections. This setting overrides any channel setting.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk_resource.UseStateForUnknown(),
				},
			},
			"key": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The complete API key including API secret.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk_resource.UseStateForUnknown(),
				},
			},
			"modified": {
				Type:        types.Int64Type,
				Computed:    true,
				Description: "Unix timestamp representing the date and time of the last modification of the key.",
			},
		},
		MarkdownDescription: "The `ably_key` resource allows you to create and manage Ably API keys.",
	}, nil
}

// New resource instance
func (r resourceKeyType) NewResource(_ context.Context, p tfsdk_provider.Provider) (tfsdk_resource.Resource, diag.Diagnostics) {
	return resourceKey{
		p: *(p.(*provider)),
	}, nil
}

type resourceKey struct {
	p provider
}

// Create a new resource
func (r resourceKey) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	// Checks whether the provider and API Client are configured. If they are not, the provider responds with an error.
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply",
		)
		return
	}

	// Gets plan values
	var plan AblyKey
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	new_key := ably_control_go.NewKey{
		Name:       plan.Name.Value,
		Capability: plan.Capability,
	}

	// Creates a new Ably Key by invoking the CreateKey function from the Client Library
	ably_key, err := r.p.client.CreateKey(plan.AppID.Value, &new_key)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Maps response body to resource schema attributes.
	resp_key := AblyKey{
		ID:         types.String{Value: ably_key.ID},
		AppID:      types.String{Value: ably_key.AppID},
		Name:       types.String{Value: ably_key.Name},
		Key:        types.String{Value: ably_key.Key},
		Capability: ably_key.Capability,
		Status:     types.Int64{Value: int64(ably_key.Status)},
		Created:    types.Int64{Value: int64(ably_key.Created)},
		Modified:   types.Int64{Value: int64(ably_key.Modified)},
	}

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, resp_key)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource
func (r resourceKey) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var state AblyKey
	found := false
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID and Ably API Key ID value for the resource
	app_id := state.AppID.Value
	key_id := state.ID.Value

	// Fetches all Ably Keys for the Ably App. The function invokes the Client Library Keys() method.
	keys, err := r.p.client.Keys(app_id)
	if err != nil {
		if is_404(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Loops through apps and if account id and key id match, sets state.
	for _, v := range keys {
		if v.AppID == app_id && v.ID == key_id && v.Status == 0 {
			resp_key := AblyKey{
				ID:         types.String{Value: v.ID},
				AppID:      types.String{Value: v.AppID},
				Name:       types.String{Value: v.Name},
				Capability: v.Capability,
				Status:     types.Int64{Value: int64(v.Status)},
				Key:        types.String{Value: v.Key},
				Created:    types.Int64{Value: int64(v.Created)},
				Modified:   types.Int64{Value: int64(v.Modified)},
			}
			// Sets state to app values.
			diags = resp.State.Set(ctx, &resp_key)
			found = true

			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
	}

}

// Update resource
func (r resourceKey) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	// Get plan values
	var plan AblyKey
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AblyKey
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the app ID and Key ID
	app_id := plan.AppID.Value
	key_id := state.ID.Value

	// Instantiates struct of type ably_control_go.NewKey and sets values to output of plan
	key_values := ably_control_go.NewKey{
		Name:       plan.Name.Value,
		Capability: plan.Capability,
	}

	// Updates an Ably API Key. The function invokes the Client Library UpdateKey method.
	ably_key, err := r.p.client.UpdateKey(app_id, key_id, &key_values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Resource",
			"Could not update resource, unexpected error: "+err.Error(),
		)
		return
	}

	resp_key := AblyKey{
		ID:         types.String{Value: ably_key.ID},
		AppID:      types.String{Value: ably_key.AppID},
		Name:       types.String{Value: ably_key.Name},
		Capability: ably_key.Capability,
		Status:     types.Int64{Value: int64(ably_key.Status)},
		Key:        types.String{Value: ably_key.Key},
		Created:    types.Int64{Value: int64(ably_key.Created)},
		Modified:   types.Int64{Value: int64(ably_key.Modified)},
	}

	// Sets state.
	diags = resp.State.Set(ctx, resp_key)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceKey) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	// Get current state
	var state AblyKey
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the current state. If it is unable to, the provider responds with an error.
	app_id := state.AppID.Value
	key_id := state.ID.Value

	err := r.p.client.RevokeKey(app_id, key_id)
	if err != nil {
		if is_404(err) {
			resp.Diagnostics.AddWarning(
				"Resource does not exist",
				"Resource does not exist, it may have already been deleted: "+err.Error(),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error deleting Resource",
				"Could not delete resource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

// // Import resource
func (r resourceKey) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
