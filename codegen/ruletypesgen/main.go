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
// Structure (the attribute tree, types, optionality, sensitivity) comes from
// reflecting each rule's XxxRulePost struct. Field descriptions come from the
// vendored OpenAPI spec (codegen/control-api.yaml), looked up by JSON property
// name, because the Go structs carry no documentation. The result is a Provider
// Code Spec JSON that tfplugingen-framework turns into schema and model code.
// Run via `make generate`.
package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"unicode"

	"github.com/ably/terraform-provider-ably/control"
	"gopkg.in/yaml.v3"
)

// rule pairs a Terraform resource name with a zero value of its create-body
// struct and the OpenAPI schema name that documents it.
type rule struct {
	name       string
	post       any
	specSchema string
}

// The families the OpenAPI generator can't produce. Each is driven from its
// control create-body type, with descriptions sourced from the matching spec
// schema.
var rules = []rule{
	{"rule_bodyguard", control.BodyguardTextModerationRulePost{}, "bodyguard_text_moderation_rule_post"},
	{"rule_tisane", control.TisaneTextModerationRulePost{}, "tisane_text_moderation_rule_post"},
	{"rule_azure_moderation", control.AzureTextModerationRulePost{}, "azure_text_moderation_rule_post"},
	{"rule_hive_text", control.HiveTextModelOnlyRulePost{}, "hive_text_model_only_rule_post"},
	{"rule_hive_dashboard", control.HiveDashboardRulePost{}, "hive_dashboard_rule_post"},
	{"rule_before_publish_webhook", control.BeforePublishWebhookRulePost{}, "before_publish_webhook_rule_post"},
	{"rule_before_publish_lambda", control.BeforePublishAWSLambdaRulePost{}, "before_publish_aws_lambda_rule_post"},
}

// sensitive field names (by snake_case) that should be marked Sensitive.
var sensitive = map[string]bool{
	"api_key":  true,
	"token":    true,
	"password": true,
}

func main() {
	schemas := loadSpecSchemas("codegen/control-api.yaml")

	resources := make([]map[string]any, 0, len(rules))
	for _, r := range rules {
		props := schemaProps(schemas, r.specSchema)
		attrs := attrsFromStruct(reflect.TypeOf(r.post), props)
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

// loadSpecSchemas reads the vendored OpenAPI spec and returns components.schemas.
func loadSpecSchemas(path string) map[string]any {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		panic(err)
	}
	return asMap(asMap(doc["components"])["schemas"])
}

// schemaProps returns the properties map of a named schema, or nil.
func schemaProps(schemas map[string]any, name string) map[string]any {
	return asMap(asMap(schemas[name])["properties"])
}

// attrsFromStruct reflects a struct type into Provider Code Spec attributes,
// pulling each field's description from the matching OpenAPI properties map.
func attrsFromStruct(t reflect.Type, props map[string]any) []map[string]any {
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
		desc := description(props, jsonName)

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
			if desc != "" {
				s["description"] = desc
			}
			if sensitive[name] {
				s["sensitive"] = true
			}
			attr["string"] = s
		case reflect.Bool:
			b := map[string]any{"computed_optional_required": mode}
			if desc != "" {
				b["description"] = desc
			}
			attr["bool"] = b
		case reflect.Int, reflect.Int64, reflect.Int32:
			n := map[string]any{"computed_optional_required": mode}
			if desc != "" {
				n["description"] = desc
			}
			attr["int64"] = n
		case reflect.Struct:
			sn := map[string]any{
				"computed_optional_required": mode,
				"attributes":                 attrsFromStruct(ft, childProps(props, jsonName)),
			}
			if desc != "" {
				sn["description"] = desc
			}
			attr["single_nested"] = sn
		case reflect.Slice:
			elem := ft.Elem()
			if elem.Kind() == reflect.Struct {
				attr["list_nested"] = map[string]any{
					"computed_optional_required": mode,
					"nested_object":              map[string]any{"attributes": attrsFromStruct(elem, itemProps(props, jsonName))},
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

// --- OpenAPI properties helpers --------------------------------------------

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

// description returns the description of a property by its JSON name.
func description(props map[string]any, jsonName string) string {
	s, _ := asMap(props[jsonName])["description"].(string)
	return s
}

// childProps returns the nested properties map of an object-typed property.
func childProps(props map[string]any, jsonName string) map[string]any {
	return asMap(asMap(props[jsonName])["properties"])
}

// itemProps returns the properties map of an array property's item schema.
func itemProps(props map[string]any, jsonName string) map[string]any {
	return asMap(asMap(asMap(props[jsonName])["items"])["properties"])
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
