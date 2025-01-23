package provider

import (
	"context"

	control "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceKey struct {
	p *provider
}

// Get Resource schema
func (r resourceKey) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
					ElemType: types.SetType{
						ElemType: types.StringType,
					},
				},
				Required:    true,
				Description: "The capabilities that this key has. More information on capabilities can be found in the [Ably documentation](https://ably.com/docs/core-features/authentication#capabilities-explained)",
			},
			"revocable_tokens": {
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				Description: "Allow tokens issued by this key to be revoked. More information on Token Revocation can be found in the [Ably documentation](https://ably.com/docs/auth/revocation)",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.BoolValue(false)),
				},
			},
			"status": {
				Type:        types.Int64Type,
				Computed:    true,
				Description: "The status of the key. 0 is enabled, 1 is revoked.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.Int64Value(0)),
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

func (r resourceKey) Metadata(ctx context.Context, req tfsdk_resource.MetadataRequest, resp *tfsdk_resource.MetadataResponse) {
	resp.TypeName = "ably_api_key"
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

	newKey := control.NewKey{
		Name:            plan.Name.ValueString(),
		Capability:      plan.Capability,
		RevocableTokens: plan.RevocableTokens.ValueBool(),
	}

	// Creates a new Ably Key by invoking the CreateKey function from the Client Library
	key, err := r.p.client.CreateKey(plan.AppID.ValueString(), &newKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Maps response body to resource schema attributes.
	respKey := AblyKey{
		ID:              types.StringValue(key.ID),
		AppID:           types.StringValue(key.AppID),
		Name:            types.StringValue(key.Name),
		Key:             types.StringValue(key.Key),
		RevocableTokens: types.BoolValue(key.RevocableTokens),
		Capability:      key.Capability,
		Status:          types.Int64Value(int64(key.Status)),
		Created:         types.Int64Value(int64(key.Created)),
		Modified:        types.Int64Value(int64(key.Modified)),
	}

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, respKey)
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
	appID := state.AppID.ValueString()
	keyID := state.ID.ValueString()

	// Fetches all Ably Keys for the Ably App. The function invokes the Client Library Keys() method.
	keys, err := r.p.client.Keys(appID)
	if err != nil {
		if is404(err) {
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
		if v.AppID == appID && v.ID == keyID && v.Status == 0 {
			respKey := AblyKey{
				ID:              types.StringValue(v.ID),
				AppID:           types.StringValue(v.AppID),
				Name:            types.StringValue(v.Name),
				RevocableTokens: types.BoolValue(v.RevocableTokens),
				Capability:      v.Capability,
				Status:          types.Int64Value(int64(v.Status)),
				Key:             types.StringValue(v.Key),
				Created:         types.Int64Value(int64(v.Created)),
				Modified:        types.Int64Value(int64(v.Modified)),
			}
			// Sets state to app values.
			diags = resp.State.Set(ctx, &respKey)
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
	appID := plan.AppID.ValueString()
	keyID := state.ID.ValueString()

	// Instantiates struct of type control.NewKey and sets values to output of plan
	keyValues := control.NewKey{
		Name:            plan.Name.ValueString(),
		Capability:      plan.Capability,
		RevocableTokens: plan.RevocableTokens.ValueBool(),
	}

	// Updates an Ably API Key. The function invokes the Client Library UpdateKey method.
	key, err := r.p.client.UpdateKey(appID, keyID, &keyValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Resource",
			"Could not update resource, unexpected error: "+err.Error(),
		)
		return
	}

	respKey := AblyKey{
		ID:              types.StringValue(key.ID),
		AppID:           types.StringValue(key.AppID),
		Name:            types.StringValue(key.Name),
		RevocableTokens: types.BoolValue(key.RevocableTokens),
		Capability:      key.Capability,
		Status:          types.Int64Value(int64(key.Status)),
		Key:             types.StringValue(key.Key),
		Created:         types.Int64Value(int64(key.Created)),
		Modified:        types.Int64Value(int64(key.Modified)),
	}

	// Sets state.
	diags = resp.State.Set(ctx, respKey)
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
	appID := state.AppID.ValueString()
	keyID := state.ID.ValueString()

	err := r.p.client.RevokeKey(appID, keyID)
	if err != nil {
		if is404(err) {
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
