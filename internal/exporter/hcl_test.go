package exporter

import (
	"math/big"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestQuoteString(t *testing.T) {
	for _, test := range []struct {
		name  string
		value string
		want  string
	}{
		{"plain", `hello`, `"hello"`},
		{"quotes", `say "hi"`, `"say \"hi\""`},
		{"backslash", `a\b`, `"a\\b"`},
		{"newline", "line\nline", `"line\nline"`},
		{"tab", "a\tb", `"a\tb"`},
		// An unescaped ${ would be read as an interpolation, and a Control API
		// value that happens to contain one must survive the round trip.
		{"interpolation", `${var.x}`, `"$${var.x}"`},
		{"directive", `%{if true}`, `"%%{if true}"`},
		{"lone dollar", `100$`, `"100$"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := quoteString(test.value); got != test.want {
				t.Errorf("quoteString(%q) = %s, want %s", test.value, got, test.want)
			}
		})
	}
}

// TestQuoteStringProducesParseableHCL checks the escaping against a real HCL
// parser, not just the expectations above.
func TestQuoteStringProducesParseableHCL(t *testing.T) {
	for _, value := range []string{
		`plain`,
		`with "quotes"`,
		`with\backslash`,
		"with\nnewline",
		`${var.interpolation}`,
		`%{if directive}`,
		`regex ^chat:[a-z]+$`,
	} {
		source := "attribute = " + quoteString(value) + "\n"
		parser := hclparse.NewParser()
		file, diagnostics := parser.ParseHCL([]byte(source), "test.tf")
		if diagnostics.HasErrors() {
			t.Errorf("%q rendered unparseable HCL %s: %s", value, source, diagnostics.Error())
			continue
		}

		attributes, diagnostics := file.Body.JustAttributes()
		if diagnostics.HasErrors() {
			t.Errorf("%q: %s", value, diagnostics.Error())
			continue
		}
		got, diagnostics := attributes["attribute"].Expr.Value(nil)
		if diagnostics.HasErrors() {
			t.Errorf("%q did not evaluate: %s", value, diagnostics.Error())
			continue
		}
		if got.AsString() != value {
			t.Errorf("round trip changed %q into %q", value, got.AsString())
		}
	}
}

func TestFormatNumber(t *testing.T) {
	for _, test := range []struct {
		value *big.Float
		want  string
	}{
		{big.NewFloat(0), "0"},
		{big.NewFloat(60), "60"},
		{big.NewFloat(10000), "10000"},
		{big.NewFloat(1.5), "1.5"},
		{big.NewFloat(-3), "-3"},
	} {
		if got := formatNumber(test.value); got != test.want {
			t.Errorf("formatNumber(%s) = %s, want %s", test.value.String(), got, test.want)
		}
	}
}

func TestSanitiseLabel(t *testing.T) {
	for _, test := range []struct {
		name string
		want string
	}{
		{"Chat Service", "chat_service"},
		{"my-app", "my_app"},
		{"  spaced  ", "spaced"},
		{"UPPER", "upper"},
		{"lots???of___junk", "lots_of_junk"},
		{"", "unnamed"},
		{"!!!", "unnamed"},
		// HCL identifiers cannot start with a digit.
		{"2024 app", "r2024_app"},
	} {
		if got := sanitiseLabel(test.name); got != test.want {
			t.Errorf("sanitiseLabel(%q) = %q, want %q", test.name, got, test.want)
		}
	}
}

func TestLabellerUniquePerResourceType(t *testing.T) {
	labeller := newLabeller()

	// Different types can share a label, so an app and its rules read alike.
	if got := labeller.assign("ably_app", "Chat Service"); got != "chat_service" {
		t.Errorf("first app label = %q, want chat_service", got)
	}
	if got := labeller.assign("ably_rule_http", "Chat Service"); got != "chat_service" {
		t.Errorf("rule label = %q, want chat_service", got)
	}

	// The same type does not.
	if got := labeller.assign("ably_rule_http", "Chat Service"); got != "chat_service_2" {
		t.Errorf("second rule label = %q, want chat_service_2", got)
	}
	if got := labeller.assign("ably_rule_http", "Chat Service"); got != "chat_service_3" {
		t.Errorf("third rule label = %q, want chat_service_3", got)
	}
}

// TestRendererSkipsUnknownValues covers a provider leaving a value unknown after
// a read: nothing to write, and it has to be reported rather than guessed.
func TestRendererSkipsUnknownValues(t *testing.T) {
	schema := &tfprotov6.Schema{Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{Name: "name", Type: tftypes.String, Required: true},
			{Name: "region", Type: tftypes.String, Optional: true},
		},
	}}
	stateType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name":   tftypes.String,
		"region": tftypes.String,
	}}
	state := tftypes.NewValue(stateType, map[string]tftypes.Value{
		"name":   tftypes.NewValue(tftypes.String, "orders"),
		"region": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	})

	renderer := newRenderer(SecretsInline, nil)
	block, err := renderer.resourceBlock("ably_queue", "orders", schema, state)
	if err != nil {
		t.Fatalf("resourceBlock: %s", err)
	}

	if !strings.Contains(block, `name = "orders"`) {
		t.Errorf("known value was not written:\n%s", block)
	}
	if strings.Contains(block, "region") {
		t.Errorf("unknown value was written:\n%s", block)
	}
	if len(renderer.skipped) != 1 || !strings.Contains(renderer.skipped[0], "region") {
		t.Errorf("the skipped attribute was not reported, got %v", renderer.skipped)
	}
}

// TestRendererWritesNestedBlocks covers block-based schemas. The provider models
// everything as attributes today, so nothing in an export hits this path yet.
func TestRendererWritesNestedBlocks(t *testing.T) {
	nestedType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"value": tftypes.String}}
	schema := &tfprotov6.Schema{Block: &tfprotov6.SchemaBlock{
		BlockTypes: []*tfprotov6.SchemaNestedBlock{{
			TypeName: "setting",
			Nesting:  tfprotov6.SchemaNestedBlockNestingModeList,
			Block: &tfprotov6.SchemaBlock{
				Attributes: []*tfprotov6.SchemaAttribute{{Name: "value", Type: tftypes.String, Required: true}},
			},
		}},
	}}
	stateType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"setting": tftypes.List{ElementType: nestedType},
	}}
	state := tftypes.NewValue(stateType, map[string]tftypes.Value{
		"setting": tftypes.NewValue(tftypes.List{ElementType: nestedType}, []tftypes.Value{
			tftypes.NewValue(nestedType, map[string]tftypes.Value{"value": tftypes.NewValue(tftypes.String, "one")}),
			tftypes.NewValue(nestedType, map[string]tftypes.Value{"value": tftypes.NewValue(tftypes.String, "two")}),
		}),
	})

	renderer := newRenderer(SecretsInline, nil)
	block, err := renderer.resourceBlock("ably_example", "example", schema, state)
	if err != nil {
		t.Fatalf("resourceBlock: %s", err)
	}

	if strings.Count(block, "setting {") != 2 {
		t.Errorf("expected two setting blocks:\n%s", block)
	}
	parser := hclparse.NewParser()
	if _, diagnostics := parser.ParseHCL([]byte(block), "test.tf"); diagnostics.HasErrors() {
		t.Errorf("block rendering is not valid HCL: %s\n%s", diagnostics.Error(), block)
	}
}

// TestRendererVariableNamesAreUniquePerResource covers an easy collision: two
// resources of different types can share a label, and a variable name built from
// the label alone would serve both secrets.
func TestRendererVariableNamesAreUniquePerResource(t *testing.T) {
	schema := &tfprotov6.Schema{Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{Name: "url", Type: tftypes.String, Required: true, Sensitive: true},
		},
	}}
	stateType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"url": tftypes.String}}
	state := tftypes.NewValue(stateType, map[string]tftypes.Value{
		"url": tftypes.NewValue(tftypes.String, "https://secret.example"),
	})

	renderer := newRenderer(SecretsVars, nil)
	for _, resourceType := range []string{"ably_rule_http", "ably_rule_amqp_external"} {
		if _, err := renderer.resourceBlock(resourceType, "chat_service", schema, state); err != nil {
			t.Fatalf("resourceBlock(%s): %s", resourceType, err)
		}
	}

	if len(renderer.variables) != 2 {
		t.Fatalf("expected two variables, got %v", renderer.variables)
	}
	if renderer.variables[0].Name == renderer.variables[1].Name {
		t.Errorf("two resources sharing a label got the same variable %q", renderer.variables[0].Name)
	}
}

// TestRendererOmitsSensitiveCollectionsInVarsMode covers a sensitive attribute
// that isn't a single value: var.x for a list would not type-check.
func TestRendererOmitsSensitiveCollectionsInVarsMode(t *testing.T) {
	listType := tftypes.List{ElementType: tftypes.String}
	schema := &tfprotov6.Schema{Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{Name: "certs", Type: listType, Optional: true, Sensitive: true},
		},
	}}
	stateType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"certs": listType}}
	state := tftypes.NewValue(stateType, map[string]tftypes.Value{
		"certs": tftypes.NewValue(listType, []tftypes.Value{tftypes.NewValue(tftypes.String, "secret-cert")}),
	})

	renderer := newRenderer(SecretsVars, nil)
	block, err := renderer.resourceBlock("ably_rule_pulsar", "chat_service", schema, state)
	if err != nil {
		t.Fatalf("resourceBlock: %s", err)
	}

	if strings.Contains(block, "secret-cert") {
		t.Errorf("vars mode leaked a sensitive collection:\n%s", block)
	}
	if strings.Contains(block, "var.") {
		t.Errorf("vars mode pointed a collection at a string variable:\n%s", block)
	}
	if !strings.Contains(block, "# certs omitted") {
		t.Errorf("no comment explains the omission:\n%s", block)
	}
	if len(renderer.variables) != 0 {
		t.Errorf("expected no variable to be declared, got %v", renderer.variables)
	}
}

func TestSanitiseLabelNonASCIIDigit(t *testing.T) {
	// Unicode Nd is not ID_Start in HCL, so a leading Arabic-Indic digit needs
	// the same prefix an ASCII one gets.
	if got := sanitiseLabel("٣ chat"); !strings.HasPrefix(got, "r") {
		t.Errorf("sanitiseLabel(\"٣ chat\") = %q, want a leading r", got)
	}
}

// TestRendererVariableNamesIncludeCollectionKeys checks two sensitive values in
// one collection get separate variables.
func TestRendererVariableNamesIncludeCollectionKeys(t *testing.T) {
	elementType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"value": tftypes.String}}
	listType := tftypes.List{ElementType: elementType}
	schema := &tfprotov6.Schema{Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{{
			Name:     "headers",
			Optional: true,
			NestedType: &tfprotov6.SchemaObject{
				Nesting:    tfprotov6.SchemaObjectNestingModeList,
				Attributes: []*tfprotov6.SchemaAttribute{{Name: "value", Type: tftypes.String, Required: true, Sensitive: true}},
			},
		}},
	}}
	stateType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"headers": listType}}
	element := func(value string) tftypes.Value {
		return tftypes.NewValue(elementType, map[string]tftypes.Value{"value": tftypes.NewValue(tftypes.String, value)})
	}
	state := tftypes.NewValue(stateType, map[string]tftypes.Value{
		"headers": tftypes.NewValue(listType, []tftypes.Value{element("first-secret"), element("second-secret")}),
	})

	renderer := newRenderer(SecretsVars, nil)
	block, err := renderer.resourceBlock("ably_rule_http", "chat_service", schema, state)
	if err != nil {
		t.Fatalf("resourceBlock: %s", err)
	}

	if len(renderer.variables) != 2 {
		t.Fatalf("expected two variables, got %v", renderer.variables)
	}
	if renderer.variables[0].Name == renderer.variables[1].Name {
		t.Errorf("both list elements share the variable %q, so one secret is lost:\n%s",
			renderer.variables[0].Name, block)
	}
}
