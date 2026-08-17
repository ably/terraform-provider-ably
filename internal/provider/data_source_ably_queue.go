// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/ably/terraform-provider-ably/internal/provider/codegen/datasource_queues"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AblyQueueAMQPDataSourceModel mirrors control.QueueAMQP.
type AblyQueueAMQPDataSourceModel struct {
	URI       types.String `tfsdk:"uri"`
	QueueName types.String `tfsdk:"queue_name"`
}

// AblyQueueStompDataSourceModel mirrors control.QueueStomp.
type AblyQueueStompDataSourceModel struct {
	URI         types.String `tfsdk:"uri"`
	Host        types.String `tfsdk:"host"`
	Destination types.String `tfsdk:"destination"`
}

// AblyQueueMessagesDataSourceModel mirrors control.QueueMessages.
type AblyQueueMessagesDataSourceModel struct {
	Ready          types.Int64 `tfsdk:"ready"`
	Unacknowledged types.Int64 `tfsdk:"unacknowledged"`
	Total          types.Int64 `tfsdk:"total"`
}

// AblyQueueStatsDataSourceModel mirrors control.QueueStats.
type AblyQueueStatsDataSourceModel struct {
	PublishRate         types.Float64 `tfsdk:"publish_rate"`
	DeliveryRate        types.Float64 `tfsdk:"delivery_rate"`
	AcknowledgementRate types.Float64 `tfsdk:"acknowledgement_rate"`
}

// AblyQueueDataSourceModel is one queue as the data sources report it. It serves
// both ably_queue and the elements of ably_queues.
//
// Unlike the ably_queue resource, which flattens the API's amqp and stomp objects
// into amqp_uri/stomp_uri and drops the live counters, the data sources carry the
// API's own shape: they are read-only, so there is no state compatibility to
// preserve and no reason to hide the messages and stats blocks.
type AblyQueueDataSourceModel struct {
	ID           types.String                      `tfsdk:"id"`
	AppID        types.String                      `tfsdk:"app_id"`
	Name         types.String                      `tfsdk:"name"`
	Region       types.String                      `tfsdk:"region"`
	State        types.String                      `tfsdk:"state"`
	TTL          types.Int64                       `tfsdk:"ttl"`
	MaxLength    types.Int64                       `tfsdk:"max_length"`
	Deadletter   types.Bool                        `tfsdk:"deadletter"`
	DeadletterID types.String                      `tfsdk:"deadletter_id"`
	AMQP         *AblyQueueAMQPDataSourceModel     `tfsdk:"amqp"`
	Stomp        *AblyQueueStompDataSourceModel    `tfsdk:"stomp"`
	Messages     *AblyQueueMessagesDataSourceModel `tfsdk:"messages"`
	Stats        *AblyQueueStatsDataSourceModel    `tfsdk:"stats"`
}

// AblyQueuesDataSourceModel is the model for the plural ably_queues data source.
type AblyQueuesDataSourceModel struct {
	AppID  types.String               `tfsdk:"app_id"`
	Queues []AblyQueueDataSourceModel `tfsdk:"queues"`
}

// queueDataSourceModel maps an API queue onto the shared model.
func queueDataSourceModel(queue control.QueueResponse) AblyQueueDataSourceModel {
	return AblyQueueDataSourceModel{
		ID:           types.StringValue(queue.ID),
		AppID:        types.StringValue(queue.AppID),
		Name:         types.StringValue(queue.Name),
		Region:       types.StringValue(queue.Region),
		State:        stringOrNull(queue.State),
		TTL:          types.Int64Value(int64(queue.TTL)),
		MaxLength:    types.Int64Value(int64(queue.MaxLength)),
		Deadletter:   types.BoolValue(queue.Deadletter),
		DeadletterID: optStringValue(queue.DeadletterID),
		AMQP: &AblyQueueAMQPDataSourceModel{
			URI:       stringOrNull(queue.AMQP.URI),
			QueueName: stringOrNull(queue.AMQP.QueueName),
		},
		Stomp: &AblyQueueStompDataSourceModel{
			URI:         stringOrNull(queue.Stomp.URI),
			Host:        stringOrNull(queue.Stomp.Host),
			Destination: stringOrNull(queue.Stomp.Destination),
		},
		Messages: &AblyQueueMessagesDataSourceModel{
			Ready:          optIntValue(queue.Messages.Ready),
			Unacknowledged: optIntValue(queue.Messages.Unacknowledged),
			Total:          optIntValue(queue.Messages.Total),
		},
		Stats: &AblyQueueStatsDataSourceModel{
			PublishRate:         optFloat64Value(queue.Stats.PublishRate),
			DeliveryRate:        optFloat64Value(queue.Stats.DeliveryRate),
			AcknowledgementRate: optFloat64Value(queue.Stats.AcknowledgementRate),
		},
	}
}

