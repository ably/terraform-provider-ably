// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetPlanIngressRule converts an ingress rule from terraform format to the control SDK format.
func GetPlanIngressRule(plan AblyIngressRule) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	status := plan.Status.ValueString()

	switch t := plan.Target.(type) {
	case *AblyIngressRuleTargetMongo:
		return control.IngressMongoDBRulePost{
			Status:   status,
			RuleType: "ingress/mongodb",
			Target: control.IngressMongoDBTarget{
				URL:                      t.Url.ValueString(),
				Database:                 t.Database.ValueString(),
				Collection:               t.Collection.ValueString(),
				Pipeline:                 t.Pipeline.ValueString(),
				FullDocument:             t.FullDocument.ValueString(),
				FullDocumentBeforeChange: t.FullDocumentBeforeChange.ValueString(),
				PrimarySite:              t.PrimarySite.ValueString(),
			},
		}, diags
	case *AblyIngressRuleTargetPostgresOutbox:
		var sslRootCert *string
		if !t.SslRootCert.IsNull() && !t.SslRootCert.IsUnknown() {
			v := t.SslRootCert.ValueString()
			sslRootCert = &v
		}
		return control.IngressPostgresOutboxRulePost{
			Status:   status,
			RuleType: "ingress-postgres-outbox",
			Target: control.IngressPostgresOutboxTarget{
				URL:               t.Url.ValueString(),
				OutboxTableSchema: t.OutboxTableSchema.ValueString(),
				OutboxTableName:   t.OutboxTableName.ValueString(),
				NodesTableSchema:  t.NodesTableSchema.ValueString(),
				NodesTableName:    t.NodesTableName.ValueString(),
				SSLMode:           t.SslMode.ValueString(),
				SSLRootCert:       sslRootCert,
				PrimarySite:       t.PrimarySite.ValueString(),
			},
		}, diags
	}

	diags.AddError(
		"Unrecognized ingress rule target type",
		fmt.Sprintf("The plan contains an unrecognized ingress rule target type: %T", plan.Target),
	)
	return nil, diags
}

// GetIngressRuleResponse maps an API rule response to the ingress rule terraform model.
// Ingress rules use the same generic RuleResponse from the client, with target unmarshalled
// according to the ruleType.
func GetIngressRuleResponse(ablyRule *control.RuleResponse, plan *AblyIngressRule, reading bool) (AblyIngressRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	rc := newReconciler(&diags)
	if reading {
		rc.forRead()
	}

	var respTarget any

	switch ablyRule.RuleType {
	case "ingress/mongodb":
		target, err := unmarshalTarget[control.IngressMongoDBTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling ingress rule target", fmt.Sprintf("Could not unmarshal ingress/mongodb target: %s", err.Error()))
			return AblyIngressRule{}, diags
		}
		var pt *AblyIngressRuleTargetMongo
		if p, ok := plan.Target.(*AblyIngressRuleTargetMongo); ok {
			pt = p
		}
		respTarget = &AblyIngressRuleTargetMongo{
			Url:                      rc.str("target.url", planStr(pt, func(t *AblyIngressRuleTargetMongo) types.String { return t.Url }), types.StringValue(target.URL), false),
			Database:                 rc.str("target.database", planStr(pt, func(t *AblyIngressRuleTargetMongo) types.String { return t.Database }), types.StringValue(target.Database), false),
			Collection:               rc.str("target.collection", planStr(pt, func(t *AblyIngressRuleTargetMongo) types.String { return t.Collection }), types.StringValue(target.Collection), false),
			Pipeline:                 rc.str("target.pipeline", planStr(pt, func(t *AblyIngressRuleTargetMongo) types.String { return t.Pipeline }), types.StringValue(target.Pipeline), false),
			FullDocument:             rc.str("target.full_document", planStr(pt, func(t *AblyIngressRuleTargetMongo) types.String { return t.FullDocument }), types.StringValue(target.FullDocument), false),
			FullDocumentBeforeChange: rc.str("target.full_document_before_change", planStr(pt, func(t *AblyIngressRuleTargetMongo) types.String { return t.FullDocumentBeforeChange }), types.StringValue(target.FullDocumentBeforeChange), false),
			PrimarySite:              rc.str("target.primary_site", planStr(pt, func(t *AblyIngressRuleTargetMongo) types.String { return t.PrimarySite }), types.StringValue(target.PrimarySite), false),
		}
	case "ingress-postgres-outbox":
		target, err := unmarshalTarget[control.IngressPostgresOutboxTarget](ablyRule.Target)
		if err != nil {
			diags.AddError("Error unmarshalling ingress rule target", fmt.Sprintf("Could not unmarshal ingress-postgres-outbox target: %s", err.Error()))
			return AblyIngressRule{}, diags
		}
		var pt *AblyIngressRuleTargetPostgresOutbox
		if p, ok := plan.Target.(*AblyIngressRuleTargetPostgresOutbox); ok {
			pt = p
		}
		respTarget = &AblyIngressRuleTargetPostgresOutbox{
			Url:               rc.str("target.url", planStr(pt, func(t *AblyIngressRuleTargetPostgresOutbox) types.String { return t.Url }), types.StringValue(target.URL), false),
			OutboxTableSchema: rc.str("target.outbox_table_schema", planStr(pt, func(t *AblyIngressRuleTargetPostgresOutbox) types.String { return t.OutboxTableSchema }), types.StringValue(target.OutboxTableSchema), false),
			OutboxTableName:   rc.str("target.outbox_table_name", planStr(pt, func(t *AblyIngressRuleTargetPostgresOutbox) types.String { return t.OutboxTableName }), types.StringValue(target.OutboxTableName), false),
			NodesTableSchema:  rc.str("target.nodes_table_schema", planStr(pt, func(t *AblyIngressRuleTargetPostgresOutbox) types.String { return t.NodesTableSchema }), types.StringValue(target.NodesTableSchema), false),
			NodesTableName:    rc.str("target.nodes_table_name", planStr(pt, func(t *AblyIngressRuleTargetPostgresOutbox) types.String { return t.NodesTableName }), types.StringValue(target.NodesTableName), false),
			SslMode:           rc.str("target.ssl_mode", planStr(pt, func(t *AblyIngressRuleTargetPostgresOutbox) types.String { return t.SslMode }), types.StringValue(target.SSLMode), false),
			SslRootCert:       rc.str("target.ssl_root_cert", planStr(pt, func(t *AblyIngressRuleTargetPostgresOutbox) types.String { return t.SslRootCert }), optStringValue(target.SSLRootCert), false),
			PrimarySite:       rc.str("target.primary_site", planStr(pt, func(t *AblyIngressRuleTargetPostgresOutbox) types.String { return t.PrimarySite }), types.StringValue(target.PrimarySite), false),
		}
	default:
		diags.AddError(
			"Unknown ingress rule type in response",
			fmt.Sprintf("Received unrecognized ingress rule type from API: %q", ablyRule.RuleType),
		)
		return AblyIngressRule{}, diags
	}

	respRule := AblyIngressRule{
		ID:     rc.str("id", plan.ID, types.StringValue(ablyRule.ID), true),
		AppID:  rc.str("app_id", plan.AppID, types.StringValue(ablyRule.AppID), false),
		Status: rc.str("status", plan.Status, types.StringValue(ablyRule.Status), true),
		Target: respTarget,
	}

	return respRule, diags
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
				Computed:    true,
				Description: "The status of the rule. Rules can be enabled or disabled.",
				Default:     stringdefault.StaticString("enabled"),
				Validators: []validator.String{
					stringvalidator.OneOf("enabled", "disabled"),
				},
			},
			"target": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The target for the ingress rule",
				Attributes:  target,
			},
		},
	}
}

