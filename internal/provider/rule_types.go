// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// RuleTypeResources maps a Control API ruleType onto the Terraform resource type
// that manages it.
//
// Every integration rule comes back from GET /apps/{id}/rules regardless of
// family: webhooks, firehoses, ingress and moderation rules all arrive in one
// list, told apart only by ruleType. Anything walking an account needs this
// mapping to know which resource owns a rule, and it is otherwise implicit in the
// type switches in rules.go and ingress_rules.go.
//
// Adding a rule resource means adding a line here. TestRuleTypeResources fails if
// you forget, or if this table names a resource the provider doesn't register.
var RuleTypeResources = map[string]string{
	"aws/kinesis":                "ably_rule_kinesis",
	"aws/sqs":                    "ably_rule_sqs",
	"aws/lambda":                 "ably_rule_lambda",
	"pulsar":                     "ably_rule_pulsar",
	"kafka":                      "ably_rule_kafka",
	"amqp":                       "ably_rule_amqp",
	"amqp/external":              "ably_rule_amqp_external",
	"http":                       "ably_rule_http",
	"http/zapier":                "ably_rule_zapier",
	"http/cloudflare-worker":     "ably_rule_cloudflare_worker",
	"http/azure-function":        "ably_rule_azure_function",
	"http/google-cloud-function": "ably_rule_google_function",
	"http/ifttt":                 "ably_rule_ifttt",
	"bodyguard/text-moderation":  "ably_rule_bodyguard",
	"tisane/text-moderation":     "ably_rule_tisane",
	"azure/text-moderation":      "ably_rule_azure_moderation",
	"hive/text-model-only":       "ably_rule_hive_text",
	"hive/dashboard":             "ably_rule_hive_dashboard",
	"http/before-publish":        "ably_rule_before_publish_webhook",
	"aws/lambda/before-publish":  "ably_rule_before_publish_lambda",
	"ingress/mongodb":            "ably_ingress_rule_mongodb",
	"ingress-postgres-outbox":    "ably_ingress_rule_postgres_outbox",
}

// ResourceTypeNames returns the sorted type name of every resource the provider
// registers. It asks each resource for its Metadata, so the list can't drift from
// what the provider serves.
func ResourceTypeNames(ctx context.Context) []string {
	p := &AblyProvider{}
	var names []string
	for _, newResource := range p.Resources(ctx) {
		var resp resource.MetadataResponse
		newResource().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "ably"}, &resp)
		names = append(names, resp.TypeName)
	}
	sort.Strings(names)
	return names
}

// IsRuleResourceType reports whether a resource type manages an integration rule,
// and so needs an entry in RuleTypeResources. Those are named ably_rule_* or
// ably_ingress_rule_*.
func IsRuleResourceType(typeName string) bool {
	return strings.Contains(typeName, "_rule_")
}
