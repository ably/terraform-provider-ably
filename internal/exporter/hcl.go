package exporter

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// SecretMode controls what the exporter writes for sensitive attributes.
type SecretMode string

const (
	// SecretsInline writes sensitive values into the configuration as
	// literals. The generated config is accurate and plans clean, but must be
	// handled like any other file holding credentials.
	SecretsInline SecretMode = "inline"
	// SecretsVars replaces sensitive values with references to sensitive
	// Terraform variables, declared in variables.tf. Safe to commit, but the
	// values have to be supplied before the config will apply.
	SecretsVars SecretMode = "vars"
	// SecretsOmit leaves sensitive attributes out entirely, with a comment
	// where each one would have gone.
	SecretsOmit SecretMode = "omit"
)

// ParseSecretMode validates a -secrets flag value.
func ParseSecretMode(value string) (SecretMode, error) {
	switch SecretMode(value) {
	case SecretsInline, SecretsVars, SecretsOmit:
		return SecretMode(value), nil
	default:
		return "", fmt.Errorf("unknown secrets mode %q, expected inline, vars or omit", value)
	}
}

// variableDecl is a sensitive Terraform variable the exporter declares in place
// of a sensitive value, in SecretsVars mode.
type variableDecl struct {
	Name        string
	Description string
	Type        string
}

// renderer turns Terraform values into HCL, using the resource schema to decide
// what belongs in a configuration: which attributes a user can set, which are
// sensitive, and how values nest.
type renderer struct {
	secrets SecretMode
	// references maps a resource type and Control API ID onto an HCL expression
	// referring to the resource that owns it, so app_id becomes ably_app.foo.id.
	references map[resourceRef]string

	// Collected while rendering.
	variables []variableDecl
	sensitive []string
	skipped   []string
	// missing records required attributes the Control API withheld, which leave
	// the config incomplete until someone fills them in.
	missing []string
}

func newRenderer(secrets SecretMode, references map[resourceRef]string) *renderer {
	return &renderer{secrets: secrets, references: references}
}

// resourceIdentity is how the renderer refers to what it is writing: an address
// for messages, and a prefix unique to the resource for naming variables.
type resourceIdentity struct {
	address      string
	variableBase string
}

func identityOf(resourceType, label string) resourceIdentity {
	return resourceIdentity{
		address:      resourceType + "." + label,
		variableBase: strings.TrimPrefix(resourceType, "ably_") + "_" + label,
	}
}

// resourceBlock renders a single resource block.
func (r *renderer) resourceBlock(resourceType, label string, schema *tfprotov6.Schema, state tftypes.Value) (string, error) {
	if schema == nil || schema.Block == nil {
		return "", fmt.Errorf("resource type %s has no schema", resourceType)
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "resource %q %q {\n", resourceType, label)
	if err := r.writeBlockBody(&buf, 1, schema.Block, state, identityOf(resourceType, label), nil); err != nil {
		return "", err
	}
	buf.WriteString("}\n")
	return buf.String(), nil
}

// writeBlockBody writes the attributes and nested blocks of one block body. path
// is the attribute path walked so far, used for variable names and reported
// notes.
func (r *renderer) writeBlockBody(buf *strings.Builder, depth int, block *tfprotov6.SchemaBlock, value tftypes.Value, identity resourceIdentity, path []string) error {
	if block == nil {
		return nil
	}
	if err := r.writeAttributes(buf, depth, block.Attributes, value, identity, path); err != nil {
		return err
	}
	return r.writeNestedBlocks(buf, depth, block.BlockTypes, value, identity, path)
}

