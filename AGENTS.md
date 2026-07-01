# Working in this repo (for AI agents)

This is the Ably Terraform provider. It manages Ably resources (apps, keys,
namespaces, queues, integration rules) via the Ably Control API. The provider's
schema and model code is being moved onto code generation from the Control API
spec; `CODEGEN_STRATEGY.md` explains the why, the decisions, and the plan.

## The test loop (run this on every change)

```sh
make test
```

runs the unit tests plus the full acceptance suite against an in-process fake
Control API (`internal/provider/fake_control_api_test.go`). It needs **no
credentials and no network**, and it is the loop to run after any change. You do
not need real Ably credentials to develop or test locally.

`make testacc` runs the acceptance suite against a real Control API (needs
`ABLY_ACCOUNT_TOKEN`, optionally `ABLY_URL`). CI runs this against staging; you
generally don't run it locally.

CI also enforces `gofmt` and `go vet`, so keep `gofmt -l .` clean and
`go vet ./...` passing.

## Code generation

Schema and model code is generated. Regenerate with:

```sh
make generate
```

Generated code lives under `internal/provider/codegen/` and **is committed; do
not hand-edit it**. Change the inputs and regenerate:

- Simple resources (app, namespace, queue) are generated from the vendored
  OpenAPI spec `codegen/control-api.yaml` (sourced from the `ably/docs` repo).
- Integration-rule families are generated from the in-repo `control` rule types
  via `codegen/ruletypesgen`, with descriptions and metadata sourced from the
  spec and an overrides table.

Generation produces schema + model only. **CRUD wiring to the control client is
always hand-written.**

Step-by-step runbooks are in `DEVELOPMENT.md`:

- "Adding a new integration rule"
- "Porting a resource onto generated code" (reference example:
  `internal/provider/resource_ably_rule_bodyguard.go`)

and the pipeline details are in `codegen/README.md`.

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
