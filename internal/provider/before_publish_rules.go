// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// This file holds the shared plumbing for the moderation and before-publish rule
// families. Unlike webhook/firehose rules (see rules.go, AblyRule and
// CreateRule[T]) these rules have NO `source`* and NO `request_mode`; they carry
// `invocation_mode`, an optional `chat_room_filter` and, for most of them, a
// `before_publish_config` block. Copying the webhook plumbing for them yields a
// resource with a bogus required `source` block and a `request_mode` the API
// rejects, so they wire to the control client through the CRUD in this file
// instead.
//
// * with one exception: before-publish AWS Lambda takes an optional source.

// AblyBeforePublishConfig mirrors control.BeforePublishConfig. Before-publish
// rules run before a message is published, so this block controls retry/backoff
// and what happens when the endpoint fails or rate limits.
//
// Kept separate from AblyRuleBodyguardBeforePublishConfig, which predates this
// file: the two are structurally identical, and bodyguard stays on its own model
// so the reference port in resource_ably_rule_bodyguard.go is not disturbed.
type AblyBeforePublishConfig struct {
	RetryTimeout          types.Int64  `tfsdk:"retry_timeout"`
	MaxRetries            types.Int64  `tfsdk:"max_retries"`
	FailedAction          types.String `tfsdk:"failed_action"`
	TooManyRequestsAction types.String `tfsdk:"too_many_requests_action"`
}

// beforePublishConfigPost converts the plan block into the control create body.
// The block is Required in every generated schema that has it, so a nil pointer
// means Terraform gave us a rule family that does not carry one and the zero
// value is correct.
func beforePublishConfigPost(c *AblyBeforePublishConfig) control.BeforePublishConfig {
	if c == nil {
		return control.BeforePublishConfig{}
	}
	return control.BeforePublishConfig{
		RetryTimeout:          int(c.RetryTimeout.ValueInt64()),
		MaxRetries:            int(c.MaxRetries.ValueInt64()),
		FailedAction:          c.FailedAction.ValueString(),
		TooManyRequestsAction: c.TooManyRequestsAction.ValueString(),
	}
}

// beforePublishConfigResponse maps the API's before-publish config back onto the
// tfsdk model, returning nil when the response omits it.
func beforePublishConfigResponse(c *control.BeforePublishConfig) *AblyBeforePublishConfig {
	if c == nil {
		return nil
	}
	return &AblyBeforePublishConfig{
		RetryTimeout:          types.Int64Value(int64(c.RetryTimeout)),
		MaxRetries:            types.Int64Value(int64(c.MaxRetries)),
		FailedAction:          types.StringValue(c.FailedAction),
		TooManyRequestsAction: types.StringValue(c.TooManyRequestsAction),
	}
}

// thresholdsPost converts a moderation target's thresholds map into the control
// map. Null/unknown becomes nil so the field is omitted from the request body.
func thresholdsPost(ctx context.Context, m types.Map) (map[string]int, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.IsNull() || m.IsUnknown() {
		return nil, diags
	}
	elements := map[string]int64{}
	diags.Append(m.ElementsAs(ctx, &elements, false)...)
	if diags.HasError() {
		return nil, diags
	}
	thresholds := make(map[string]int, len(elements))
	for k, v := range elements {
		thresholds[k] = int(v)
	}
	return thresholds, diags
}

// thresholdsResponse maps the control thresholds map back onto the tfsdk model,
// returning null when the response omits it so an unset optional attribute
// round-trips as null rather than an empty map.
func thresholdsResponse(ctx context.Context, thresholds map[string]int) (types.Map, diag.Diagnostics) {
	if len(thresholds) == 0 {
		return types.MapNull(types.Int64Type), nil
	}
	values := make(map[string]int64, len(thresholds))
	for k, v := range thresholds {
		values[k] = int64(v)
	}
	return types.MapValueFrom(ctx, types.Int64Type, values)
}

// beforePublishCRUD is the CRUD shared by every moderation/before-publish rule
// resource. The per-resource parts are the create body and the response
// mapping; everything else (error wording, 404 handling, state writes) is
// identical, so each resource supplies the two functions and delegates.
//
// M is the resource's tfsdk model.
type beforePublishCRUD[M any] struct {
	p *AblyProvider
	// name is the human-readable rule name used in diagnostics.
	name string
	// appID and id read the parent app and rule IDs out of the model.
	appID func(M) string
	id    func(M) string
	// post builds the Control API create/update body from the model.
	post func(context.Context, M) (any, diag.Diagnostics)
	// response maps an API response onto the model. The second argument is the
	// plan (create/update) or prior state (read), for the fields the API does
	// not return.
	response func(context.Context, *control.RuleResponse, *M) (M, diag.Diagnostics)
}

