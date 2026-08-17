// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/ably/terraform-provider-ably/internal/provider/codegen/datasource_namespaces"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AblyNamespaceDataSourceModel is one namespace (channel rule) as the data
// sources report it. It serves both ably_namespace and the elements of
// ably_namespaces.
type AblyNamespaceDataSourceModel struct {
	ID                      types.String `tfsdk:"id"`
	AppID                   types.String `tfsdk:"app_id"`
	Identified              types.Bool   `tfsdk:"identified"`
	Authenticated           types.Bool   `tfsdk:"authenticated"`
	Persisted               types.Bool   `tfsdk:"persisted"`
	PersistLast             types.Bool   `tfsdk:"persist_last"`
	PushEnabled             types.Bool   `tfsdk:"push_enabled"`
	TLSOnly                 types.Bool   `tfsdk:"tls_only"`
	ExposeTimeserial        types.Bool   `tfsdk:"expose_timeserial"`
	MutableMessages         types.Bool   `tfsdk:"mutable_messages"`
	PopulateChannelRegistry types.Bool   `tfsdk:"populate_channel_registry"`
	BatchingEnabled         types.Bool   `tfsdk:"batching_enabled"`
	BatchingInterval        types.Int64  `tfsdk:"batching_interval"`
	ConflationEnabled       types.Bool   `tfsdk:"conflation_enabled"`
	ConflationInterval      types.Int64  `tfsdk:"conflation_interval"`
	ConflationKey           types.String `tfsdk:"conflation_key"`
	Created                 types.Int64  `tfsdk:"created"`
	Modified                types.Int64  `tfsdk:"modified"`
}

// AblyNamespacesDataSourceModel is the model for the plural ably_namespaces data
// source.
type AblyNamespacesDataSourceModel struct {
	AppID      types.String                   `tfsdk:"app_id"`
	Namespaces []AblyNamespaceDataSourceModel `tfsdk:"namespaces"`
}

// namespaceDataSourceModel maps an API namespace onto the shared model.
func namespaceDataSourceModel(namespace control.NamespaceResponse) AblyNamespaceDataSourceModel {
	return AblyNamespaceDataSourceModel{
		ID:    types.StringValue(namespace.ID),
		AppID: types.StringValue(namespace.AppID),
		// identified and authenticated are two names for one platform setting, so
		// both report the same effective value rather than whichever field this
		// API deployment happened to populate. Reporting authenticated as false
		// next to identified as true would be a lie about the namespace.
		Identified:              types.BoolValue(namespaceIdentifiedValue(namespace)),
		Authenticated:           types.BoolValue(namespaceIdentifiedValue(namespace)),
		Persisted:               types.BoolValue(namespace.Persisted),
		PersistLast:             types.BoolValue(namespace.PersistLast),
		PushEnabled:             types.BoolValue(namespace.PushEnabled),
		TLSOnly:                 types.BoolValue(namespace.TLSOnly),
		ExposeTimeserial:        types.BoolValue(namespace.ExposeTimeserial),
		MutableMessages:         types.BoolValue(namespace.MutableMessages),
		PopulateChannelRegistry: types.BoolValue(namespace.PopulateChannelRegistry),
		BatchingEnabled:         optBoolValue(namespace.BatchingEnabled),
		BatchingInterval:        optIntValue(namespace.BatchingInterval),
		ConflationEnabled:       optBoolValue(namespace.ConflationEnabled),
		ConflationInterval:      optIntValue(namespace.ConflationInterval),
		ConflationKey:           optStringValue(namespace.ConflationKey),
		Created:                 types.Int64Value(namespace.Created),
		Modified:                types.Int64Value(namespace.Modified),
	}
}

