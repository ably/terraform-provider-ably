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
	"fmt"
	"maps"
	"os"
	"reflect"
	"regexp"
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
	"api_key":           true,
	"token":             true,
	"password":          true,
	"secret_access_key": true,
}

// customExpr is a code expression plus the imports it needs, used to emit
// validators, defaults and plan modifiers into the Provider Code Spec.
type customExpr struct {
	imports []string
	expr    string
}

// override augments a generated attribute with metadata the Go types can't
// express. Keyed by snake_case attribute name (top-level or nested).
type override struct {
	mode          string // overrides computed_optional_required when set
	staticDefault any    // sets a static default when non-nil
	allowEmpty    bool   // suppresses the LengthAtLeast(1) validator
	validators    []customExpr
	planModifiers []customExpr
}

const (
	pkgStringValidator    = "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	pkgInt64Validator     = "github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	pkgStringPlanModifier = "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	pkgPlanModifiers      = "github.com/ably/terraform-provider-ably/internal/provider/planmodifiers"
	pkgRegexp             = "regexp"
)

// attrOverrides supplies the generic, spec-absent metadata: the envelope plan
// modifiers and the rule status field (which the spec doesn't enumerate).
// Field-level enums, patterns and their defaults are sourced from the spec
// per rule instead (see the string case in attrsFromStruct), so they stay
// correct per family, e.g. moderation rules enforce invocation_mode
// BEFORE_PUBLISH while hive/dashboard enforces AFTER_PUBLISH.
var attrOverrides = map[string]override{
	"id":     {planModifiers: []customExpr{{[]string{pkgStringPlanModifier}, "stringplanmodifier.UseStateForUnknown()"}}},
	"app_id": {planModifiers: []customExpr{{[]string{pkgStringPlanModifier}, "stringplanmodifier.RequiresReplace()"}}},
	"status": {
		mode:          "computed_optional",
		staticDefault: "enabled",
		validators:    []customExpr{{[]string{pkgStringValidator}, `stringvalidator.OneOf("enabled", "disabled")`}},
	},
	// The rule PATCH schemas neither accept null nor let the pattern-bound
	// chatRoomFilter be "", so an in-place update can never unset it (verified
	// against the live API, 2026-07-08): removing it must recreate the rule.
	"chat_room_filter": {planModifiers: []customExpr{{[]string{pkgPlanModifiers}, "planmodifiers.RequiresReplaceWhenCleared()"}}},
	// "" is a meaningful value for a source channel filter: the spec documents
	// it as "apply to all channels", and unlike chatRoomFilter it round-trips
	// (verified against the live API, 2026-07-20: create with "" and update
	// non-empty -> "" both persist and read back as "").
	"channel_filter": {allowEmpty: true},
}

// applyOverride mutates an attribute's type map with any configured metadata.
func applyOverride(name string, m map[string]any) {
	ov, ok := attrOverrides[name]
	if !ok {
		return
	}
	if ov.mode != "" {
		m["computed_optional_required"] = ov.mode
	}
	if ov.staticDefault != nil {
		m["default"] = map[string]any{"static": ov.staticDefault}
	}
	if len(ov.validators) > 0 {
		m["validators"] = customList(ov.validators)
	}
	if len(ov.planModifiers) > 0 {
		m["plan_modifiers"] = customList(ov.planModifiers)
	}
}

func customList(exprs []customExpr) []map[string]any {
	out := make([]map[string]any, 0, len(exprs))
	for _, e := range exprs {
		imports := make([]map[string]any, 0, len(e.imports))
		for _, p := range e.imports {
			imports = append(imports, map[string]any{"path": p})
		}
		out = append(out, map[string]any{
			"custom": map[string]any{
				"imports":           imports,
				"schema_definition": e.expr,
			},
		})
	}
	return out
}

// specSchemas holds components.schemas from the vendored spec so property
// lookups can resolve $ref and oneOf nodes anywhere in the tree.
var specSchemas map[string]any

