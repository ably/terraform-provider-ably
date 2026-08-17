// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/ably/terraform-provider-ably/internal/provider/codegen/datasource_apps"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AblyAppDataSourceModel is one app as the data sources report it.
//
// It mirrors the generated element attributes of the ably_apps collection, and
// serves as the model for both data sources: the whole point of lifting the
// element attributes for the singular schema (see data_sources.go) is that one
// struct fits both.
//
// Note the timestamps are Unix milliseconds here, matching the API and the
// generated schema, where the ably_app resource carries RFC3339 strings for
// backwards compatibility with existing state.
type AblyAppDataSourceModel struct {
	ID                          types.String `tfsdk:"id"`
	AccountID                   types.String `tfsdk:"account_id"`
	Name                        types.String `tfsdk:"name"`
	Status                      types.String `tfsdk:"status"`
	TLSOnly                     types.Bool   `tfsdk:"tls_only"`
	APNSUseSandboxEndpoint      types.Bool   `tfsdk:"apns_use_sandbox_endpoint"`
	APNSAuthType                types.String `tfsdk:"apns_auth_type"`
	APNSIssuerKey               types.String `tfsdk:"apns_issuer_key"`
	APNSSigningKeyID            types.String `tfsdk:"apns_signing_key_id"`
	APNSTopicHeader             types.String `tfsdk:"apns_topic_header"`
	APNSCertificateConfigured   types.Bool   `tfsdk:"apns_certificate_configured"`
	APNSSigningKeyConfigured    types.Bool   `tfsdk:"apns_signing_key_configured"`
	FCMProjectID                types.String `tfsdk:"fcm_project_id"`
	FCMServiceAccountConfigured types.Bool   `tfsdk:"fcm_service_account_configured"`
	Created                     types.Int64  `tfsdk:"created"`
	Modified                    types.Int64  `tfsdk:"modified"`
}

// AblyAppsDataSourceModel is the model for the plural ably_apps data source.
type AblyAppsDataSourceModel struct {
	AccountID types.String             `tfsdk:"account_id"`
	Apps      []AblyAppDataSourceModel `tfsdk:"apps"`
}

// appDataSourceModel maps an API app onto the shared model.
func appDataSourceModel(app control.AppResponse) AblyAppDataSourceModel {
	return AblyAppDataSourceModel{
		ID:                          types.StringValue(app.ID),
		AccountID:                   types.StringValue(app.AccountID),
		Name:                        types.StringValue(app.Name),
		Status:                      types.StringValue(app.Status),
		TLSOnly:                     optBoolValue(app.TLSOnly),
		APNSUseSandboxEndpoint:      optBoolValue(app.APNSUseSandboxEndpoint),
		APNSAuthType:                optStringValue(app.APNSAuthType),
		APNSIssuerKey:               optStringValue(app.APNSIssuerKey),
		APNSSigningKeyID:            optStringValue(app.APNSSigningKeyID),
		APNSTopicHeader:             optStringValue(app.APNSTopicHeader),
		APNSCertificateConfigured:   optBoolValue(app.APNSCertificateConfigured),
		APNSSigningKeyConfigured:    optBoolValue(app.APNSSigningKeyConfigured),
		FCMProjectID:                optStringValue(app.FCMProjectID),
		FCMServiceAccountConfigured: optBoolValue(app.FCMServiceAccountConfigured),
		Created:                     types.Int64Value(app.Created),
		Modified:                    types.Int64Value(app.Modified),
	}
}

// appsDataSourceSchema returns the generated ably_apps schema, ready to serve.
//
// The generated schema models the API's self link as a typed nested object. It is
// no use to anyone in Terraform, so both app data sources drop it rather than
// carry it in their models.
//
// TODO(INF-7992): delete this strip once the spec marks _links opaque or omits
// it.
func appsDataSourceSchema(ctx context.Context) schema.Schema {
	s := datasource_apps.AppsDataSourceSchema(ctx)
	stripSetNestedCustomTypes(&s, "apps")

	set := s.Attributes["apps"].(schema.SetNestedAttribute)
	delete(set.NestedObject.Attributes, "_links")
	s.Attributes["apps"] = set

	return s
}

// --- ably_apps -------------------------------------------------------------

type DataSourceApps struct {
	p *AblyProvider
}

var _ datasource.DataSource = &DataSourceApps{}

func (d DataSourceApps) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "ably_apps"
}

// Schema defines the schema for the data source.
//
// GENERATED SCHEMA: the attribute set, types, nesting, sensitivity and
// descriptions come from internal/provider/codegen, produced by `make generate`
// from the Control API spec.
func (d DataSourceApps) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	s := appsDataSourceSchema(ctx)

	// The provider already knows the account, so asking again is noise.
	optionalString(s.Attributes, "account_id",
		"The account to list apps for. Defaults to the account the provider's token belongs to.")

	s.MarkdownDescription = "The `ably_apps` data source lists every Ably app in an account, including apps Terraform does not manage. Use `ably_app` to look up a single app by id or name."

	resp.Schema = s
}

func (d DataSourceApps) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	var config AblyAppsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID := dataSourceAccountID(d.p, config.AccountID)
	apps, err := d.p.client.ListApps(ctx, accountID)
	if err != nil {
		readDataSourceError(&resp.Diagnostics, "ably_apps", err)
		return
	}

	state := AblyAppsDataSourceModel{
		AccountID: types.StringValue(accountID),
		Apps:      make([]AblyAppDataSourceModel, 0, len(apps)),
	}
	for _, app := range apps {
		state.Apps = append(state.Apps, appDataSourceModel(app))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// --- ably_app --------------------------------------------------------------

type DataSourceApp struct {
	p *AblyProvider
}

var _ datasource.DataSource = &DataSourceApp{}

func (d DataSourceApp) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "ably_app"
}

// Schema defines the schema for the data source. The attributes are lifted from
// the generated ably_apps element, so this schema tracks the spec too.
func (d DataSourceApp) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := elementAttributes(appsDataSourceSchema(ctx), "apps")

	optionalString(attributes, "id", "The app ID. Set either this or name.")
	optionalString(attributes, "name", "The app name. Set either this or id. Names are not unique, so a name matching more than one app is an error.")
	optionalString(attributes, "account_id", "The account to search. Defaults to the account the provider's token belongs to.")

	resp.Schema = schema.Schema{
		Attributes:          attributes,
		MarkdownDescription: "The `ably_app` data source looks up a single Ably app by id or name, so you can reference an app this Terraform configuration does not manage. The Control API has no fetch-by-id endpoint for apps, so this lists the account's apps and matches locally.",
	}
}

func (d DataSourceApp) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	var config AblyAppDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID := dataSourceAccountID(d.p, config.AccountID)
	apps, err := d.p.client.ListApps(ctx, accountID)
	if err != nil {
		readDataSourceError(&resp.Diagnostics, "ably_app", err)
		return
	}

	app, diags := findOne(
		lookup{dataSourceName: "ably_app", id: config.ID, name: config.Name},
		apps,
		func(a control.AppResponse) string { return a.ID },
		func(a control.AppResponse) string { return a.Name },
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := appDataSourceModel(app)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
