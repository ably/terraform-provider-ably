// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
				ElementType: types.SetType{
					ElemType: types.StringType,
				},
				Required:    true,
				Description: "The capabilities that this key has. More information on capabilities can be found in the [Ably documentation](https://ably.com/docs/core-features/authentication#capabilities-explained)",
				PlanModifiers: []planmodifier.Map{
					SortSetsInMap(),
				},
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
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.Int64Attribute{
				Computed:    true,
				Description: "The timestamp of when the key was created.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
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

// buildKeyState reconciles plan/state input with an API response.
func buildKeyState(rc *reconciler, input AblyKey, api control.KeyResponse) AblyKey {
	return AblyKey{
		ID:              rc.str("id", input.ID, types.StringValue(api.ID), true),
		AppID:           rc.str("app_id", input.AppID, types.StringValue(api.AppID), false),
		Name:            rc.str("name", input.Name, types.StringValue(api.Name), false),
		Key:             rc.str("key", input.Key, types.StringValue(api.Key), true),
		RevocableTokens: rc.boolean("revocable_tokens", input.RevocableTokens, optBoolValue(api.RevocableTokens), true),
		Capability:      rc.mapSet("capabilities", input.Capability, mapToTypedSet(api.Capability), false),
		Status:          rc.int64val("status", input.Status, types.Int64Value(int64(api.Status)), true),
		Created:         rc.int64val("created", input.Created, types.Int64Value(api.Created), true),
		Modified:        rc.int64val("modified", input.Modified, types.Int64Value(api.Modified), true),
	}
}

// Create creates a new resource.
func (r ResourceKey) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.ensureConfigured(&resp.Diagnostics) {
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
	capability := mapFromSet(ctx, plan.Capability)

	revocable := plan.RevocableTokens.ValueBool()
	newKey := control.KeyPost{
		Name:            plan.Name.ValueString(),
		Capability:      capability,
		RevocableTokens: &revocable,
	}

	// Creates a new Ably Key by invoking the CreateKey function from the Client Library
	ablyKey, err := r.p.client.CreateKey(ctx, plan.AppID.ValueString(), newKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ably_api_key",
			"Could not create ably_api_key, unexpected error: "+err.Error(),
		)
		return
	}

	// Read back the resource via GET to ensure computed fields like `modified`
	// reflect the settled server state (the POST response may return a value
	// that the server updates asynchronously).
	appID := plan.AppID.ValueString()
	keys, err := r.p.client.ListKeys(ctx, appID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading back ably_api_key after create",
			"Could not read back ably_api_key, unexpected error: "+err.Error(),
		)
		return
	}
	for _, k := range keys {
		if k.ID == ablyKey.ID {
			ablyKey = k
			break
		}
	}

	// Maps response body to resource schema attributes.
	rc := newReconciler(&resp.Diagnostics)
	respKey := buildKeyState(rc, plan, ablyKey)
	if resp.Diagnostics.HasError() {
		return
	}

	// Sets state for the new Ably key.
	diags = resp.State.Set(ctx, respKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read reads the resource.
func (r ResourceKey) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

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

	// Fetches all Ably Keys for the Ably App. The function invokes the Client Library ListKeys() method.
	keys, err := r.p.client.ListKeys(ctx, appID)
	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading ably_api_key",
			"Could not read ably_api_key, unexpected error: "+err.Error(),
		)
		return
	}

	// Loops through apps and if account id and key id match, sets state.
	for _, v := range keys {
		if v.AppID == appID && v.ID == keyID && v.Status == 0 {
			rc := newReconciler(&resp.Diagnostics).forRead()
			respKey := buildKeyState(rc, state, v)
			if resp.Diagnostics.HasError() {
				return
			}

			// Sets state to key values.
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
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

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

	// Instantiates struct of type control.KeyPatch and sets values to output of plan
	updateRevocable := plan.RevocableTokens.ValueBool()
	keyValues := control.KeyPatch{
		Name:            plan.Name.ValueString(),
		Capability:      mapFromSet(ctx, plan.Capability),
		RevocableTokens: &updateRevocable,
	}

	// Updates an Ably API Key. The function invokes the Client Library UpdateKey method.
	ablyKey, err := r.p.client.UpdateKey(ctx, appID, keyID, keyValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ably_api_key",
			"Could not update ably_api_key, unexpected error: "+err.Error(),
		)
		return
	}

	// Read back via GET to get settled computed fields.
	keys, err := r.p.client.ListKeys(ctx, appID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading back ably_api_key after update",
			"Could not read back ably_api_key, unexpected error: "+err.Error(),
		)
		return
	}
	for _, k := range keys {
		if k.ID == ablyKey.ID {
			ablyKey = k
			break
		}
	}

	rc := newReconciler(&resp.Diagnostics)
	respKey := buildKeyState(rc, plan, ablyKey)
	if resp.Diagnostics.HasError() {
		return
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
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

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

	err := r.p.client.RevokeKey(ctx, appID, keyID)
	if err != nil {
		if is404(err) {
			resp.Diagnostics.AddWarning(
				"Resource does not exist",
				"Resource does not exist, it may have already been deleted: "+err.Error(),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error deleting ably_api_key",
				"Could not delete ably_api_key, unexpected error: "+err.Error(),
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
