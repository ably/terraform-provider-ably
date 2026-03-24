// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ably/terraform-provider-ably/control"
)

const controlAPIDefaultURL = "https://control.ably.net/v1"

// Ensure AblyProvider satisfies various provider interfaces.
var _ provider.Provider = &AblyProvider{}

type AblyProvider struct {
	// configured is set to true after the provider has been successfully configured.
	// All CRUD methods (Create, Read, Update, Delete) in resources should check
	// p.configured before making API calls to ensure the provider is properly initialized.
	configured bool
	client     *control.Client
	accountID  string
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
		MarkdownDescription: "The Ably provider allows you to manage [Ably](https://ably.com) resources including apps, keys, namespaces, queues, and integration rules.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Description: "The Ably account token used for authentication. Can also be set via the `ABLY_ACCOUNT_TOKEN` environment variable.",
				Sensitive:   true,
				Optional:    true,
			},
			"url": schema.StringAttribute{
				Description: "The Ably Control API URL. Can also be set via the `ABLY_URL` environment variable. Defaults to the production API.",
				Optional:    true,
			},
		},
	}
}

// AblyProviderData contains configuration data for the Ably provider.
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
		resp.Diagnostics.AddError(
			"Unable to create client",
			"The Ably API token is unknown. The provider cannot be configured without a known token value.",
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
	// Use const controlAPIDefaultURL if url is empty
	if url == "" {
		url = controlAPIDefaultURL
	}
	c := control.NewClient(token)
	c.BaseURL = url
	c.UserAgent += " terraform-provider-ably/" + p.version

	p.client = c

	// Fetch account ID via /me endpoint for use by resources that need it.
	me, err := c.Me(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to fetch account information",
			"Could not retrieve account ID from Ably API: "+err.Error(),
		)
		return
	}

	if me.Account == nil || me.Account.ID == "" {
		resp.Diagnostics.AddError(
			"Unable to determine account ID",
			"Failed to determine account ID from the Ably API. Please verify your token has account-level access.",
		)
		return
	}

	p.accountID = me.Account.ID
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
		func() resource.Resource { return ResourceRuleAMQP{p} },
		func() resource.Resource { return ResourceRuleAMQPExternal{p} },
		func() resource.Resource { return ResourceIngressRuleMongo{p} },
		func() resource.Resource { return ResourceIngressRulePostgresOutbox{p} },
	}

}

// DataSources - Gets the data sources this provider provides
func (p *AblyProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
