package ably_control

import (
	"context"
	"fmt"

	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"

	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceQueueType struct{}

// Get Queue Resource schema
func (r resourceQueueType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
				Type:        types.Int64Type,
				Computed:    true,
				Description: "The rate at which messages are published to the queue. Rate is messages per minute.",
			},
			"stats_delivery_rate": {
				Type:        types.Int64Type,
				Computed:    true,
				Description: "The rate at which messages are delivered from the queue. Rate is messages per minute.",
			},
			"stats_acknowledgement_rate": {
				Type:        types.Int64Type,
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

// New resource instance
func (r resourceQueueType) NewResource(_ context.Context, p tfsdk_provider.Provider) (tfsdk_resource.Resource, diag.Diagnostics) {
	return resourceQueue{
		p: *(p.(*provider)),
	}, nil
}

type resourceQueue struct {
	p provider
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

	var region ably_control_go.Region
	switch plan.Region.Value {
	case string(ably_control_go.UsEast1A):
		region = ably_control_go.UsEast1A
	case string(ably_control_go.EuWest1A):
		region = ably_control_go.EuWest1A
	default:
		resp.Diagnostics.AddError(
			"Provider not configured",
			fmt.Sprintf("Invalid value for Queue.Region '%s'", plan.Region.Value),
		)
		return
	}

	// Generates an API request body from the plan values
	queue_values := ably_control_go.NewQueue{
		Name:      plan.Name.Value,
		Ttl:       int(plan.Ttl.Value),
		MaxLength: int(plan.MaxLength.Value),
		Region:    region,
	}

	// Creates a new Ably queue by invoking the CreateQueue function from the Client Library
	ably_queue, err := r.p.client.CreateQueue(plan.AppID.Value, &queue_values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Maps response body to resource schema attributes.
	resp_apps := AblyQueue{
		AppID:     types.String{Value: plan.AppID.Value},
		ID:        types.String{Value: ably_queue.ID},
		Name:      types.String{Value: ably_queue.Name},
		Ttl:       types.Int64{Value: int64(ably_queue.Ttl)},
		MaxLength: types.Int64{Value: int64(ably_queue.MaxLength)},
		Region:    types.String{Value: string(ably_queue.Region)},

		AmqpUri:                  types.String{Value: ably_queue.Amqp.Uri},
		AmqpQueueName:            types.String{Value: ably_queue.Amqp.QueueName},
		StompURI:                 types.String{Value: ably_queue.Stomp.Uri},
		StompHost:                types.String{Value: ably_queue.Stomp.Host},
		StompDestination:         types.String{Value: ably_queue.Stomp.Destination},
		State:                    types.String{Value: ably_queue.State},
		MessagesReady:            types.Int64{Value: int64(ably_queue.Messages.Ready)},
		MessagesUnacknowledged:   types.Int64{Value: int64(ably_queue.Messages.Unacknowledged)},
		MessagesTotal:            types.Int64{Value: int64(ably_queue.Messages.Total)},
		StatsPublishRate:         types.Int64{Value: int64(ably_queue.Stats.PublishRate)},
		StatsDeliveryRate:        types.Int64{Value: int64(ably_queue.Stats.DeliveryRate)},
		StatsAcknowledgementRate: types.Int64{Value: int64(ably_queue.Stats.AcknowledgementRate)},
		Deadletter:               types.Bool{Value: ably_queue.DeadLetter},
		DeadletterID:             types.String{Value: ably_queue.DeadLetterID},
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
	app_id := state.AppID.Value
	queue_id := state.ID.Value

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
				AppID:     types.String{Value: v.AppID},
				ID:        types.String{Value: v.ID},
				Name:      types.String{Value: v.Name},
				Ttl:       types.Int64{Value: int64(v.Ttl)},
				MaxLength: types.Int64{Value: int64(v.MaxLength)},
				Region:    types.String{Value: string(v.Region)},

				AmqpUri:                  types.String{Value: v.Amqp.Uri},
				AmqpQueueName:            types.String{Value: v.Amqp.QueueName},
				StompURI:                 types.String{Value: v.Stomp.Uri},
				StompHost:                types.String{Value: v.Stomp.Host},
				StompDestination:         types.String{Value: v.Stomp.Destination},
				State:                    types.String{Value: v.State},
				MessagesReady:            types.Int64{Value: int64(v.Messages.Ready)},
				MessagesUnacknowledged:   types.Int64{Value: int64(v.Messages.Unacknowledged)},
				MessagesTotal:            types.Int64{Value: int64(v.Messages.Total)},
				StatsPublishRate:         types.Int64{Value: int64(v.Stats.PublishRate)},
				StatsDeliveryRate:        types.Int64{Value: int64(v.Stats.DeliveryRate)},
				StatsAcknowledgementRate: types.Int64{Value: int64(v.Stats.AcknowledgementRate)},
				Deadletter:               types.Bool{Value: v.DeadLetter},
				DeadletterID:             types.String{Value: v.DeadLetterID},
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
		"Error Modifing Resource",
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
	app_id := state.AppID.Value
	queue_id := state.ID.Value

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
