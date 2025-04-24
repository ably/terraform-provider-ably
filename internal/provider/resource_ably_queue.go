package ably_control

import (
	"context"
	"fmt"

	control "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var _ resource.Resource = &ResourceQueue{}
var _ resource.ResourceWithImportState = &ResourceQueue{}
var _ resource.ResourceWithModifyPlan = &ResourceQueue{}

type ResourceQueue struct {
	p *AblyProvider
}

// Get Queue Resource schema
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
			},
			"ttl": schema.Int64Attribute{
				Required:    true,
				Description: "Time to live in minutes.",
			},
			"max_length": schema.Int64Attribute{
				Required:    true,
				Description: "Message limit in number of messages.",
			},
			"region": schema.StringAttribute{
				Required:    true,
				Description: "The data center region. US East (Virginia) or EU West (Ireland). Values are us-east-1-a or eu-west-1-a.",
			},

			"amqp_uri": schema.StringAttribute{
				Computed:    true,
				Description: "URI for the AMQP queue interface.",
			},
			"amqp_queue_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the Ably queue.",
			},
			"stomp_uri": schema.StringAttribute{
				Computed:    true,
				Description: "URI for the STOMP queue interface.",
			},
			"stomp_host": schema.StringAttribute{
				Computed:    true,
				Description: "The host type for the queue.",
			},
			"stomp_destination": schema.StringAttribute{
				Computed:    true,
				Description: "Destination queue.",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "The current state of the queue.",
			},
			"messages_ready": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of ready messages in the queue.",
			},
			"messages_unacknowledged": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of unacknowledged messages in the queue.",
			},
			"messages_total": schema.Int64Attribute{
				Computed:    true,
				Description: "The total number of messages in the queue.",
			},
			"stats_publish_rate": schema.Float64Attribute{
				Computed:    true,
				Description: "The rate at which messages are published to the queue. Rate is messages per minute.",
			},
			"stats_delivery_rate": schema.Float64Attribute{
				Computed:    true,
				Description: "The rate at which messages are delivered from the queue. Rate is messages per minute.",
			},
			"stats_acknowledgement_rate": schema.Float64Attribute{
				Computed:    true,
				Description: "The rate at which messages are acknowledged. Rate is messages per minute.",
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

// Create a new resource
func (r ResourceQueue) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Checks whether the provider and API Client are configured. If they are not, the provider responds with an error.
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply",
		)
		return
	}

	// Gets plan values
	var plan AblyQueue
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var region control.Region
	switch plan.Region.ValueString() {
	case string(control.UsEast1A):
		region = control.UsEast1A
	case string(control.EuWest1A):
		region = control.EuWest1A
	default:
		resp.Diagnostics.AddError(
			"Provider not configured",
			fmt.Sprintf("Invalid value for Queue.Region '%s'", plan.Region.ValueString()),
		)
		return
	}

	// Generates an API request body from the plan values
	queue_values := control.NewQueue{
		Name:      plan.Name.ValueString(),
		Ttl:       int(plan.Ttl.ValueInt64()),
		MaxLength: int(plan.MaxLength.ValueInt64()),
		Region:    region,
	}

	// Creates a new Ably queue by invoking the CreateQueue function from the Client Library
	ably_queue, err := r.p.client.CreateQueue(plan.AppID.ValueString(), &queue_values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Maps response body to resource schema attributes.
	resp_apps := AblyQueue{
		AppID:     types.StringValue(plan.AppID.ValueString()),
		ID:        types.StringValue(ably_queue.ID),
		Name:      types.StringValue(ably_queue.Name),
		Ttl:       types.Int64Value(int64(ably_queue.Ttl)),
		MaxLength: types.Int64Value(int64(ably_queue.MaxLength)),
		Region:    types.StringValue(string(ably_queue.Region)),

		AmqpUri:                  types.StringValue(ably_queue.Amqp.Uri),
		AmqpQueueName:            types.StringValue(ably_queue.Amqp.QueueName),
		StompURI:                 types.StringValue(ably_queue.Stomp.Uri),
		StompHost:                types.StringValue(ably_queue.Stomp.Host),
		StompDestination:         types.StringValue(ably_queue.Stomp.Destination),
		State:                    types.StringValue(ably_queue.State),
		MessagesReady:            types.Int64Value(int64(ably_queue.Messages.Ready)),
		MessagesUnacknowledged:   types.Int64Value(int64(ably_queue.Messages.Unacknowledged)),
		MessagesTotal:            types.Int64Value(int64(ably_queue.Messages.Total)),
		StatsPublishRate:         types.Float64Value(ably_queue.Stats.PublishRate),
		StatsDeliveryRate:        types.Float64Value(ably_queue.Stats.DeliveryRate),
		StatsAcknowledgementRate: types.Float64Value(ably_queue.Stats.AcknowledgementRate),
		Deadletter:               types.BoolValue(ably_queue.DeadLetter),
		DeadletterID:             types.StringValue(ably_queue.DeadLetterID),
	}

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, resp_apps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource
func (r ResourceQueue) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var state AblyQueue
	found := false
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID and queue ID value for the resource
	app_id := state.AppID.ValueString()
	queue_id := state.ID.ValueString()

	// Fetches all Ably Queues in the app. The function invokes the Client Library Queues() method.
	// NOTE: Control API & Client Lib do not currently support fetching single queue given queue id
	queues, err := r.p.client.Queues(app_id)
	if err != nil {
		if is_404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error updating Resource",
			"Could not update resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Loops through queues and if id matches, sets state.
	for _, v := range queues {
		if v.ID == queue_id {
			resp_queues := AblyQueue{
				AppID:     types.StringValue(v.AppID),
				ID:        types.StringValue(v.ID),
				Name:      types.StringValue(v.Name),
				Ttl:       types.Int64Value(int64(v.Ttl)),
				MaxLength: types.Int64Value(int64(v.MaxLength)),
				Region:    types.StringValue(string(v.Region)),

				AmqpUri:                  types.StringValue(v.Amqp.Uri),
				AmqpQueueName:            types.StringValue(v.Amqp.QueueName),
				StompURI:                 types.StringValue(v.Stomp.Uri),
				StompHost:                types.StringValue(v.Stomp.Host),
				StompDestination:         types.StringValue(v.Stomp.Destination),
				State:                    types.StringValue(v.State),
				MessagesReady:            types.Int64Value(int64(v.Messages.Ready)),
				MessagesUnacknowledged:   types.Int64Value(int64(v.Messages.Unacknowledged)),
				MessagesTotal:            types.Int64Value(int64(v.Messages.Total)),
				StatsPublishRate:         types.Float64Value(v.Stats.PublishRate),
				StatsDeliveryRate:        types.Float64Value(v.Stats.DeliveryRate),
				StatsAcknowledgementRate: types.Float64Value(v.Stats.AcknowledgementRate),
				Deadletter:               types.BoolValue(v.DeadLetter),
				DeadletterID:             types.StringValue(v.DeadLetterID),
			}
			// Sets state to queue values.
			diags = resp.State.Set(ctx, &resp_queues)
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

// Update resource
func (r ResourceQueue) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This function should never end up being run but needs to exist to satisfy the interface
	// this error is just in case terraform decides to call it.
	resp.Diagnostics.AddError(
		"Error modifying Resource",
		"Queue can not be modified",
	)
}

// Delete resource
func (r ResourceQueue) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state AblyQueue
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the current state. If it is unable to, the provider responds with an error.
	app_id := state.AppID.ValueString()
	queue_id := state.ID.ValueString()

	err := r.p.client.DeleteQueue(app_id, queue_id)
	if err != nil {
		if is_404(err) {
			resp.Diagnostics.AddWarning(
				"Resource does not exist",
				"Resource does not exist, it may have already been deleted: "+err.Error(),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error deleting Resource",
				"Could not delete resource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

// Import resource
func (r ResourceQueue) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "id")
}

func (r ResourceQueue) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Make all attributes require replace
	// Get all attributes from the schema using a temporary response
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	// Mark all attributes as requiring replacement
	for attrName := range schemaResp.Schema.Attributes {
		resp.RequiresReplace.Append(path.Root(attrName))
	}
}