// writeAttributes writes every attribute a user is allowed to set.
func (r *renderer) writeAttributes(buf *strings.Builder, depth int, attributes []*tfprotov6.SchemaAttribute, value tftypes.Value, identity resourceIdentity, path []string) error {
	if value.IsNull() || !value.IsKnown() {
		return nil
	}

	values := map[string]tftypes.Value{}
	if err := value.As(&values); err != nil {
		return fmt.Errorf("decoding object at %s: %w", pathString(path), err)
	}

	for _, attribute := range sortAttributes(attributes) {
		attributeValue, ok := values[attribute.Name]
		if !ok {
			continue
		}
		attributePath := childPath(path, attribute.Name)

		// Computed-only attributes are the provider's to set, not the user's:
		// writing them back would be rejected by Terraform.
		if isComputedOnly(attribute) {
			continue
		}
		// Write-only attributes are never persisted, so there is nothing to
		// read back and nothing to export.
		if attribute.WriteOnly {
			continue
		}
		if !attributeValue.IsKnown() {
			r.skipped = append(r.skipped, fmt.Sprintf("%s.%s (value unknown after read)", identity.address, pathString(attributePath)))
			continue
		}
		// A required value the API withheld has to be flagged: left alone it
		// either produces HCL Terraform refuses (null) or overwrites a live
		// credential with an empty string on apply. Kafka SASL arrives this way.
		if attribute.Required && withheld(attribute, attributeValue) {
			r.writeMissingRequired(buf, depth, attribute, identity, attributePath)
			continue
		}
		if attributeValue.IsNull() {
			continue
		}

		if attribute.Sensitive {
			written, err := r.writeSensitive(buf, depth, attribute, identity, attributePath)
			if err != nil {
				return err
			}
			if written {
				continue
			}
		}

		writeIndent(buf, depth)
		fmt.Fprintf(buf, "%s = ", attribute.Name)
		if err := r.writeValue(buf, depth, attribute, attributeValue, identity, attributePath); err != nil {
			return err
		}
		buf.WriteString("\n")
	}

	return nil
}

// writeMissingRequired handles a required attribute the Control API withheld.
//
// Kafka SASL passwords, AWS secret keys and Pulsar TLS certs are accepted on
// write and never read back. Nothing can recover them, so the gap is made
// obvious rather than left to fail at plan time. In vars mode it wires up a
// variable, so the config is complete once the value is supplied.
func (r *renderer) writeMissingRequired(buf *strings.Builder, depth int, attribute *tfprotov6.SchemaAttribute, identity resourceIdentity, path []string) {
	name := pathString(path)

	if r.secrets == SecretsVars && isSimpleType(attribute) {
		variable := r.declareVariable(attribute, identity, path,
			fmt.Sprintf("%s of %s. Required, but the Control API does not return it.", name, identity.address))
		writeIndent(buf, depth)
		fmt.Fprintf(buf, "%s = var.%s\n", attribute.Name, variable)
		r.missing = append(r.missing, fmt.Sprintf("%s.%s (required; supply it in variables.tf)", identity.address, name))
		return
	}

	writeIndent(buf, depth)
	fmt.Fprintf(buf, "# TODO: %s is required but the Control API does not return it. Fill it in before applying.\n", attribute.Name)
	r.missing = append(r.missing, fmt.Sprintf("%s.%s (required; the Control API does not return it)", identity.address, name))
}

// withheld reports whether a value looks like one the Control API withheld.
//
// Null is the obvious case. An empty string counts only for sensitive
// attributes: the provider maps absent credentials to "" rather than null (see
// GetRuleResponse), and an empty credential is never intended. Elsewhere an empty
// string can be legitimate.
func withheld(attribute *tfprotov6.SchemaAttribute, value tftypes.Value) bool {
	if value.IsNull() {
		return true
	}
	if !attribute.Sensitive || !value.Type().Is(tftypes.String) {
		return false
	}
	var literal string
	if err := value.As(&literal); err != nil {
		return false
	}
	return literal == ""
}

// declareVariable records a variable and returns its name. The name carries the
// resource type and label as well as the path, because two resources of different
// types can share a label and would otherwise share one variable.
func (r *renderer) declareVariable(attribute *tfprotov6.SchemaAttribute, identity resourceIdentity, path []string, description string) string {
	variable := sanitiseLabel(identity.variableBase + "_" + strings.Join(path, "_"))
	r.variables = append(r.variables, variableDecl{
		Name:        variable,
		Description: description,
		Type:        terraformType(attribute.Type),
	})
	return variable
}

