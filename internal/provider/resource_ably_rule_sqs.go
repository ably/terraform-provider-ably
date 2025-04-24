// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceRuleSqs struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleSqs{}
var _ resource.ResourceWithImportState = &ResourceRuleSqs{}

// Schema defines the schema for the resource.
func (r ResourceRuleSqs) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetRuleSchema(
		map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "The region is which AWS SQS is hosted",
			},
			"aws_account_id": schema.StringAttribute{
				Optional:    true,
				Description: "Your AWS account ID",
			},
			"queue_name": schema.StringAttribute{
				Optional:    true,
				Description: "The AWS SQS queue name",
			},
			"enveloped":      GetEnvelopedSchema(),
			"format":         GetFormatSchema(),
			"authentication": GetAwsAuthSchema(),
		},
		"The `ably_rule_sqs` resource allows you to create and manage an Ably integration rule for AWS SQS. Read more at https://ably.com/docs/general/firehose/sqs-rule",
	)
}

func (r ResourceRuleSqs) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_sqs"
}

func (r *ResourceRuleSqs) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleSqs) Name() string {
	return "AWS Sqs"
}

// Create creates a new resource.
func (r ResourceRuleSqs) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetSqs](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRuleSqs) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetSqs](&r, ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceRuleSqs) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetSqs](&r, ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceRuleSqs) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetSqs](&r, ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceRuleSqs) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