// CreateIngressRule creates a new ingress rule resource.
func CreateIngressRule[T any](r Rule, ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.Provider().ensureConfigured(&resp.Diagnostics) {
		return
	}

	var p AblyIngressRuleDecoder[*T]
	diags := req.Plan.Get(ctx, &p)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan := p.IngressRule()
	planValues, planDiags := GetPlanIngressRule(plan)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.Provider().client.CreateRule(ctx, plan.AppID.ValueString(), planValues)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating resource %s", r.Name()),
			fmt.Sprintf("Could not create resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	responseValues, respDiags := GetIngressRuleResponse(&rule, &plan, false)
	resp.Diagnostics.Append(respDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, responseValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// ReadIngressRule reads an existing ingress rule resource.
func ReadIngressRule[T any](r Rule, ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.Provider().ensureConfigured(&resp.Diagnostics) {
		return
	}

	var s AblyIngressRuleDecoder[*T]
	diags := req.State.Get(ctx, &s)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := s.IngressRule()

	appID := s.AppID.ValueString()
	ruleID := s.ID.ValueString()

	rule, err := r.Provider().client.GetRule(ctx, appID, ruleID)
	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading resource %s", r.Name()),
			fmt.Sprintf("Could not read resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	responseValues, respDiags := GetIngressRuleResponse(&rule, &state, true)
	resp.Diagnostics.Append(respDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &responseValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// UpdateIngressRule updates an existing ingress rule resource.
func UpdateIngressRule[T any](r Rule, ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.Provider().ensureConfigured(&resp.Diagnostics) {
		return
	}

	var p AblyIngressRuleDecoder[*T]
	diags := req.Plan.Get(ctx, &p)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan := p.IngressRule()

	ruleValues, planDiags := GetPlanIngressRule(plan)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := plan.AppID.ValueString()
	ruleID := plan.ID.ValueString()

	rule, err := r.Provider().client.UpdateRule(ctx, appID, ruleID, ruleValues)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating resource %s", r.Name()),
			fmt.Sprintf("Could not update resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	responseValues, respDiags := GetIngressRuleResponse(&rule, &plan, false)
	resp.Diagnostics.Append(respDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &responseValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// DeleteIngressRule deletes an ingress rule resource.
func DeleteIngressRule[T any](r Rule, ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.Provider().ensureConfigured(&resp.Diagnostics) {
		return
	}

	var s AblyIngressRuleDecoder[*T]
	diags := req.State.Get(ctx, &s)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := s.IngressRule()

	appID := state.AppID.ValueString()
	ruleID := state.ID.ValueString()

	err := r.Provider().client.DeleteRule(ctx, appID, ruleID)
	if err != nil {
		if is404(err) {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Resource %s does not exist", r.Name()),
				fmt.Sprintf("Resource %s does not exist, it may have already been deleted: %s", r.Name(), err.Error()),
			)
		} else {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error deleting resource %s", r.Name()),
				fmt.Sprintf("Could not delete resource %s, unexpected error: %s", r.Name(), err.Error()),
			)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}