// writeSensitive handles a sensitive attribute according to the secrets mode.
// It returns true when it has dealt with the attribute itself, and false when
// the value should be written inline as normal.
func (r *renderer) writeSensitive(buf *strings.Builder, depth int, attribute *tfprotov6.SchemaAttribute, identity resourceIdentity, path []string) (bool, error) {
	name := pathString(path)
	r.sensitive = append(r.sensitive, fmt.Sprintf("%s.%s", identity.address, name))

	switch r.secrets {
	case SecretsOmit:
		writeIndent(buf, depth)
		fmt.Fprintf(buf, "# %s omitted: sensitive value. Re-run with -secrets=inline or -secrets=vars to export it.\n", attribute.Name)
		return true, nil

	case SecretsVars:
		// A variable stands in for a single string, number or bool. An object or
		// collection has no equivalent, so leave it out rather than emit config
		// that cannot type-check.
		if !isSimpleType(attribute) {
			writeIndent(buf, depth)
			fmt.Fprintf(buf, "# %s omitted: sensitive %s, which the exporter cannot replace with a variable. Re-run with -secrets=inline to export it.\n",
				attribute.Name, attribute.Type)
			return true, nil
		}

		variable := r.declareVariable(attribute, identity, path,
			fmt.Sprintf("%s of %s. Sensitive, so the exporter did not write its value.", name, identity.address))
		writeIndent(buf, depth)
		fmt.Fprintf(buf, "%s = var.%s\n", attribute.Name, variable)
		return true, nil

	default:
		return false, nil
	}
}

// isSimpleType reports whether an attribute holds a single primitive value, the
// only shape a variable can stand in for.
func isSimpleType(attribute *tfprotov6.SchemaAttribute) bool {
	if attribute.NestedType != nil || attribute.Type == nil {
		return false
	}
	return attribute.Type.Is(tftypes.String) || attribute.Type.Is(tftypes.Number) || attribute.Type.Is(tftypes.Bool)
}

// terraformType names the Terraform type to declare a variable with.
func terraformType(valueType tftypes.Type) string {
	switch {
	case valueType.Is(tftypes.Number):
		return "number"
	case valueType.Is(tftypes.Bool):
		return "bool"
	default:
		return "string"
	}
}

// writeValue writes a single value, recursing into nested objects.
func (r *renderer) writeValue(buf *strings.Builder, depth int, attribute *tfprotov6.SchemaAttribute, value tftypes.Value, identity resourceIdentity, path []string) error {
	if attribute.NestedType != nil {
		return r.writeNestedValue(buf, depth, attribute.NestedType, value, identity, path)
	}

	// An ID that belongs to another exported resource is written as a
	// reference so the config carries the dependency rather than a literal.
	if len(path) > 0 && value.Type().Is(tftypes.String) {
		if resourceType, referenceable := referenceableAttributes[path[len(path)-1]]; referenceable {
			var literal string
			if err := value.As(&literal); err == nil {
				if expression, ok := r.references[resourceRef{resourceType, literal}]; ok {
					buf.WriteString(expression)
					return nil
				}
			}
		}
	}

	return r.writePrimitive(buf, depth, value, path)
}

// writeNestedValue writes a nested attribute: a single object, or a
// list/set/map of objects.
func (r *renderer) writeNestedValue(buf *strings.Builder, depth int, nested *tfprotov6.SchemaObject, value tftypes.Value, identity resourceIdentity, path []string) error {
	// elementKey distinguishes collection members, so a variable name or reported
	// path points at one element.
	writeObject := func(depth int, object tftypes.Value, elementKey string) error {
		objectPath := path
		if elementKey != "" {
			objectPath = childPath(path, elementKey)
		}
		buf.WriteString("{\n")
		if err := r.writeAttributes(buf, depth+1, nested.Attributes, object, identity, objectPath); err != nil {
			return err
		}
		writeIndent(buf, depth)
		buf.WriteString("}")
		return nil
	}

	switch nested.Nesting {
	case tfprotov6.SchemaObjectNestingModeSingle:
		return writeObject(depth, value, "")

	case tfprotov6.SchemaObjectNestingModeList, tfprotov6.SchemaObjectNestingModeSet:
		elements := []tftypes.Value{}
		if err := value.As(&elements); err != nil {
			return fmt.Errorf("decoding %s: %w", pathString(path), err)
		}
		buf.WriteString("[\n")
		for index, element := range elements {
			writeIndent(buf, depth+1)
			if err := writeObject(depth+1, element, strconv.Itoa(index)); err != nil {
				return err
			}
			buf.WriteString(",\n")
		}
		writeIndent(buf, depth)
		buf.WriteString("]")
		return nil

	case tfprotov6.SchemaObjectNestingModeMap:
		elements := map[string]tftypes.Value{}
		if err := value.As(&elements); err != nil {
			return fmt.Errorf("decoding %s: %w", pathString(path), err)
		}
		buf.WriteString("{\n")
		for _, key := range sortedKeys(elements) {
			writeIndent(buf, depth+1)
			fmt.Fprintf(buf, "%s = ", quoteString(key))
			if err := writeObject(depth+1, elements[key], key); err != nil {
				return err
			}
			buf.WriteString("\n")
		}
		writeIndent(buf, depth)
		buf.WriteString("}")
		return nil

	default:
		return fmt.Errorf("unsupported nesting mode %d at %s", nested.Nesting, pathString(path))
	}
}

