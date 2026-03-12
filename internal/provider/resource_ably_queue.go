// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	control "github.com/ably/terraform-provider-ably/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var _ resource.Resource = &ResourceQueue{}
var _ resource.ResourceWithImportState = &ResourceQueue{}

type ResourceQueue struct {
	p *AblyProvider
}

// Schema defines the schema for the resource.
func (r ResourceQueue) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app_id": schema.StringAttribute{
				Required:    true,
				Description: "The application ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the queue",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the queue.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ttl": schema.Int64Attribute{
				Required:    true,
				Description: "Time to live in minutes.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"max_length": schema.Int64Attribute{
				Required:    true,
				Description: "Message limit in number of messages.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"region": schema.StringAttribute{
				Required:    true,
				Description: "The data center region. US East (Virginia) or EU West (Ireland). Values are us-east-1-a or eu-west-1-a.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("us-east-1-a", "eu-west-1-a"),
				},
			},

			"amqp_uri": schema.StringAttribute{
				Computed:    true,
				Description: "URI for the AMQP queue interface.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"amqp_queue_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the Ably queue.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stomp_uri": schema.StringAttribute{
				Computed:    true,
				Description: "URI for the STOMP queue interface.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stomp_host": schema.StringAttribute{
				Computed:    true,
				Description: "The host type for the queue.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stomp_destination": schema.StringAttribute{
				Computed:    true,
				Description: "Destination queue.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "The current state of the queue.",
			},
			"deadletter": schema.BoolAttribute{
				Computed:    true,
				Description: "A boolean that indicates whether this is a dead letter queue or not.",
			},
			"deadletter_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the dead letter queue.",
			},
		},
		MarkdownDescription: "The ably_queue resource allows you to create and manage Ably queues. Read more about Ably queues in Ably documentation: https://ably.com/docs/general/queues.",
	}
}

func (r ResourceQueue) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_queue"
}

// Create creates a new resource.
func (r ResourceQueue) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	// Gets plan values
	var plan AblyQueue
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generates an API request body from the plan values
	queueValues := control.Queue{
		Name:      plan.Name.ValueString(),
		TTL:       int(plan.Ttl.ValueInt64()),
		MaxLength: int(plan.MaxLength.ValueInt64()),
		Region:    plan.Region.ValueString(),
	}

	// Creates a new Ably queue by invoking the CreateQueue function from the Client Library
	ablyQueue, err := r.p.client.CreateQueue(ctx, plan.AppID.ValueString(), queueValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ably_queue",
			"Could not create ably_queue, unexpected error: "+err.Error(),
		)
		return
	}

	// Maps response body to resource schema attributes.
	respApps := AblyQueue{
		AppID:     types.StringValue(plan.AppID.ValueString()),
		ID:        types.StringValue(ablyQueue.ID),
		Name:      types.StringValue(ablyQueue.Name),
		Ttl:       types.Int64Value(int64(ablyQueue.TTL)),
		MaxLength: types.Int64Value(int64(ablyQueue.MaxLength)),
		Region:    types.StringValue(ablyQueue.Region),

		AmqpUri:          types.StringValue(ablyQueue.AMQP.URI),
		AmqpQueueName:    types.StringValue(ablyQueue.AMQP.QueueName),
		StompURI:         types.StringValue(ablyQueue.Stomp.URI),
		StompHost:        types.StringValue(ablyQueue.Stomp.Host),
		StompDestination: types.StringValue(ablyQueue.Stomp.Destination),
		State:            types.StringValue(ablyQueue.State),
		Deadletter:       types.BoolValue(ablyQueue.Deadletter),
		DeadletterID:     optStringValue(ablyQueue.DeadletterID),
	}

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, respApps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read reads the resource.
func (r ResourceQueue) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	// Gets the current state. If it is unable to, the provider responds with an error.
	var state AblyQueue
	found := false
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID and queue ID value for the resource
	appID := state.AppID.ValueString()
	queueID := state.ID.ValueString()

	// Fetches all Ably Queues in the app. The function invokes the Client Library Queues() method.
	// NOTE: Control API & Client Lib do not currently support fetching single queue given queue id
	queues, err := r.p.client.ListQueues(ctx, appID)
	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading ably_queue",
			"Could not read ably_queue, unexpected error: "+err.Error(),
		)
		return
	}

	// Loops through queues and if id matches, sets state.
	for _, v := range queues {
		if v.ID == queueID {
			respQueues := AblyQueue{
				AppID:     types.StringValue(v.AppID),
				ID:        types.StringValue(v.ID),
				Name:      types.StringValue(v.Name),
				Ttl:       types.Int64Value(int64(v.TTL)),
				MaxLength: types.Int64Value(int64(v.MaxLength)),
				Region:    types.StringValue(v.Region),

				AmqpUri:          types.StringValue(v.AMQP.URI),
				AmqpQueueName:    types.StringValue(v.AMQP.QueueName),
				StompURI:         types.StringValue(v.Stomp.URI),
				StompHost:        types.StringValue(v.Stomp.Host),
				StompDestination: types.StringValue(v.Stomp.Destination),
				State:            types.StringValue(v.State),
				Deadletter:       types.BoolValue(v.Deadletter),
				DeadletterID:     optStringValue(v.DeadletterID),
			}
			// Sets state to queue values.
			diags = resp.State.Set(ctx, &respQueues)
			found = true

			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
	}
}

// Update updates an existing resource.
func (r ResourceQueue) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	// This function should never end up being run but needs to exist to satisfy the interface
	// this error is just in case terraform decides to call it.
	resp.Diagnostics.AddError(
		"Error updating ably_queue",
		"ably_queue can not be modified",
	)
}

// Delete deletes the resource.
func (r ResourceQueue) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	// Get current state
	var state AblyQueue
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the current state. If it is unable to, the provider responds with an error.
	appID := state.AppID.ValueString()
	queueID := state.ID.ValueString()

	err := r.p.client.DeleteQueue(ctx, appID, queueID)
	if err != nil {
		if is404(err) {
			resp.Diagnostics.AddWarning(
				"Resource does not exist",
				"Resource does not exist, it may have already been deleted: "+err.Error(),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error deleting ably_queue",
				"Could not delete ably_queue, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

// ImportState handles the import state functionality.
func (r ResourceQueue) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
