package ably_control

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	control "github.com/ably/ably-control-go"
)

const CONTROL_API_DEFAULT_URL = "https://control.ably.net/v1"

// Ensure AblyProvider satisfies various provider interfaces.
var _ provider.Provider = &AblyProvider{}

type AblyProvider struct {
	configured bool
	client     *control.Client
	version    string
}

func (p *AblyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ably"
	resp.Version = p.version
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AblyProvider{
			version: version,
		}
	}
}

func (p *AblyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Sensitive: true,
				Optional:  true,
			},
			"url": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// Provider schema struct
type AblyProviderData struct {
	Token types.String `tfsdk:"token"`
	Url   types.String `tfsdk:"url"`
}

func (p *AblyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
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
	c, _, err := control.NewClientWithURL(token, url)

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
func (p *AblyProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return ResourceApp{p} },
		func() resource.Resource { return ResourceNamespace{p} },
		func() resource.Resource { return &ResourceKey{p} },
		func() resource.Resource { return ResourceQueue{p} },
		func() resource.Resource { return ResourceRuleKinesis{p} },
		func() resource.Resource { return ResourceRuleSqs{p} },
		func() resource.Resource { return ResourceRuleLambda{p} },
		func() resource.Resource { return ResourceRulePulsar{p} },
		func() resource.Resource { return ResourceRuleZapier{p} },
		func() resource.Resource { return ResourceRuleGoogleFunction{p} },
		func() resource.Resource { return ResourceRuleIFTTT{p} },
		func() resource.Resource { return ResourceRuleCloudflareWorker{p} },
		func() resource.Resource { return ResourceRuleAzureFunction{p} },
		func() resource.Resource { return ResourceRuleHTTP{p} },
		func() resource.Resource { return ResourceRuleKafka{p} },
		func() resource.Resource { return ResourceRuleAmqp{p} },
		func() resource.Resource { return ResourceRuleAmqpExternal{p} },
		func() resource.Resource { return ResourceIngressRuleMongo{p} },
		func() resource.Resource { return ResourceIngressRulePostgresOutbox{p} },
	}

}

// DataSources - Gets the data sources this provider provides
func (p *AblyProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
