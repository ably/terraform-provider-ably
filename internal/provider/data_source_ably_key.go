// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/ably/terraform-provider-ably/internal/provider/codegen/datasource_keys"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AblyKeyDataSourceModel is one API key as the data sources report it. It serves
// both ably_api_key and the elements of ably_api_keys.
//
// The key attribute carries the complete API key including its secret, and is
// marked sensitive by the generator's shared sensitive-name set (see
// codegen/ruletypesgen). Anything reading this data source is reading a
// credential.
type AblyKeyDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	AppID           types.String `tfsdk:"app_id"`
	Name            types.String `tfsdk:"name"`
	Key             types.String `tfsdk:"key"`
	Status          types.Int64  `tfsdk:"status"`
	RevocableTokens types.Bool   `tfsdk:"revocable_tokens"`
	Capabilities    types.Map    `tfsdk:"capabilities"`
	Created         types.Int64  `tfsdk:"created"`
	Modified        types.Int64  `tfsdk:"modified"`
}

// AblyKeysDataSourceModel is the model for the plural ably_api_keys data source.
type AblyKeysDataSourceModel struct {
	AppID types.String             `tfsdk:"app_id"`
	Keys  []AblyKeyDataSourceModel `tfsdk:"keys"`
}

// keyDataSourceModel maps an API key onto the shared model.
func keyDataSourceModel(ctx context.Context, key control.KeyResponse) (AblyKeyDataSourceModel, diag.Diagnostics) {
	capability, diags := capabilityToMap(ctx, key.Capability)

	return AblyKeyDataSourceModel{
		ID:              types.StringValue(key.ID),
		AppID:           types.StringValue(key.AppID),
		Name:            types.StringValue(key.Name),
		Key:             types.StringValue(key.Key),
		Status:          types.Int64Value(int64(key.Status)),
		RevocableTokens: optBoolValue(key.RevocableTokens),
		Capabilities:    capability,
		Created:         types.Int64Value(key.Created),
		Modified:        types.Int64Value(key.Modified),
	}, diags
}

// capabilityToMap converts the API's capability map into the tfsdk map of lists,
// returning null when the key has no capabilities rather than an empty map.
func capabilityToMap(ctx context.Context, capability map[string][]string) (types.Map, diag.Diagnostics) {
	listType := types.ListType{ElemType: types.StringType}
	if len(capability) == 0 {
		return types.MapNull(listType), nil
	}
	return types.MapValueFrom(ctx, listType, capability)
}

// keysDataSourceSchema returns the generated ably_api_keys schema, ready to serve.
//
// The API calls the permissions map `capability` and the generated schema follows
// it, but the ably_api_key resource has always exposed it as `capabilities`. Matching
// the resource matters more here than matching the spec's property name, so the
// attribute is renamed. It stays a map of lists rather than the resource's map of
// sets: the data source is read-only, so preserving the API's ordering costs
// nothing and needs no sorting plan modifier.
func keysDataSourceSchema(ctx context.Context) schema.Schema {
	s := datasource_keys.KeysDataSourceSchema(ctx)
	stripSetNestedCustomTypes(&s, "keys")

	set := s.Attributes["keys"].(schema.SetNestedAttribute)
	if capability, ok := set.NestedObject.Attributes["capability"]; ok {
		set.NestedObject.Attributes["capabilities"] = capability
		delete(set.NestedObject.Attributes, "capability")
	}
	s.Attributes["keys"] = set

	return s
}

// --- ably_api_keys -------------------------------------------------------------

type DataSourceKeys struct {
	p *AblyProvider
}

var _ datasource.DataSource = &DataSourceKeys{}

func (d DataSourceKeys) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "ably_api_keys"
}

// Schema defines the schema for the data source.
//
// GENERATED SCHEMA: the attribute set, types, nesting, sensitivity and
// descriptions come from internal/provider/codegen, produced by `make generate`
// from the Control API spec.
func (d DataSourceKeys) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	s := keysDataSourceSchema(ctx)

	s.MarkdownDescription = "The `ably_api_keys` data source lists every API key in an Ably app, including keys Terraform does not manage. Each key includes its secret, so treat anything derived from this data source as a credential. Use `ably_api_key` to look up a single key by id or name."

	resp.Schema = s
}

func (d DataSourceKeys) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	var config AblyKeysDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keys, err := d.p.client.ListKeys(ctx, config.AppID.ValueString())
	if err != nil {
		readDataSourceError(&resp.Diagnostics, "ably_api_keys", err)
		return
	}

	state := AblyKeysDataSourceModel{
		AppID: config.AppID,
		Keys:  make([]AblyKeyDataSourceModel, 0, len(keys)),
	}
	for _, key := range keys {
		model, diags := keyDataSourceModel(ctx, key)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Keys = append(state.Keys, model)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// --- ably_api_key --------------------------------------------------------------

type DataSourceKey struct {
	p *AblyProvider
}

var _ datasource.DataSource = &DataSourceKey{}

func (d DataSourceKey) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "ably_api_key"
}

// Schema defines the schema for the data source. The attributes are lifted from
// the generated ably_api_keys element, so this schema tracks the spec too.
func (d DataSourceKey) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := elementAttributes(keysDataSourceSchema(ctx), "keys")

	requireString(attributes, "app_id", "The Ably app the key belongs to.")
	optionalString(attributes, "id", "The key ID. Set either this or name.")
	optionalString(attributes, "name", "The key name. Set either this or id. Names are not unique, so a name matching more than one key is an error.")

	resp.Schema = schema.Schema{
		Attributes:          attributes,
		MarkdownDescription: "The `ably_api_key` data source looks up a single Ably API key by id or name, including its secret. The Control API has no fetch-by-id endpoint for keys, so this lists the app's keys and matches locally.",
	}
}

func (d DataSourceKey) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	var config AblyKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keys, err := d.p.client.ListKeys(ctx, config.AppID.ValueString())
	if err != nil {
		readDataSourceError(&resp.Diagnostics, "ably_api_key", err)
		return
	}

	key, diags := findOne(
		lookup{dataSourceName: "ably_api_key", id: config.ID, name: config.Name},
		keys,
		func(k control.KeyResponse) string { return k.ID },
		func(k control.KeyResponse) string { return k.Name },
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, modelDiags := keyDataSourceModel(ctx, key)
	resp.Diagnostics.Append(modelDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