func main() {
	schemas := loadSpecSchemas("codegen/control-api.yaml")
	specSchemas = schemas

	resources := make([]map[string]any, 0, len(rules))
	for _, r := range rules {
		props := schemaProps(schemas, r.specSchema)
		attrs := attrsFromStruct(reflect.TypeOf(r.post), props)
		// Every rule resource carries the same envelope: a computed id and the
		// required parent app_id. These are not on the create body.
		idMap := map[string]any{"computed_optional_required": "computed", "description": "The rule ID."}
		applyOverride("id", idMap)
		appIDMap := map[string]any{"computed_optional_required": "required", "description": "The Ably application ID."}
		applyOverride("app_id", appIDMap)
		envelope := []map[string]any{
			{"name": "id", "string": idMap},
			{"name": "app_id", "string": appIDMap},
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

// schemaProps returns the properties map of a named schema. A missing or
// property-less schema fails generation immediately: returning nil here would
// silently strip descriptions, enums, patterns and defaults from the whole
// resource next time someone renames a schema in the vendored spec.
func schemaProps(schemas map[string]any, name string) map[string]any {
	props := asMap(asMap(schemas[name])["properties"])
	if len(props) == 0 {
		panic(fmt.Sprintf("spec schema %q is missing or has no properties: the rules table in ruletypesgen and the vendored spec (codegen/control-api.yaml) are out of sync", name))
	}
	return props
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
			applyOverride(name, s)
			// Source enum and pattern constraints from the spec so they stay
			// correct per rule. A single-value enum (e.g. invocation_mode) is
			// also made computed_optional with that value as the default.
			var specVals []customExpr
			if enums := specEnum(props, jsonName); len(enums) > 0 {
				specVals = append(specVals, customExpr{[]string{pkgStringValidator}, oneOfExpr(enums)})
				if len(enums) == 1 {
					s["computed_optional_required"] = "computed_optional"
					s["default"] = map[string]any{"static": enums[0]}
				}
			}
			if p := specPattern(props, jsonName); p != "" {
				specVals = append(specVals, customExpr{[]string{pkgStringValidator, pkgRegexp}, regexExpr(p, jsonName)})
			}
			// Reject explicit "" on string attributes: empty and unset mean the
			// same thing to the Control API, and a known "" in the plan reads
			// back as null, aborting the apply with an opaque "inconsistent
			// values" error. A plan-time validator turns that into a clear
			// message. Enum-valued attributes already exclude "" via OneOf, and
			// attributes for which "" is a real value opt out via allowEmpty.
			if len(specEnum(props, jsonName)) == 0 && len(attrOverrides[name].validators) == 0 && !attrOverrides[name].allowEmpty {
				specVals = append(specVals, customExpr{[]string{pkgStringValidator}, "stringvalidator.LengthAtLeast(1)"})
			}
			if len(specVals) > 0 {
				existing, _ := s["validators"].([]map[string]any)
				s["validators"] = append(existing, customList(specVals)...)
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
			// Source minimum/maximum bounds from the spec, mirroring what the
			// string case does for enums and patterns, so out-of-range values
			// fail at plan time instead of only at the real API (the echoing
			// fake accepts anything, so nothing else catches them).
			minV, hasMin := specBound(props, jsonName, "minimum")
			maxV, hasMax := specBound(props, jsonName, "maximum")
			var expr string
			switch {
			case hasMin && hasMax:
				expr = fmt.Sprintf("int64validator.Between(%d, %d)", minV, maxV)
			case hasMin:
				expr = fmt.Sprintf("int64validator.AtLeast(%d)", minV)
			case hasMax:
				expr = fmt.Sprintf("int64validator.AtMost(%d)", maxV)
			}
			if expr != "" {
				n["validators"] = customList([]customExpr{{[]string{pkgInt64Validator}, expr}})
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
		case reflect.Map:
			if ft.Key().Kind() != reflect.String {
				panic(fmt.Sprintf("ruletypesgen: %s.%s (%s) has non-string map key %s; the generator cannot emit it", t.Name(), f.Name, jsonName, ft.Key().Kind()))
			}
			m := map[string]any{
				"computed_optional_required": mode,
				"element_type":               map[string]any{elementType(ft.Elem()): map[string]any{}},
			}
			if desc != "" {
				m["description"] = desc
			}
			attr["map"] = m
		default:
			// Fail loudly rather than emitting an incomplete schema: a silent
			// skip here is invisible until a user discovers a field their rule
			// supports doesn't exist in the provider (this happened to the
			// moderation thresholds maps).
			panic(fmt.Sprintf("ruletypesgen: %s.%s (%s) has kind %s which the generator cannot emit; add support or exclude it explicitly", t.Name(), f.Name, jsonName, ft.Kind()))
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

// resolveNode follows $ref and oneOf on a schema node so lookups see the
// effective schema. A $ref resolves to the named component (failing loudly if
// it doesn't exist); a oneOf resolves each variant and merges their
// properties, unioning enums per property, so a discriminated union (e.g. the
// lambda target's aws_access_keys/aws_assume_role authentication) yields one
// property map whose discriminator carries the enum of all variants. Nodes
// without either pass through unchanged.
func resolveNode(node map[string]any) map[string]any {
	if ref, ok := node["$ref"].(string); ok {
		name := ref[strings.LastIndex(ref, "/")+1:]
		target := asMap(specSchemas[name])
		if target == nil {
			panic(fmt.Sprintf("unresolvable $ref %q in the vendored spec", ref))
		}
		return resolveNode(target)
	}
	variants, ok := node["oneOf"].([]any)
	if !ok {
		return node
	}
	merged := map[string]any{}
	for _, v := range variants {
		mergeProps(merged, asMap(resolveNode(asMap(v))["properties"]))
	}
	out := map[string]any{"properties": merged}
	if d, ok := node["description"].(string); ok {
		out["description"] = d
	}
	return out
}

// mergeProps merges src properties into dst: new properties are referenced
// as-is, existing ones union their enum values (in encounter order) and keep
// the first description seen. A property is never mutated in place — the
// values alias the parsed spec, shared by every schema that references them —
// so a merged property is a fresh copy.
func mergeProps(dst, src map[string]any) {
	for name, v := range src {
		existing, ok := dst[name]
		if !ok {
			dst[name] = v
			continue
		}
		vEnum, _ := asMap(v)["enum"].([]any)
		if len(vEnum) == 0 {
			continue
		}
		em := asMap(existing)
		merged := make(map[string]any, len(em))
		maps.Copy(merged, em)
		eEnum, _ := em["enum"].([]any)
		combined := append([]any{}, eEnum...)
		seen := map[any]bool{}
		for _, e := range combined {
			seen[e] = true
		}
		for _, e := range vEnum {
			if !seen[e] {
				combined = append(combined, e)
			}
		}
		merged["enum"] = combined
		dst[name] = merged
	}
}

// description returns the description of a property by its JSON name,
// following $ref/oneOf to the effective schema when the property itself has
// no inline description.
func description(props map[string]any, jsonName string) string {
	node := asMap(props[jsonName])
	if s, ok := node["description"].(string); ok {
		return s
	}
	s, _ := resolveNode(node)["description"].(string)
	return s
}

// childProps returns the nested properties map of an object-typed property,
// resolving $ref and oneOf indirection.
func childProps(props map[string]any, jsonName string) map[string]any {
	return asMap(resolveNode(asMap(props[jsonName]))["properties"])
}

// itemProps returns the properties map of an array property's item schema,
// resolving $ref and oneOf indirection.
func itemProps(props map[string]any, jsonName string) map[string]any {
	return asMap(resolveNode(asMap(asMap(props[jsonName])["items"]))["properties"])
}

// specEnum returns the string enum values declared for a property, if any.
func specEnum(props map[string]any, jsonName string) []string {
	raw, _ := asMap(props[jsonName])["enum"].([]any)
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// specPattern returns the regex pattern declared for a property, if any.
func specPattern(props map[string]any, jsonName string) string {
	s, _ := asMap(props[jsonName])["pattern"].(string)
	return s
}

// specBound returns an integer bound (minimum/maximum) declared for a
// property. YAML numbers decode as int or float64 depending on their form.
func specBound(props map[string]any, jsonName, key string) (int64, bool) {
	switch v := asMap(props[jsonName])[key].(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	}
	return 0, false
}

// oneOfExpr builds a stringvalidator.OneOf(...) expression for the given values.
func oneOfExpr(vals []string) string {
	quoted := make([]string, len(vals))
	for i, v := range vals {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	return "stringvalidator.OneOf(" + strings.Join(quoted, ", ") + ")"
}

// regexExpr builds a stringvalidator.RegexMatches(...) expression for a
// pattern, compiling it first so a malformed pattern in the vendored spec
// fails generation here rather than panicking the generated provider at init.
func regexExpr(pattern, jsonName string) string {
	if _, err := regexp.Compile(pattern); err != nil {
		panic(fmt.Sprintf("invalid regex pattern %q on spec property %q: %v", pattern, jsonName, err))
	}
	return fmt.Sprintf("stringvalidator.RegexMatches(regexp.MustCompile(%q), %q)", pattern, "must match the pattern "+pattern)
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
