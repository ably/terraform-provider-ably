// Command ruletypesgen emits a Terraform Provider Code Specification for the
// integration-rule families that cannot be generated from the OpenAPI spec.
//
// The Control API models rules as a oneOf + discriminator union, which
// tfplugingen-openapi cannot handle (see CODEGEN_STRATEGY.md). Instead we drive
// generation from the in-repo control rule types, which are already the
// curated, per-family-correct model: the moderation and before-publish
// families correctly drop the webhook source/request_mode fields and carry
// before_publish_config/invocation_mode/chat_room_filter instead.
//
// This program reflects over each rule's XxxRulePost struct and writes a
// Provider Code Spec JSON that tfplugingen-framework then turns into schema and
// model code. Run via `make generate-rules`.
package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"unicode"

	"github.com/ably/terraform-provider-ably/control"
)

// rule pairs a Terraform resource name with a zero value of its create-body
// struct. The struct's fields, reflected below, define the resource's schema.
type rule struct {
	name string
	post any
}

// The families the OpenAPI generator can't produce. Each is driven from its
// control create-body type.
var rules = []rule{
	{"rule_bodyguard", control.BodyguardTextModerationRulePost{}},
	{"rule_tisane", control.TisaneTextModerationRulePost{}},
	{"rule_azure_moderation", control.AzureTextModerationRulePost{}},
	{"rule_hive_text", control.HiveTextModelOnlyRulePost{}},
	{"rule_hive_dashboard", control.HiveDashboardRulePost{}},
	{"rule_before_publish_webhook", control.BeforePublishWebhookRulePost{}},
	{"rule_before_publish_lambda", control.BeforePublishAWSLambdaRulePost{}},
}

// sensitive field names (by snake_case) that should be marked Sensitive.
var sensitive = map[string]bool{
	"api_key":  true,
	"token":    true,
	"password": true,
}

func main() {
	resources := make([]map[string]any, 0, len(rules))
	for _, r := range rules {
		t := reflect.TypeOf(r.post)
		attrs := attrsFromStruct(t)
		// Every rule resource carries the same envelope: a computed id and the
		// required parent app_id. These are not on the create body.
		envelope := []map[string]any{
			{"name": "id", "string": map[string]any{
				"computed_optional_required": "computed",
				"description":                "The rule ID.",
			}},
			{"name": "app_id", "string": map[string]any{
				"computed_optional_required": "required",
				"description":                "The Ably application ID.",
			}},
		}
		attrs = append(envelope, attrs...)
		resources = append(resources, map[string]any{
			"name":   r.name,
			"schema": map[string]any{"attributes": attrs},
		})
	}

	spec := map[string]any{
		"provider":  map[string]any{"name": "ably"},
		"resources": resources,
		"version":   "0.1",
	}

	out, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("codegen/rules_spec.json", append(out, '\n'), 0o644); err != nil {
		panic(err)
	}
}

// attrsFromStruct reflects a struct type into Provider Code Spec attributes.
func attrsFromStruct(t reflect.Type) []map[string]any {
	var attrs []map[string]any
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		jsonName, omitempty := parseJSONTag(f)
		name := snake(jsonName)
		// ruleType is the discriminator, fixed per resource and not user facing.
		if name == "rule_type" {
			continue
		}

		ft := f.Type
		optional := omitempty
		if ft.Kind() == reflect.Pointer {
			optional = true
			ft = ft.Elem()
		}
		mode := "required"
		if optional {
			mode = "optional"
		}

		attr := map[string]any{"name": name}
		switch ft.Kind() {
		case reflect.String:
			s := map[string]any{"computed_optional_required": mode}
			if sensitive[name] {
				s["sensitive"] = true
			}
			attr["string"] = s
		case reflect.Bool:
			attr["bool"] = map[string]any{"computed_optional_required": mode}
		case reflect.Int, reflect.Int64, reflect.Int32:
			attr["int64"] = map[string]any{"computed_optional_required": mode}
		case reflect.Struct:
			attr["single_nested"] = map[string]any{
				"computed_optional_required": mode,
				"attributes":                 attrsFromStruct(ft),
			}
		case reflect.Slice:
			elem := ft.Elem()
			if elem.Kind() == reflect.Struct {
				attr["list_nested"] = map[string]any{
					"computed_optional_required": mode,
					"nested_object":              map[string]any{"attributes": attrsFromStruct(elem)},
				}
			} else {
				attr["list"] = map[string]any{
					"computed_optional_required": mode,
					"element_type":               map[string]any{elementType(elem): map[string]any{}},
				}
			}
		default:
			// maps and anything else are skipped; the families here don't use them.
			continue
		}
		attrs = append(attrs, attr)
	}
	return attrs
}

func parseJSONTag(f reflect.StructField) (name string, omitempty bool) {
	tag := f.Tag.Get("json")
	parts := strings.Split(tag, ",")
	name = parts[0]
	if name == "" || name == "-" {
		name = f.Name
	}
	for _, p := range parts[1:] {
		if p == "omitempty" {
			omitempty = true
		}
	}
	return name, omitempty
}

func elementType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int64, reflect.Int32:
		return "int64"
	default:
		return "string"
	}
}

// snake converts a camelCase JSON property name to snake_case.
func snake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