// writePrimitive writes a value whose shape comes from its type: strings,
// numbers, bools and collections of them.
func (r *renderer) writePrimitive(buf *strings.Builder, depth int, value tftypes.Value, path []string) error {
	valueType := value.Type()

	switch {
	case valueType.Is(tftypes.String):
		var literal string
		if err := value.As(&literal); err != nil {
			return fmt.Errorf("decoding string at %s: %w", pathString(path), err)
		}
		buf.WriteString(quoteString(literal))
		return nil

	case valueType.Is(tftypes.Bool):
		var literal bool
		if err := value.As(&literal); err != nil {
			return fmt.Errorf("decoding bool at %s: %w", pathString(path), err)
		}
		fmt.Fprintf(buf, "%t", literal)
		return nil

	case valueType.Is(tftypes.Number):
		number := new(big.Float)
		if err := value.As(&number); err != nil {
			return fmt.Errorf("decoding number at %s: %w", pathString(path), err)
		}
		buf.WriteString(formatNumber(number))
		return nil

	case valueType.Is(tftypes.List{}), valueType.Is(tftypes.Set{}), valueType.Is(tftypes.Tuple{}):
		elements := []tftypes.Value{}
		if err := value.As(&elements); err != nil {
			return fmt.Errorf("decoding collection at %s: %w", pathString(path), err)
		}
		return r.writeCollection(buf, depth, elements, path)

	case valueType.Is(tftypes.Map{}), valueType.Is(tftypes.Object{}):
		elements := map[string]tftypes.Value{}
		if err := value.As(&elements); err != nil {
			return fmt.Errorf("decoding map at %s: %w", pathString(path), err)
		}
		buf.WriteString("{\n")
		for _, key := range sortedKeys(elements) {
			writeIndent(buf, depth+1)
			fmt.Fprintf(buf, "%s = ", quoteString(key))
			if err := r.writePrimitive(buf, depth+1, elements[key], childPath(path, key)); err != nil {
				return err
			}
			buf.WriteString("\n")
		}
		writeIndent(buf, depth)
		buf.WriteString("}")
		return nil

	default:
		return fmt.Errorf("cannot render type %s at %s", valueType, pathString(path))
	}
}

// writeCollection writes a list, set or tuple. Short primitive collections stay
// on one line.
func (r *renderer) writeCollection(buf *strings.Builder, depth int, elements []tftypes.Value, path []string) error {
	if len(elements) == 0 {
		buf.WriteString("[]")
		return nil
	}

	if collectionFitsOneLine(elements) {
		buf.WriteString("[")
		for i, element := range elements {
			if i > 0 {
				buf.WriteString(", ")
			}
			if err := r.writePrimitive(buf, depth, element, path); err != nil {
				return err
			}
		}
		buf.WriteString("]")
		return nil
	}

	buf.WriteString("[\n")
	for _, element := range elements {
		writeIndent(buf, depth+1)
		if err := r.writePrimitive(buf, depth+1, element, path); err != nil {
			return err
		}
		buf.WriteString(",\n")
	}
	writeIndent(buf, depth)
	buf.WriteString("]")
	return nil
}

