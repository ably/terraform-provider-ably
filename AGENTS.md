# Working in this repo (for AI agents)

This is the Ably Terraform provider. It manages Ably resources (apps, keys,
namespaces, queues, integration rules) via the Ably Control API. Its schema and
model code is generated from the Control API's OpenAPI spec.

## The test loop (run this on every change)

```sh
make test
```

runs the unit tests plus the full acceptance suite against an in-process fake
Control API (`internal/provider/fake_control_api_test.go`). It needs **no
credentials and no network**, and it is the loop to run after any change. You do
not need real Ably credentials to develop or test locally.

`make testacc` runs the acceptance suite against a real Control API (needs
`ABLY_ACCOUNT_TOKEN` and an explicit `ABLY_URL`). CI runs this against staging;
you generally don't run it locally. Real runs are strictly opt-in: with
`TF_ACC` set but no `ABLY_URL`, the suite refuses to start rather than
defaulting to the production API, and falsy `TF_ACC` values (`0`, `false`)
fall back to the hermetic fake.

CI also enforces `gofmt` and `go vet`, so keep `gofmt -l .` clean and
`go vet ./...` passing.

## Code generation

Schema and model code for resources and data sources is generated with `make
generate`, and the output is committed under `internal/provider/codegen/`. **Do
not hand-edit it**: change the inputs in `codegen/` and regenerate. Generation
produces schema + model only, so **CRUD wiring to the control client is always
hand-written**. The pipeline is described in `codegen/README.md`, and
`DEVELOPMENT.md` has the runbooks for adding a rule or a data source and for
porting a resource onto generated code.

`internal/provider/spec_coverage_test.go` fails when the Control API spec carries
rule types or operations the provider hasn't accounted for, so new API surface has
to be a decision someone writes down rather than something we miss.

## The account exporter

`cmd/ably-exporter` (package `internal/exporter`) generates Terraform config for
an existing account. It drives the provider in-process over protocol v6, so
schema changes need no work here, but two things do:

- **A new integration rule** needs a line in `RuleTypeResources`
  (`internal/provider/rule_types.go`) mapping its `ruleType` to the resource type.
  `TestRuleTypeResources` catches a missing entry.
- **A new family of resources** needs a lister in `internal/exporter/discover.go`
  and its type in `SupportedResourceTypes`.
  `TestSupportedResourceTypesCoverage` catches a gap.

`EXPORTER.md` covers the design, including the validate-and-repair pass for
provider validators the protocol schema doesn't expose.

## Things that will bite you (learned the hard way)

- **Stale `dev_overrides`.** A `dev_overrides` block in `~/.terraformrc` takes
  precedence over the test framework's in-process provider, so a stale installed
  binary can silently run instead of your code (edits appear to do nothing). The
  hermetic harness builds a fresh provider into a temp dir to avoid this; don't
  undo that.
- **The fake echoes.** The hermetic fake returns what it is sent, so it cannot
  catch real-API contract mismatches (wrong field names, enum values, defaults).
  It catches schema, diff, import and CRUD-wiring bugs; only the staging
  acceptance run catches real-API bugs. Encode real-API behaviour in unit tests
  (see the bodyguard preserve-from-plan test).
- **`control/` is a separate Go module** (`control/go.mod`). `go test ./...`
  from the repo root does not descend into it; test it with
  `cd control && go test ./...`.
- **CI installs Terraform explicitly** (via `setup-terraform` in
  `.github/workflows/check.yml`) because the test framework's auto-install fails
  on an expired HashiCorp GPG key. Leave that step in place.

## Conventions

Commit, PR and release conventions are in `CONTRIBUTING.md`. After any schema
change, regenerate the registry docs (`go run
github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate`) and
commit them.
