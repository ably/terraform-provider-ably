# Code generation

This directory holds the inputs for generating Terraform schema and model code
from the Ably Control API's OpenAPI spec, the first step of the strategy in
[`../CODEGEN_STRATEGY.md`](../CODEGEN_STRATEGY.md).

## What's here

- `swagger.yaml` — a vendored snapshot of the Control API OpenAPI spec. The
  source of truth lives in the `ably/website` repo
  (`apps/website/swagger/v1/swagger.yaml`); we vendor a copy so generation is
  self-contained and runnable in CI without checking out that repo. Refresh it
  by copying the latest spec over this file.
- `generator_config.yml` — maps each resource to its create/read/update/delete
  path and method, plus the per-resource aliases needed to get past spec quirks.
- `spec.json` — the intermediate Provider Code Specification, produced by
  `tfplugingen-openapi`. Regenerated, not hand-edited.

## How to regenerate

```
make generate
```

That runs HashiCorp's two tools in sequence (pinned versions, fetched via
`go run`):

1. `tfplugingen-openapi` turns `swagger.yaml` + `generator_config.yml` into
   `spec.json`.
2. `tfplugingen-framework` turns `spec.json` into Go schema + model code under
   `internal/provider/codegen/resource_<name>/`.

The output is committed so changes are reviewable and a future CI check can
assert that regeneration produces no diff.

## Scope and caveats

This is deliberately limited right now:

- **Simple resources only.** `app`, `namespace` and `queue` are generated. The
  integration rules use an OpenAPI `oneOf` + discriminator that
  `tfplugingen-openapi` cannot handle, so they are not generated from the spec;
  the rule families are driven from the in-repo `control` types instead (see the
  strategy doc).
- **Schema + model only.** The tools do not emit CRUD wiring. All wiring to the
  `control` client stays hand-written and is not generated here.
- **Both tools are tech preview.** `tfplugingen-openapi` last shipped v0.3.0
  (Jan 2024). It works on our spec today; we are not betting anything load
  bearing on a future release.
- **The generated code is not yet wired into the live resources.** It is
  committed as the reviewable output of the pipeline. Retrofitting the existing
  hand-written resources onto it is a separate, deliberate step, because the
  spec is description-poor (the generated schemas lose the hand-written
  descriptions and plan modifiers) and some resources diverge from the spec
  shape on purpose (e.g. `queue` flattens the API's nested `amqp`/`stomp`
  objects into flat attributes). See the Phase 1 findings in the strategy doc.

## Known per-resource quirks (encoded in `generator_config.yml`)

- A parent path parameter (`account_id` for apps, `app_id` for namespaces and
  queues) collides with the same-named field in the response body, which makes
  `tfplugingen-framework` error on a duplicate attribute. We alias the path
  parameter (e.g. `account_id` -> `parent_account_id`) to get past it; the
  redundant attribute is dropped during integration.