// queuesDataSourceSchema returns the generated ably_queues schema, ready to
// serve.
func queuesDataSourceSchema(ctx context.Context) schema.Schema {
	s := datasource_queues.QueuesDataSourceSchema(ctx)
	stripSetNestedCustomTypes(&s, "queues")
	return s
}

// --- ably_queues -----------------------------------------------------------

type DataSourceQueues struct {
	p *AblyProvider
}

var _ datasource.DataSource = &DataSourceQueues{}

func (d DataSourceQueues) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "ably_queues"
}

// Schema defines the schema for the data source.
//
// GENERATED SCHEMA: the attribute set, types, nesting and descriptions come from
// internal/provider/codegen, produced by `make generate` from the Control API
// spec.
func (d DataSourceQueues) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	s := queuesDataSourceSchema(ctx)

	s.MarkdownDescription = "The `ably_queues` data source lists every queue in an Ably app, including queues Terraform does not manage, with their connection details and live message counts. Use `ably_queue` to look up a single queue by id or name."

	resp.Schema = s
}

func (d DataSourceQueues) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	var config AblyQueuesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	queues, err := d.p.client.ListQueues(ctx, config.AppID.ValueString())
	if err != nil {
		readDataSourceError(&resp.Diagnostics, "ably_queues", err)
		return
	}

	state := AblyQueuesDataSourceModel{
		AppID:  config.AppID,
		Queues: make([]AblyQueueDataSourceModel, 0, len(queues)),
	}
	for _, queue := range queues {
		state.Queues = append(state.Queues, queueDataSourceModel(queue))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// --- ably_queue ------------------------------------------------------------

type DataSourceQueue struct {
	p *AblyProvider
}

var _ datasource.DataSource = &DataSourceQueue{}

func (d DataSourceQueue) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "ably_queue"
}

// Schema defines the schema for the data source. The attributes are lifted from
// the generated ably_queues element, so this schema tracks the spec too.
func (d DataSourceQueue) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := elementAttributes(queuesDataSourceSchema(ctx), "queues")

	requireString(attributes, "app_id", "The Ably app the queue belongs to.")
	optionalString(attributes, "id", "The queue ID. Set either this or name.")
	optionalString(attributes, "name", "The queue name. Set either this or id. Names are not unique, so a name matching more than one queue is an error.")

	resp.Schema = schema.Schema{
		Attributes:          attributes,
		MarkdownDescription: "The `ably_queue` data source looks up a single Ably queue by id or name, with its AMQP and STOMP connection details. The Control API has no fetch-by-id endpoint for queues, so this lists the app's queues and matches locally.",
	}
}

func (d DataSourceQueue) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	var config AblyQueueDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	queues, err := d.p.client.ListQueues(ctx, config.AppID.ValueString())
	if err != nil {
		readDataSourceError(&resp.Diagnostics, "ably_queue", err)
		return
	}

	queue, diags := findOne(
		lookup{dataSourceName: "ably_queue", id: config.ID, name: config.Name},
		queues,
		func(q control.QueueResponse) string { return q.ID },
		func(q control.QueueResponse) string { return q.Name },
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := queueDataSourceModel(queue)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
