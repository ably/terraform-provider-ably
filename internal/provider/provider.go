package ably_control

import (
	"context"
	"os"

	ably_control_go "github.com/ably/ably-control-go"
	tfsdk_datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const CONTROL_API_DEFAULT_URL = "https://control.ably.net/v1"

func New(version string) tfsdk_provider.Provider {
	return &AblyProvider{
		version: version,
	}
}

type AblyProvider struct {
	configured bool
	client     *ably_control_go.Client
	version    string
}

// GetSchema
func (p *AblyProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"token": {
				Type:      types.StringType,
				Sensitive: true,
				Optional:  true,
			},
			"url": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

// Provider schema struct
type AblyProviderData struct {
	Token types.String `tfsdk:"token"`
	Url   types.String `tfsdk:"url"`
}

func (p *AblyProvider) Configure(ctx context.Context, req tfsdk_provider.ConfigureRequest, resp *tfsdk_provider.ConfigureResponse) {
	// Retrieve AblyProvider data from configuration
	var config AblyProviderData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// User must provide a Ably token to the provider
	var token string
	if config.Token.IsUnknown() {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Ably API Token required",
		)
		return
	}

	if config.Token.IsNull() {
		token = os.Getenv("ABLY_ACCOUNT_TOKEN")
	} else {
		token = config.Token.ValueString()
	}

	if token == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find Ably API token",
			"Ably API token cannot be an empty string. Ensure the providers token parameter is configured",
		)
		return
	}

	// User must specify an Ably Control API Url
	var url string
	if config.Url.IsUnknown() {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Cannot use unknown value as Ably Control API URL. Ensure the provider's url parameter is configured",
		)
		return
	}

	if config.Url.IsNull() {
		url = os.Getenv("ABLY_URL")
	} else {
		url = config.Url.ValueString()
	}

	// Create a new Ably client and set it to the provider client
	// Use const CONTROL_API_DEFAULT_URL if url is empty
	if url == "" {
		url = CONTROL_API_DEFAULT_URL
	}
	c, _, err := ably_control_go.NewClientWithURL(token, url)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Unable to create Ably client:\n\n"+err.Error(),
		)
		return
	}
	c.AppendAblyAgent("terraform-provider-ably", p.version)

	p.client = &c
	p.configured = true
}

// Resources - Gets the resources that this provider provides
func (p *AblyProvider) Resources(context.Context) []func() tfsdk_resource.Resource {
	return []func() tfsdk_resource.Resource{
		func() tfsdk_resource.Resource { return resourceApp{p} },
		func() tfsdk_resource.Resource { return resourceNamespace{p} },
		func() tfsdk_resource.Resource { return resourceKey{p} },
		func() tfsdk_resource.Resource { return resourceQueue{p} },
		func() tfsdk_resource.Resource { return resourceRuleKinesis{p} },
		func() tfsdk_resource.Resource { return resourceRuleSqs{p} },
		func() tfsdk_resource.Resource { return resourceRuleLambda{p} },
		func() tfsdk_resource.Resource { return resourceRulePulsar{p} },
		func() tfsdk_resource.Resource { return resourceRuleZapier{p} },
		func() tfsdk_resource.Resource { return resourceRuleGoogleFunction{p} },
		func() tfsdk_resource.Resource { return resourceRuleIFTTT{p} },
		func() tfsdk_resource.Resource { return resourceRuleCloudflareWorker{p} },
		func() tfsdk_resource.Resource { return resourceRuleAzureFunction{p} },
		func() tfsdk_resource.Resource { return resourceRuleHTTP{p} },
		func() tfsdk_resource.Resource { return resourceRuleKafka{p} },
		func() tfsdk_resource.Resource { return resourceRuleAmqp{p} },
		func() tfsdk_resource.Resource { return resourceRuleAmqpExternal{p} },
		func() tfsdk_resource.Resource { return resourceIngressRuleMongo{p} },
		func() tfsdk_resource.Resource { return resourceIngressRulePostgresOutbox{p} },
	}

}

// DataSources - Gets the data sources this provider provides
func (p *AblyProvider) DataSources(context.Context) []func() tfsdk_datasource.DataSource {
	return []func() tfsdk_datasource.DataSource{}

}
