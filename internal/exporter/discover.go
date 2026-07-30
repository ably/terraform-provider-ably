package exporter

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/ably/terraform-provider-ably/internal/provider"
)

// Resource types with their own Control API collection endpoint: one
// account-level list of apps and three app-level lists. Everything else in an
// account is a rule, discovered via provider.RuleTypeResources.
const (
	resourceTypeApp       = "ably_app"
	resourceTypeKey       = "ably_api_key"
	resourceTypeNamespace = "ably_namespace"
	resourceTypeQueue     = "ably_queue"
)

// SupportedResourceTypes returns every resource type the exporter can find.
//
// Rule types come from the provider's registry, so a new rule resource is
// exported as soon as it appears there. Non-rule resources need a lister in
// discover; TestSupportedResourceTypesCoverage checks for that.
func SupportedResourceTypes() []string {
	types := []string{
		resourceTypeApp,
		resourceTypeKey,
		resourceTypeNamespace,
		resourceTypeQueue,
	}
	for _, resourceType := range provider.RuleTypeResources {
		types = append(types, resourceType)
	}
	sort.Strings(types)
	return types
}

// Target is one resource found in an account, in the order it will be written.
type Target struct {
	// ResourceType is the Terraform resource type, e.g. "ably_rule_http".
	ResourceType string
	// ID is the resource's own Control API ID.
	ID string
	// AppID is the owning app, empty for account-level resources.
	AppID string
	// AppName is the owning app's name, used to build HCL labels.
	AppName string
	// Name is the resource's own name where it has one. Rules do not.
	Name string
	// RuleType is the Control API ruleType for rule resources, empty otherwise.
	RuleType string
}

// discovery is the result of walking an account.
type discovery struct {
	Targets  []Target
	Warnings []string
}

// discover walks an account and returns every resource the exporter can
// export. appFilter, when non-empty, keeps only apps whose ID or name matches.
func discover(ctx context.Context, client *control.Client, accountID string, appFilter []string) (*discovery, error) {
	result := &discovery{}

	apps, err := client.ListApps(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("listing apps: %w", err)
	}

	apps, unmatched := filterApps(apps, appFilter)
	if len(apps) == 0 {
		if len(appFilter) > 0 {
			return nil, fmt.Errorf("no apps in the account match %s", strings.Join(appFilter, ", "))
		}
		result.Warnings = append(result.Warnings, "the account has no apps")
		return result, nil
	}
	// A filter that matched nothing is almost certainly a typo, and silently
	// exporting the rest would look like it worked.
	if len(unmatched) > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"no app matched %s, so nothing was exported for it", strings.Join(unmatched, ", ")))
	}

	// Sort apps by name so output ordering doesn't depend on API ordering.
	sort.Slice(apps, func(i, j int) bool {
		if apps[i].Name != apps[j].Name {
			return apps[i].Name < apps[j].Name
		}
		return apps[i].ID < apps[j].ID
	})

	for _, app := range apps {
		result.Targets = append(result.Targets, Target{
			ResourceType: resourceTypeApp,
			ID:           app.ID,
			AppName:      app.Name,
			Name:         app.Name,
		})

		appTargets, warnings, err := discoverApp(ctx, client, app)
		if err != nil {
			return nil, err
		}
		result.Targets = append(result.Targets, appTargets...)
		result.Warnings = append(result.Warnings, warnings...)
	}

	return result, nil
}

// discoverApp lists everything that lives inside a single app.
func discoverApp(ctx context.Context, client *control.Client, app control.AppResponse) ([]Target, []string, error) {
	var targets []Target
	var warnings []string

	target := func(resourceType, id, name, ruleType string) Target {
		return Target{
			ResourceType: resourceType,
			ID:           id,
			AppID:        app.ID,
			AppName:      app.Name,
			Name:         name,
			RuleType:     ruleType,
		}
	}

	keys, err := client.ListKeys(ctx, app.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("listing keys for app %s: %w", app.ID, err)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].ID < keys[j].ID })
	for _, key := range keys {
		// A non-zero status means revoked. The provider's Read treats those as
		// gone, so exporting one gives config for a resource that isn't there.
		if key.Status != 0 {
			continue
		}
		targets = append(targets, target(resourceTypeKey, key.ID, key.Name, ""))
	}

	namespaces, err := client.ListNamespaces(ctx, app.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("listing namespaces for app %s: %w", app.ID, err)
	}
	sort.Slice(namespaces, func(i, j int) bool { return namespaces[i].ID < namespaces[j].ID })
	for _, namespace := range namespaces {
		targets = append(targets, target(resourceTypeNamespace, namespace.ID, namespace.ID, ""))
	}

	queues, err := client.ListQueues(ctx, app.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("listing queues for app %s: %w", app.ID, err)
	}
	sort.Slice(queues, func(i, j int) bool { return queues[i].ID < queues[j].ID })
	for _, queue := range queues {
		targets = append(targets, target(resourceTypeQueue, queue.ID, queue.Name, ""))
	}

	rules, err := client.ListRules(ctx, app.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("listing rules for app %s: %w", app.ID, err)
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
	for _, rule := range rules {
		resourceType, ok := provider.RuleTypeResources[rule.RuleType]
		if !ok {
			warnings = append(warnings, fmt.Sprintf(
				"skipped rule %s in app %s: the provider has no resource for ruleType %q",
				rule.ID, app.ID, rule.RuleType))
			continue
		}
		targets = append(targets, target(resourceType, rule.ID, "", rule.RuleType))
	}

	return targets, warnings, nil
}

// filterApps keeps the apps matching one of the filters by ID or name, or every
// app when no filter is given. It also returns the filters that matched nothing.
func filterApps(apps []control.AppResponse, appFilter []string) (kept []control.AppResponse, unmatched []string) {
	if len(appFilter) == 0 {
		return apps, nil
	}

	matched := make(map[string]bool, len(appFilter))
	for _, app := range apps {
		for _, filter := range appFilter {
			wanted := strings.ToLower(strings.TrimSpace(filter))
			if wanted == strings.ToLower(app.ID) || wanted == strings.ToLower(app.Name) {
				matched[filter] = true
				kept = append(kept, app)
				break
			}
		}
	}

	for _, filter := range appFilter {
		if !matched[filter] {
			unmatched = append(unmatched, filter)
		}
	}
	return kept, unmatched
}
