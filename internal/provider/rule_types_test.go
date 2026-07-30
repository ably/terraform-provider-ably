package provider

import (
	"context"
	"testing"
)

// TestRuleTypeResources keeps RuleTypeResources in step with the registered
// resources, so a new rule resource can't be invisible to anything walking an
// account by ruleType.
func TestRuleTypeResources(t *testing.T) {
	ctx := context.Background()

	registered := map[string]bool{}
	for _, name := range ResourceTypeNames(ctx) {
		registered[name] = true
	}

	mapped := map[string]string{}
	for ruleType, resourceType := range RuleTypeResources {
		if !registered[resourceType] {
			t.Errorf("RuleTypeResources maps ruleType %q to %q, which the provider does not register", ruleType, resourceType)
		}
		if other, dup := mapped[resourceType]; dup {
			t.Errorf("resource %q is mapped from two ruleTypes, %q and %q", resourceType, other, ruleType)
		}
		mapped[resourceType] = ruleType
	}

	for name := range registered {
		if !IsRuleResourceType(name) {
			continue
		}
		if _, ok := mapped[name]; !ok {
			t.Errorf("rule resource %q has no ruleType in RuleTypeResources.\n"+
				"Add the Control API ruleType it manages to the table in rule_types.go, "+
				"otherwise the exporter cannot tell which rules belong to it.", name)
		}
	}
}
