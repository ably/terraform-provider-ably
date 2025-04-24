// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	control "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ResourceKey{}
var _ resource.ResourceWithImportState = &ResourceKey{}

type ResourceKey struct {
	p *AblyProvider
}

// Schema defines the schema for the resource.
func (r *ResourceKey) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The key ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_id": schema.StringAttribute{
				Required:    true,
				Description: "The Ably application ID which this key is associated with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name for your API key. This is a friendly name for your reference.",
			},
			"capabilities": schema.MapAttribute{
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
				Required:    true,
				Description: "The capabilities that this key has. More information on capabilities can be found in the [Ably documentation](https://ably.com/docs/core-features/authentication#capabilities-explained)",
			},
			"revocable_tokens": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Allow tokens issued by this key to be revoked. More information on Token Revocation can be found in the [Ably documentation](https://ably.com/docs/auth/revocation)",
				PlanModifiers: []planmodifier.Bool{
					DefaultBoolAttribute(types.BoolValue(false)),
				},
			},
			"status": schema.Int64Attribute{
				Computed:    true,
				Description: "The status of the key. 0 is enabled, 1 is revoked.",
				PlanModifiers: []planmodifier.Int64{
					DefaultInt64Attribute(types.Int64Value(0)),
				},
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "Enforce TLS for all connections. This setting overrides any channel setting.",
				PlanModifiers: []planmodifier.Int64{
					DefaultInt64Attribute(types.Int64Value(0)),
				},
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Description: "The complete API key including API secret.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp representing the date and time of the last modification of the key.",
			},
		},
		MarkdownDescription: "The `ably_key` resource allows you to create and manage Ably API keys.",
	}
}

func (r ResourceKey) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_api_key"
}

// Create creates a new resource.
func (r ResourceKey) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

	// Convert capability map from Terraform types to Go strings
	capability := mapFromStringSlice(plan.Capability)

	newKey := control.NewKey{
		Name:            plan.Name.ValueString(),
		Capability:      capability,
		RevocableTokens: plan.RevocableTokens.ValueBool(),
	}

	// Creates a new Ably Key by invoking the CreateKey function from the Client Library
	ablyKey, err := r.p.client.CreateKey(plan.AppID.ValueString(), &newKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Maps response body to resource schema attributes.
	// Convert capability map from Go strings to Terraform types
	tfCapability := mapToTypedStringSlice(ablyKey.Capability)

	respKey := AblyKey{
		ID:              types.StringValue(ablyKey.ID),
		AppID:           types.StringValue(ablyKey.AppID),
		Name:            types.StringValue(ablyKey.Name),
		Key:             types.StringValue(ablyKey.Key),
		RevocableTokens: types.BoolValue(ablyKey.RevocableTokens),
		Capability:      tfCapability,
		Status:          types.Int64Value(int64(ablyKey.Status)),
		Created:         types.Int64Value(int64(ablyKey.Created)),
		Modified:        types.Int64Value(int64(ablyKey.Modified)),
	}

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, respKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read reads the resource.
func (r ResourceKey) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
			// Convert capability map from Go strings to Terraform types
			tfCapability := mapToTypedStringSlice(v.Capability)

			respKey := AblyKey{
				ID:              types.StringValue(v.ID),
				AppID:           types.StringValue(v.AppID),
				Name:            types.StringValue(v.Name),
				RevocableTokens: types.BoolValue(v.RevocableTokens),
				Capability:      tfCapability,
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

// Update updates an existing resource.
func (r ResourceKey) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
		Capability:      mapFromStringSlice(plan.Capability),
		RevocableTokens: plan.RevocableTokens.ValueBool(),
	}

	// Updates an Ably API Key. The function invokes the Client Library UpdateKey method.
	ablyKey, err := r.p.client.UpdateKey(appID, keyID, &keyValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Resource",
			"Could not update resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert capability map from Go strings to Terraform types
	tfCapability := mapToTypedStringSlice(ablyKey.Capability)

	respKey := AblyKey{
		ID:              types.StringValue(ablyKey.ID),
		AppID:           types.StringValue(ablyKey.AppID),
		Name:            types.StringValue(ablyKey.Name),
		RevocableTokens: types.BoolValue(ablyKey.RevocableTokens),
		Capability:      tfCapability,
		Status:          types.Int64Value(int64(ablyKey.Status)),
		Key:             types.StringValue(ablyKey.Key),
		Created:         types.Int64Value(int64(ablyKey.Created)),
		Modified:        types.Int64Value(int64(ablyKey.Modified)),
	}

	// Sets state.
	diags = resp.State.Set(ctx, respKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource.
func (r ResourceKey) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

// ImportState handles the import state functionality.
func (r ResourceKey) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