// namespacesDataSourceSchema returns the generated ably_namespaces schema, ready
// to serve.
//
// The vendored spec documents only the deprecated authenticated flag, not the
// canonical identified one the API now returns, so identified is added here. When
// ably/docs catches up this can come from the generated schema and the addition
// can go; the spec-drift workflow is what will tell us it has.
func namespacesDataSourceSchema(ctx context.Context) schema.Schema {
	s := datasource_namespaces.NamespacesDataSourceSchema(ctx)
	stripSetNestedCustomTypes(&s, "namespaces")

	set := s.Attributes["namespaces"].(schema.SetNestedAttribute)
	set.NestedObject.Attributes["identified"] = schema.BoolAttribute{
		Computed:            true,
		Description:         identifiedDescription,
		MarkdownDescription: identifiedDescription,
	}
	if authenticated, ok := set.NestedObject.Attributes["authenticated"].(schema.BoolAttribute); ok {
		authenticated.DeprecationMessage = "Use identified instead; authenticated is the deprecated alias for the same flag."
		set.NestedObject.Attributes["authenticated"] = authenticated
	}
	s.Attributes["namespaces"] = set

	return s
}

const identifiedDescription = "If `true`, clients are not permitted to use any channel in this namespace unless they are identified, that is, authenticated with a client ID. This is the canonical name for the flag the API also reports as `authenticated`."

// --- ably_namespaces -------------------------------------------------------

type DataSourceNamespaces struct {
	p *AblyProvider
}

var _ datasource.DataSource = &DataSourceNamespaces{}

func (d DataSourceNamespaces) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "ably_namespaces"
}

// Schema defines the schema for the data source.
//
// GENERATED SCHEMA: the attribute set, types, nesting and descriptions come from
// internal/provider/codegen, produced by `make generate` from the Control API
// spec, plus the identified flag the spec does not carry yet.
func (d DataSourceNamespaces) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	s := namespacesDataSourceSchema(ctx)

	s.MarkdownDescription = "The `ably_namespaces` data source lists every namespace (channel rule) in an Ably app, including namespaces Terraform does not manage. Use `ably_namespace` to look up a single namespace by id."

	resp.Schema = s
}

func (d DataSourceNamespaces) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	var config AblyNamespacesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespaces, err := d.p.client.ListNamespaces(ctx, config.AppID.ValueString())
	if err != nil {
		readDataSourceError(&resp.Diagnostics, "ably_namespaces", err)
		return
	}

	state := AblyNamespacesDataSourceModel{
		AppID:      config.AppID,
		Namespaces: make([]AblyNamespaceDataSourceModel, 0, len(namespaces)),
	}
	for _, namespace := range namespaces {
		state.Namespaces = append(state.Namespaces, namespaceDataSourceModel(namespace))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// --- ably_namespace --------------------------------------------------------

type DataSourceNamespace struct {
	p *AblyProvider
}

var _ datasource.DataSource = &DataSourceNamespace{}

func (d DataSourceNamespace) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "ably_namespace"
}

// Schema defines the schema for the data source. The attributes are lifted from
// the generated ably_namespaces element, so this schema tracks the spec too.
//
// Namespaces have no name: the id (the channel prefix, e.g. "chat") is the name,
// so this looks up by id only.
func (d DataSourceNamespace) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := elementAttributes(namespacesDataSourceSchema(ctx), "namespaces")

	requireString(attributes, "app_id", "The Ably app the namespace belongs to.")
	requireString(attributes, "id", "The namespace ID, which is the channel name prefix (for example `chat`).")

	resp.Schema = schema.Schema{
		Attributes:          attributes,
		MarkdownDescription: "The `ably_namespace` data source looks up a single Ably namespace (channel rule) by id, so you can read the settings of a namespace this Terraform configuration does not manage. The Control API has no fetch-by-id endpoint for namespaces, so this lists the app's namespaces and matches locally.",
	}
}

func (d DataSourceNamespace) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	var config AblyNamespaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespaces, err := d.p.client.ListNamespaces(ctx, config.AppID.ValueString())
	if err != nil {
		readDataSourceError(&resp.Diagnostics, "ably_namespace", err)
		return
	}

	namespace, diags := findOne(
		lookup{dataSourceName: "ably_namespace", id: config.ID},
		namespaces,
		func(n control.NamespaceResponse) string { return n.ID },
		func(n control.NamespaceResponse) string { return n.ID },
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := namespaceDataSourceModel(namespace)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
