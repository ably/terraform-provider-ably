package ably_control

import (
	"context"

	"fmt"
	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
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
			},
			"app_id": {
				Type:        types.StringType,
				Required:    true,
				Description: "The Ably application ID which this key is associated with.",
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
				Description: "The capabilities that this key has. More information on capabilities can be found in the Ably documentation.",
			},
			"status": {
				Type:        types.Int64Type,
				Computed:    true,
				Description: "The status of the key. 0 is enabled, 1 is revoked.",
			},
			"created": {
				Type:        types.Int64Type,
				Computed:    true,
				Description: "Enforce TLS for all connections. This setting overrides any channel setting.",
			},
			"key": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The complete API key including API secret.",
			},
			"modified": {
				Type:        types.Int64Type,
				Computed:    true,
				Description: "Unix timestamp representing the date and time of the last modification of the key.",
			},
		},
	}, nil
}

// New resource instance
func (r resourceKeyType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceKey{
		p: *(p.(*provider)),
	}, nil
}

type resourceKey struct {
	p provider
}

// Create a new resource
func (r resourceKey) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
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
func (r resourceKey) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var state AblyKey
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID and Ably API Key ID value for the resource
	app_id := state.AppID.Value
	key_id := state.ID.Value

	// Fetches all Ably Keys for the Ably App. The function invokes the Client Library Keys() method.
	keys, _ := r.p.client.Keys(app_id)

	// Loops through apps and if account id and key id match, sets state.
	for _, v := range keys {
		if v.AppID == app_id && v.ID == key_id {
			resp_key := AblyKey{
				ID:         types.String{Value: v.ID},
				AppID:      types.String{Value: v.AppID},
				Name:       types.String{Value: v.Name},
				Capability: v.Capability,
				Status:     types.Int64{Value: int64(v.Status)},
				Created:    types.Int64{Value: int64(v.Created)},
				Modified:   types.Int64{Value: int64(v.Modified)},
			}
			// Sets state to app values.
			diags = resp.State.Set(ctx, &resp_key)

			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}
}

// Update resource
func (r resourceKey) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan values
	var plan AblyKey
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state AblyKey
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the app ID and Key ID
	app_id := state.AppID.Value
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
func (r resourceKey) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
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
func (r resourceKey) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	// identifier should be in the format app_id,key_id
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: 'app_id,key_id'. Got: %q", req.ID),
		)
		return
	}
	// Recent PR in TF Plugin Framework for paths but Hashicorp examples not updated - https://github.com/hashicorp/terraform-plugin-framework/pull/390
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