func (c beforePublishCRUD[M]) create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !c.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	var plan M
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := c.post(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := c.p.client.CreateRule(ctx, c.appID(plan), body)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating resource %s", c.name),
			fmt.Sprintf("Could not create resource %s, unexpected error: %s", c.name, err.Error()),
		)
		return
	}

	responseValues, respDiags := c.response(ctx, &rule, &plan)
	resp.Diagnostics.Append(respDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, responseValues)...)
}

func (c beforePublishCRUD[M]) read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state M
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := c.p.client.GetRule(ctx, c.appID(state), c.id(state))
	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading resource %s", c.name),
			fmt.Sprintf("Could not read resource %s, unexpected error: %s", c.name, err.Error()),
		)
		return
	}

	responseValues, respDiags := c.response(ctx, &rule, &state)
	resp.Diagnostics.Append(respDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &responseValues)...)
}

func (c beforePublishCRUD[M]) update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan M
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := c.post(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := c.p.client.UpdateRule(ctx, c.appID(plan), c.id(plan), body)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating resource %s", c.name),
			fmt.Sprintf("Could not update resource %s, unexpected error: %s", c.name, err.Error()),
		)
		return
	}

	responseValues, respDiags := c.response(ctx, &rule, &plan)
	resp.Diagnostics.Append(respDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &responseValues)...)
}

func (c beforePublishCRUD[M]) delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state M
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := c.p.client.DeleteRule(ctx, c.appID(state), c.id(state))
	if err != nil {
		if is404(err) {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Resource %s does not exist", c.name),
				fmt.Sprintf("Resource %s does not exist, it may have already been deleted: %s", c.name, err.Error()),
			)
		} else {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error deleting resource %s", c.name),
				fmt.Sprintf("Could not delete resource %s, unexpected error: %s", c.name, err.Error()),
			)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

// stripNestedCustomTypes removes the generated CustomType from the named
// single-nested blocks so a plain-struct tfsdk model reflects cleanly. Every
// port needs this (see DEVELOPMENT.md "Porting a resource onto generated
// code"), so it lives here rather than being repeated per resource.
func stripNestedCustomTypes(s *schema.Schema, names ...string) {
	for _, name := range names {
		attr, ok := s.Attributes[name].(schema.SingleNestedAttribute)
		if !ok {
			continue
		}
		attr.CustomType = nil
		s.Attributes[name] = attr
	}
}

// stripSingleNestedCustomType does the same for a single-nested block nested
// inside another single-nested block (the Lambda target's authentication).
func stripSingleNestedCustomType(s *schema.Schema, parent, name string) {
	parentAttr, ok := s.Attributes[parent].(schema.SingleNestedAttribute)
	if !ok {
		return
	}
	childAttr, ok := parentAttr.Attributes[name].(schema.SingleNestedAttribute)
	if !ok {
		return
	}
	childAttr.CustomType = nil
	parentAttr.Attributes[name] = childAttr
	s.Attributes[parent] = parentAttr
}

// stripListNestedCustomType does the same for a list-nested block inside a
// single-nested block (the webhook target's headers), where the CustomType sits
// on the list's nested object.
func stripListNestedCustomType(s *schema.Schema, parent, name string) {
	parentAttr, ok := s.Attributes[parent].(schema.SingleNestedAttribute)
	if !ok {
		return
	}
	listAttr, ok := parentAttr.Attributes[name].(schema.ListNestedAttribute)
	if !ok {
		return
	}
	listAttr.NestedObject.CustomType = nil
	parentAttr.Attributes[name] = listAttr
	s.Attributes[parent] = parentAttr
}

// checkRuleType guards against the API returning a rule of a different family,
// which would otherwise be silently mis-mapped onto this resource's model.
func checkRuleType(rule *control.RuleResponse, want string) diag.Diagnostics {
	var diags diag.Diagnostics
	if rule.RuleType != want {
		diags.AddError(
			"Unexpected rule type in response",
			fmt.Sprintf("Expected rule type %q but received %q", want, rule.RuleType),
		)
	}
	return diags
}
