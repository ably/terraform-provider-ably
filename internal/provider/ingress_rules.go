// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"
	"strings"

	control "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetPlanIngressRule converts an ingress rule from terraform format to control SDK format.
func GetPlanIngressRule(plan AblyIngressRule) control.NewIngressRule {
	var target control.IngressTarget

	switch t := plan.Target.(type) {
	case *AblyIngressRuleTargetMongo:
		target = &control.IngressMongoTarget{
			Url:                      t.Url,
			Database:                 t.Database,
			Collection:               t.Collection,
			Pipeline:                 t.Pipeline,
			FullDocument:             t.FullDocument,
			FullDocumentBeforeChange: t.FullDocumentBeforeChange,
			PrimarySite:              t.PrimarySite,
		}
	case *AblyIngressRuleTargetPostgresOutbox:
		target = &control.IngressPostgresOutboxTarget{
			Url:               t.Url,
			OutboxTableSchema: t.OutboxTableSchema,
			OutboxTableName:   t.OutboxTableName,
			NodesTableSchema:  t.NodesTableSchema,
			NodesTableName:    t.NodesTableName,
			SslMode:           t.SslMode,
			SslRootCert:       t.SslRootCert,
			PrimarySite:       t.PrimarySite,
		}
	}

	ruleValues := control.NewIngressRule{
		Status: plan.Status.ValueString(),
		Target: target,
	}

	return ruleValues
}

// GetIngressRuleResponse maps response body to resource schema attributes.
// Using plan to fill in values that the api does not return.
func GetIngressRuleResponse(ablyIngressRule *control.IngressRule, plan *AblyIngressRule) AblyIngressRule {
	var respTarget any

	switch v := ablyIngressRule.Target.(type) {
	case *control.IngressMongoTarget:
		respTarget = &AblyIngressRuleTargetMongo{
			Url:                      v.Url,
			Database:                 v.Database,
			Collection:               v.Collection,
			Pipeline:                 v.Pipeline,
			FullDocument:             v.FullDocument,
			FullDocumentBeforeChange: v.FullDocumentBeforeChange,
			PrimarySite:              v.PrimarySite,
		}
	case *control.IngressPostgresOutboxTarget:
		respTarget = &AblyIngressRuleTargetPostgresOutbox{
			Url:               v.Url,
			OutboxTableSchema: v.OutboxTableSchema,
			OutboxTableName:   v.OutboxTableName,
			NodesTableSchema:  v.NodesTableSchema,
			NodesTableName:    v.NodesTableName,
			SslMode:           v.SslMode,
			SslRootCert:       v.SslRootCert,
			PrimarySite:       v.PrimarySite,
		}
	}

	respRule := AblyIngressRule{
		ID:     types.StringValue(ablyIngressRule.ID),
		AppID:  types.StringValue(ablyIngressRule.AppID),
		Status: types.StringValue(ablyIngressRule.Status),
		Target: respTarget,
	}

	return respRule
}

func GetIngressRuleSchema(target map[string]schema.Attribute, markdownDescription string) schema.Schema {
	return schema.Schema{
		MarkdownDescription: markdownDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The rule ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_id": schema.StringAttribute{
				Required:    true,
				Description: "The Ably application ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Description: "The status of the rule. Rules can be enabled or disabled.",
			},
			"target": schema.SingleNestedAttribute{
				Required:    true,
				Description: "object (rule_source)",
				Attributes:  target,
			},
		},
	}
}

type IngressRule interface {
	Provider() *AblyProvider
	Name() string
}

// CreateIngressRule creates a new ingress rule resource.
func CreateIngressRule[T any](r IngressRule, ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Checks whether the provider and API Client are configured. If they are not, the provider responds with an error.
	if !r.Provider().configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply",
		)
		return
	}

	// Gets plan values
	var p AblyIngressRuleDecoder[*T]
	diags := req.Plan.Get(ctx, &p)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan := p.IngressRule()
	planValues := GetPlanIngressRule(plan)

	// Creates a new Ably Ingress Rule by invoking the CreateRule function from the Client Library
	ingressRule, err := r.Provider().client.CreateIngressRule(plan.AppID.ValueString(), &planValues)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating Resource '%s'", r.Name()),
			fmt.Sprintf("Could not create resource '%s', unexpected error: %s", r.Name(), err.Error()),
		)

		return
	}

	responseValues := GetIngressRuleResponse(&ingressRule, &plan)

	// Sets state for the new Ably Ingress Rule.
	diags = resp.State.Set(ctx, responseValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ReadIngressRule reads an existing ingress rule resource.
func ReadIngressRule[T any](r IngressRule, ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var s AblyIngressRuleDecoder[*T]
	diags := req.State.Get(ctx, &s)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	state := s.IngressRule()

	// Gets the Ably App ID and Ably Ingress Rule ID value for the resource
	appID := s.AppID.ValueString()
	ingressRuleID := s.ID.ValueString()

	// Get Ingress Rule data
	ingressRule, err := r.Provider().client.IngressRule(appID, ingressRuleID)

	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading Resource %s", r.Name()),
			fmt.Sprintf("Could not read resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	responseValues := GetIngressRuleResponse(&ingressRule, &state)

	// Sets state to app values.
	diags = resp.State.Set(ctx, &responseValues)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// UpdateIngressRule updates an existing ingress rule resource.
func UpdateIngressRule[T any](r IngressRule, ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Gets plan values
	var p AblyIngressRuleDecoder[*T]
	diags := req.Plan.Get(ctx, &p)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	plan := p.IngressRule()

	ruleValues := GetPlanIngressRule(plan)

	// Gets the Ably App ID and Ably Ingress Rule ID value for the resource
	appID := plan.AppID.ValueString()
	ruleID := plan.ID.ValueString()

	// Update Ably Ingress Rule
	ingressRule, err := r.Provider().client.UpdateIngressRule(appID, ruleID, &ruleValues)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating Resource %s", r.Name()),
			fmt.Sprintf("Could not update resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	responseValues := GetIngressRuleResponse(&ingressRule, &plan)

	// Sets state to app values.
	diags = resp.State.Set(ctx, &responseValues)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

// DeleteIngressRule deletes an ingress rule resource.
func DeleteIngressRule[T any](r IngressRule, ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var s AblyIngressRuleDecoder[*T]
	diags := req.State.Get(ctx, &s)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	state := s.IngressRule()

	// Gets the Ably App ID and Ably Rule ID value for the resource
	appID := state.AppID.ValueString()
	ingressRuleID := state.ID.ValueString()

	err := r.Provider().client.DeleteIngressRule(appID, ingressRuleID)
	if err != nil {
		if is404(err) {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Resource does %s not exist", r.Name()),
				fmt.Sprintf("Resource does %s not exist, it may have already been deleted: %s", r.Name(), err.Error()),
			)
		} else {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error deleting Resource %s'", r.Name()),
				fmt.Sprintf("Could not delete resource '%s', unexpected error: %s", r.Name(), err.Error()),
			)
			return
		}
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

// ImportIngressRuleResource handles importing an ingress rule resource.
func ImportIngressRuleResource(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse, fields ...string) {
	// Save the import identifier in the id attribute
	// identifier should be in the format app_id,key_id
	idParts := strings.Split(req.ID, ",")
	anyEmpty := false

	for _, v := range idParts {
		if v == "" {
			anyEmpty = true
		}
	}

	if len(idParts) != len(fields) || anyEmpty {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: '%s'. Got: %q", strings.Join(fields, ","), req.ID),
		)
		return
	}

	for i, v := range fields {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(v), idParts[i])...)
	}
}
