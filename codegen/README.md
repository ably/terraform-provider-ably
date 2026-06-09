# Code generation

This directory holds the inputs for generating Terraform schema and model code
from the Ably Control API's OpenAPI spec, the first step of the strategy in
[`../CODEGEN_STRATEGY.md`](../CODEGEN_STRATEGY.md).

## What's here

- `control-api.yaml` — a vendored snapshot of the Control API OpenAPI spec. We
  source it from the `ably/docs` repo (`static/open-specs/control-v1.yaml`),
  which is the published, description-rich version (~1,160 field descriptions
  versus ~150 in the `ably/website` rswag output). Generating from it gives the
  generated schemas correct attribute documentation. We vendor a copy so
  generation is self-contained and runnable in CI without checking out that
  repo; refresh it by copying the latest spec over this file.

  One local patch is applied on top of the upstream copy: `conflationEnabled` in
  the namespace schemas is missing `type: boolean` upstream (it has a default,
  description and example but no type), which makes `tfplugingen-openapi` skip
  it. We add the type back. This should be fixed in `ably/docs` and the patch
  dropped on the next refresh.
- `generator_config.yml` — maps each simple resource to its
  create/read/update/delete path and method, plus the per-resource aliases
  needed to get past spec quirks.
- `spec.json` — the intermediate Provider Code Specification for the simple
  resources, produced by `tfplugingen-openapi`. Regenerated, not hand-edited.
- `ruletypesgen/` — a small Go program that reflects over the in-repo `control`
  rule types and emits a Provider Code Spec for the integration-rule families
  the OpenAPI generator can't handle (the `oneOf` + discriminator union). This
  is "Track B": the rules are generated from the curated control types, not the
  spec.
- `rules_spec.json` — the Provider Code Spec emitted by `ruletypesgen`.
  Regenerated, not hand-edited.

## How to regenerate

```
make generate
```

That runs both tracks:

1. **Track A (simple resources).** `tfplugingen-openapi` turns
   `control-api.yaml` + `generator_config.yml` into `spec.json`, then
   `tfplugingen-framework` turns that into Go schema + model code.
2. **Track B (rule families).** `ruletypesgen` reflects the control rule types
   into `rules_spec.json`, then `tfplugingen-framework` turns that into Go code
   the same way.

Both write to `internal/provider/codegen/resource_<name>/` (pinned tool
versions, fetched via `go run`).

The output is committed so changes are reviewable and a future CI check can
assert that regeneration produces no diff.

## Scope and caveats

This is deliberately limited right now:

- **Two tracks.** Simple resources (`app`, `namespace`, `queue`) generate from
  the OpenAPI spec. The integration rules use an OpenAPI `oneOf` + discriminator
  that `tfplugingen-openapi` cannot handle, so the moderation and before-publish
  rule families are generated from the in-repo `control` types instead via
  `ruletypesgen` (see the strategy doc). The webhook/firehose rule families are
  not generated yet.
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
