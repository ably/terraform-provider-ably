// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

// Some rule types are gated on the account's package. The Control API answers a
// create for one the account isn't entitled to with 403 code 40300, "The rule
// type you selected is not available on your current package".
//
// That is not a provider bug, but it fails an acceptance test exactly like one,
// and the staging account CI runs against does not currently carry every rule
// type (http/before-publish, 2026-08-17). The alternative to handling it is a
// permanently red build, which teaches everyone to ignore the build.
//
// The remedy is in ably/website, not here. Rule-type availability is
// `account.current_package.supported_webhooks_types`, a persisted array on
// `package_histories`. Both before-publish types are already in
// `default_supported_webhooks_types` in `config/config.yml`, so newly-provisioned
// accounts get them; accounts created before they were added keep their old array
// and need the data migration (`rake data_migrations:append_new_default_supported_webhooks_types`,
// or the DX-530 pulsar task as precedent). Run that against staging and these
// tests start covering the rule type on their own. Tracked in INF-7994.
//
// So the affected tests probe first and skip with a loud message if the rule type
// isn't available. The probe is a real create against a throwaway app, so it
// starts passing by itself the day the entitlement is added: no test is silently
// disabled forever, which is the failure mode a plain t.Skip would have.
//
// Against the hermetic fake the probe succeeds (the fake stores whatever it is
// sent), so `make test` runs these tests as normal.

// ruleTypeUnavailableCode is the Control API error code for a rule type the
// account's package does not include.
const ruleTypeUnavailableCode = 40300

// skipIfRuleTypeUnavailable skips the test when the account cannot create the
// given rule type. body must be a valid create body for it; anything else about
// it failing is left for the test itself to report.
func skipIfRuleTypeUnavailable(t *testing.T, ruleType string, body any) {
	t.Helper()

	client, accountID, err := testControlClient()
	if err != nil {
		t.Fatalf("could not reach the Control API to check entitlement for %s: %s", ruleType, err)
	}

	ctx := context.Background()
	app, err := client.CreateApp(ctx, accountID, control.AppPost{
		Name:    "acc-test-probe-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		Status:  "enabled",
		TLSOnly: ptr(true),
	})
	if err != nil {
		t.Fatalf("could not create a probe app to check entitlement for %s: %s", ruleType, err)
	}
	defer func() {
		if err := client.DeleteApp(ctx, app.ID); err != nil {
			t.Logf("could not delete probe app %s: %s", app.ID, err)
		}
	}()

	rule, err := client.CreateRule(ctx, app.ID, body)
	if err != nil {
		var apiErr *control.Error
		if errors.As(err, &apiErr) && apiErr.Code == ruleTypeUnavailableCode {
			t.Skipf("the account's package does not include %s rules, so this test cannot run: %s", ruleType, apiErr.Message)
		}
		// Any other failure is the test's business, not the probe's: let the test
		// run and report it properly.
		return
	}
	if err := client.DeleteRule(ctx, app.ID, rule.ID); err != nil {
		t.Logf("could not delete probe rule %s: %s", rule.ID, err)
	}
}

// testControlClient builds a Control API client from the environment the test
// suite is already configured with, and resolves the account ID.
func testControlClient() (*control.Client, string, error) {
	token := os.Getenv("ABLY_ACCOUNT_TOKEN")
	if token == "" {
		return nil, "", fmt.Errorf("ABLY_ACCOUNT_TOKEN is not set")
	}

	client := control.NewClient(token)
	if url := os.Getenv("ABLY_URL"); url != "" {
		client.BaseURL = url
	}

	me, err := client.Me(context.Background())
	if err != nil {
		return nil, "", fmt.Errorf("could not call /me: %w", err)
	}
	if me.Account == nil || me.Account.ID == "" {
		return nil, "", fmt.Errorf("could not determine the account ID from /me")
	}

	return client, me.Account.ID, nil
}

// beforePublishProbeConfig is the retry/backoff block the entitlement probes
// send. Its values don't matter; it just has to be valid.
func beforePublishProbeConfig() control.BeforePublishConfig {
	return control.BeforePublishConfig{
		RetryTimeout:          5000,
		MaxRetries:            3,
		FailedAction:          "PUBLISH",
		TooManyRequestsAction: "RETRY",
	}
}