// writeNestedBlocks renders nested blocks. The provider models everything as
// attributes today; this is here so a resource added with real blocks still
// exports.
func (r *renderer) writeNestedBlocks(buf *strings.Builder, depth int, blocks []*tfprotov6.SchemaNestedBlock, value tftypes.Value, identity resourceIdentity, path []string) error {
	if len(blocks) == 0 || value.IsNull() || !value.IsKnown() {
		return nil
	}

	values := map[string]tftypes.Value{}
	if err := value.As(&values); err != nil {
		return fmt.Errorf("decoding object at %s: %w", pathString(path), err)
	}

	sorted := make([]*tfprotov6.SchemaNestedBlock, 0, len(blocks))
	for _, block := range blocks {
		if block != nil {
			sorted = append(sorted, block)
		}
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].TypeName < sorted[j].TypeName })

	for _, block := range sorted {
		blockValue, ok := values[block.TypeName]
		if !ok || blockValue.IsNull() || !blockValue.IsKnown() {
			continue
		}
		blockPath := childPath(path, block.TypeName)

		writeBlock := func(body tftypes.Value, labels ...string) error {
			writeIndent(buf, depth)
			buf.WriteString(block.TypeName)
			for _, blockLabel := range labels {
				fmt.Fprintf(buf, " %s", quoteString(blockLabel))
			}
			buf.WriteString(" {\n")
			if err := r.writeBlockBody(buf, depth+1, block.Block, body, identity, blockPath); err != nil {
				return err
			}
			writeIndent(buf, depth)
			buf.WriteString("}\n")
			return nil
		}

		switch block.Nesting {
		case tfprotov6.SchemaNestedBlockNestingModeSingle, tfprotov6.SchemaNestedBlockNestingModeGroup:
			if err := writeBlock(blockValue); err != nil {
				return err
			}

		case tfprotov6.SchemaNestedBlockNestingModeList, tfprotov6.SchemaNestedBlockNestingModeSet:
			elements := []tftypes.Value{}
			if err := blockValue.As(&elements); err != nil {
				return fmt.Errorf("decoding %s: %w", pathString(blockPath), err)
			}
			for _, element := range elements {
				if err := writeBlock(element); err != nil {
					return err
				}
			}

		case tfprotov6.SchemaNestedBlockNestingModeMap:
			elements := map[string]tftypes.Value{}
			if err := blockValue.As(&elements); err != nil {
				return fmt.Errorf("decoding %s: %w", pathString(blockPath), err)
			}
			for _, key := range sortedKeys(elements) {
				if err := writeBlock(elements[key], key); err != nil {
					return err
				}
			}

		default:
			return fmt.Errorf("unsupported block nesting mode %d at %s", block.Nesting, pathString(blockPath))
		}
	}

	return nil
}

// sortAttributes puts app_id first, which ties the resource to its app, then the
// rest alphabetically.
func sortAttributes(attributes []*tfprotov6.SchemaAttribute) []*tfprotov6.SchemaAttribute {
	sorted := make([]*tfprotov6.SchemaAttribute, 0, len(attributes))
	for _, attribute := range attributes {
		if attribute != nil {
			sorted = append(sorted, attribute)
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		if (sorted[i].Name == "app_id") != (sorted[j].Name == "app_id") {
			return sorted[i].Name == "app_id"
		}
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}

// collectionFitsOneLine reports whether a collection is small enough to write
// inline.
func collectionFitsOneLine(elements []tftypes.Value) bool {
	if len(elements) > 4 {
		return false
	}
	width := 0
	for _, element := range elements {
		elementType := element.Type()
		if !elementType.Is(tftypes.String) && !elementType.Is(tftypes.Number) && !elementType.Is(tftypes.Bool) {
			return false
		}
		width += len(element.String())
	}
	return width <= 72
}

// formatNumber writes integers without a decimal point and everything else at
// its shortest exact form.
func formatNumber(number *big.Float) string {
	if number.IsInt() {
		integer, _ := number.Int(nil)
		return integer.String()
	}
	return number.Text('f', -1)
}

// quoteString renders a string as an HCL quoted template. ${ and %{ are escaped
// so values containing them aren't read as interpolations.
func quoteString(value string) string {
	var buf strings.Builder
	buf.WriteByte('"')
	for i := 0; i < len(value); i++ {
		switch character := value[i]; character {
		case '\\':
			buf.WriteString(`\\`)
		case '"':
			buf.WriteString(`\"`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		case '$', '%':
			buf.WriteByte(character)
			if i+1 < len(value) && value[i+1] == '{' {
				buf.WriteByte(character)
			}
		default:
			buf.WriteByte(character)
		}
	}
	buf.WriteByte('"')
	return buf.String()
}

func writeIndent(buf *strings.Builder, depth int) {
	buf.WriteString(strings.Repeat("  ", depth))
}

// childPath extends an attribute path without sharing its backing array.
func childPath(path []string, name string) []string {
	child := make([]string, len(path), len(path)+1)
	copy(child, path)
	return append(child, name)
}

func pathString(path []string) string {
	if len(path) == 0 {
		return "(root)"
	}
	return strings.Join(path, ".")
}

func sortedKeys(values map[string]tftypes.Value) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
