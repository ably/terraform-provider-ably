package ably_control

import (
	"context"
	"fmt"

	control "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"

	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceQueue struct {
	p *AblyProvider
}

// Get Queue Resource schema
func (r resourceQueue) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"app_id": {
				Type:        types.StringType,
				Required:    true,
				Description: "The application ID.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk_resource.RequiresReplace(),
				},
			},
			"id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The ID of the queue",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk_resource.UseStateForUnknown(),
				},
			},
			"name": {
				Type:        types.StringType,
				Required:    true,
				Description: "The name of the queue.",
			},
			"ttl": {
				Type:        types.Int64Type,
				Required:    true,
				Description: "Time to live in minutes.",
			},
			"max_length": {
				Type:        types.Int64Type,
				Required:    true,
				Description: "Message limit in number of messages.",
			},
			"region": {
				Type:        types.StringType,
				Required:    true,
				Description: "The data center region. US East (Virginia) or EU West (Ireland). Values are us-east-1-a or eu-west-1-a.",
			},

			"amqp_uri": {
				Type:        types.StringType,
				Computed:    true,
				Description: "URI for the AMQP queue interface.",
			},
			"amqp_queue_name": {
				Type:        types.StringType,
				Computed:    true,
				Description: "Name of the Ably queue.",
			},
			"stomp_uri": {
				Type:        types.StringType,
				Computed:    true,
				Description: "URI for the STOMP queue interface.",
			},
			"stomp_host": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The host type for the queue.",
			},
			"stomp_destination": {
				Type:        types.StringType,
				Computed:    true,
				Description: "Destination queue.",
			},
			"state": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The current state of the queue.",
			},
			"messages_ready": {
				Type:        types.Int64Type,
				Computed:    true,
				Description: "The number of ready messages in the queue.",
			},
			"messages_unacknowledged": {
				Type:        types.Int64Type,
				Computed:    true,
				Description: "The number of unacknowledged messages in the queue.",
			},
			"messages_total": {
				Type:        types.Int64Type,
				Computed:    true,
				Description: "The total number of messages in the queue.",
			},
			"stats_publish_rate": {
				Type:        types.Float64Type,
				Computed:    true,
				Description: "The rate at which messages are published to the queue. Rate is messages per minute.",
			},
			"stats_delivery_rate": {
				Type:        types.Float64Type,
				Computed:    true,
				Description: "The rate at which messages are delivered from the queue. Rate is messages per minute.",
			},
			"stats_acknowledgement_rate": {
				Type:        types.Float64Type,
				Computed:    true,
				Description: "The rate at which messages are acknowledged. Rate is messages per minute.",
			},
			"deadletter": {
				Type:        types.BoolType,
				Computed:    true,
				Description: "A boolean that indicates whether this is a dead letter queue or not.",
			},
			"deadletter_id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The ID of the dead letter queue.",
			},
		},
		MarkdownDescription: "The ably_queue resource allows you to create and manage Ably queues. Read more about Ably queues in Ably documentation: https://ably.com/docs/general/queues.",
	}, nil
}

func (r resourceQueue) Metadata(ctx context.Context, req tfsdk_resource.MetadataRequest, resp *tfsdk_resource.MetadataResponse) {
	resp.TypeName = "ably_queue"
}

// Create a new resource
func (r resourceQueue) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
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
func (r resourceQueue) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
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
func (r resourceQueue) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	// This function should never end up being run but needs to exist to satisfy the interface
	// this error is just in case terraform decides to call it.
	resp.Diagnostics.AddError(
		"Error modifying Resource",
		"Queue can not be modified",
	)
}

// Delete resource
func (r resourceQueue) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
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
func (r resourceQueue) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "id")
}

var _ tfsdk_resource.ResourceWithModifyPlan = resourceQueue{}

func (r resourceQueue) ModifyPlan(ctx context.Context, req tfsdk_resource.ModifyPlanRequest, resp *tfsdk_resource.ModifyPlanResponse) {
	for k := range req.Plan.Schema.Attributes {
		resp.RequiresReplace.Append(path.Root(k))
	}
}
