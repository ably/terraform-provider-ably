// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/ably/terraform-provider-ably/internal/provider/codegen/datasource_me"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AblyMeTokenModel mirrors control.MeToken.
type AblyMeTokenModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Capabilities types.List   `tfsdk:"capabilities"`
	ExpiresAt    types.String `tfsdk:"expires_at"`
	LastUsedAt   types.String `tfsdk:"last_used_at"`
}

// AblyMeUserModel mirrors control.MeUser.
type AblyMeUserModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
}

// AblyMeAccountModel mirrors control.MeAccount.
type AblyMeAccountModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// AblyMeDataSourceModel is the model for the ably_me data source.
type AblyMeDataSourceModel struct {
	Token   *AblyMeTokenModel   `tfsdk:"token"`
	User    *AblyMeUserModel    `tfsdk:"user"`
	Account *AblyMeAccountModel `tfsdk:"account"`
}

type DataSourceMe struct {
	p *AblyProvider
}

var _ datasource.DataSource = &DataSourceMe{}

func (d DataSourceMe) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "ably_me"
}

// Schema defines the schema for the data source.
//
// GENERATED SCHEMA: everything comes from internal/provider/codegen, produced by
// `make generate` from the Control API spec.
func (d DataSourceMe) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	s := datasource_me.MeDataSourceSchema(ctx)
	stripDataSourceCustomTypes(s.Attributes)

	s.MarkdownDescription = "The `ably_me` data source describes the token the provider is configured with, and the account and user it belongs to. It is the way to reference your account ID without hardcoding it."

	resp.Schema = s
}

func (d DataSourceMe) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	me, err := d.p.client.Me(ctx)
	if err != nil {
		readDataSourceError(&resp.Diagnostics, "ably_me", err)
		return
	}

	var state AblyMeDataSourceModel

	// Every block is nullable: which of them come back depends on the token type
	// and its capabilities, so a missing block is normal rather than an error.
	if me.Token != nil {
		capabilities, diags := stringsToList(ctx, me.Token.Capabilities)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Token = &AblyMeTokenModel{
			ID:           types.StringValue(me.Token.ID),
			Name:         types.StringValue(me.Token.Name),
			Capabilities: capabilities,
			ExpiresAt:    optStringValue(me.Token.ExpiresAt),
			LastUsedAt:   optStringValue(me.Token.LastUsedAt),
		}
	}
	if me.User != nil {
		state.User = &AblyMeUserModel{
			ID:    types.Int64Value(int64(me.User.ID)),
			Email: types.StringValue(me.User.Email),
		}
	}
	if me.Account != nil {
		state.Account = &AblyMeAccountModel{
			ID:   types.StringValue(me.Account.ID),
			Name: types.StringValue(me.Account.Name),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
