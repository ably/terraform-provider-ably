package ably_control

import (
	"context"
	"os"

	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const CONTROL_API_DEFAULT_URL = "https://control.ably.net/v1"

func New() tfsdk_provider.Provider {
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
type providerData struct {
	Token types.String `tfsdk:"token"`
	Url   types.String `tfsdk:"url"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk_provider.ConfigureRequest, resp *tfsdk_provider.ConfigureResponse) {
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
			"Ably API Token required",
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
			"Unable to find Ably API token",
			"Ably API token cannot be an empty string. Ensure the providers token parameter is configured",
		)
		return
	}

	// User must specify an Ably Control API Url
	var url string
	if config.Url.Unknown {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Cannot use unknown value as Ably Control API URL. Ensure the provider's url parameter is configured",
		)
		return
	}

	if config.Url.Null {
		url = os.Getenv("ABLY_URL")
	} else {
		url = config.Url.Value
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

	p.client = &c
	p.configured = true
}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk_provider.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk_provider.ResourceType{
		"ably_app":          resourceAppType{},
		"ably_namespace":    resourceNamespaceType{},
		"ably_api_key":      resourceKeyType{},
		"ably_queue":        resourceQueueType{},
		"ably_rule_kinesis": resourceRuleKinesisType{},
		"ably_rule_sqs":     resourceRuleSqsType{},
		"ably_rule_lambda":  resourceRuleLambdaType{},
		"ably_rule_zapier":  resourceRuleZapierType{},
		"ably_rule_ifttt":   resourceRuleIFTTTType{},
	}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk_provider.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk_provider.DataSourceType{}, nil
}
