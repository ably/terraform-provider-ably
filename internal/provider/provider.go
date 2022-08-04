package ably_control

import (
	"context"
	"os"

	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured bool
	client     *ably_control_go.Client
}

// GetSchema
func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"token": {
				Type:     types.StringType,
				Sensitive: true,
				Required: true,
			},
			"url": {
				Type:      types.StringType,
				Required: true,
			},
		},
	}, nil
}

// Provider schema struct
type providerData struct {
	Token types.String `tfsdk:"token"`
	Url types.String `tfsdk:"url"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	// Retrieve provider data from configuration
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// User must provide a Ably token to the provider
	var token string
	if config.Token.Unknown {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Ably Token required",
		)
		return
	}

	if config.Token.Null {
		token = os.Getenv("ABLY_ACCOUNT_TOKEN")
	} else {
		token = config.Token.Value
	}

	if token == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find Ably token",
			"Username cannot be an empty string",
		)
		return
	}

	// User must specify an Ably Control API Url
	var url string
	if config.Url.Unknown {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Cannot use unknown value as Urlt",
		)
		return
	}

	if config.Url.Null {
		url = os.Getenv("ABLY_URL")
	} else {
		url = config.Url.Value
	}

	if url == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find url",
			"Url cannot be an empty string",
		)
		return
	}

	// Create a new Ably client and set it to the provider client
	c, _, err := ably_control_go.NewClientWithURL(token, url)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Unable to create Ably client:\n\n"+err.Error(),
		)
		return
	}
	
	p.client = &c
	p.configured = true
}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
	}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}
