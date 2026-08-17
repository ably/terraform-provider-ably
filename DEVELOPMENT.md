# Local Development

Update `~/.terraformrc` file with overrides:

```
plugin_cache_dir   = "$HOME/.terraform.d/plugin-cache"
disable_checkpoint = true

provider_installation {
  dev_overrides {
      # This should be the path to where the repository is cloned
      "ably/ably" = "/path/to/terraform-provider-ably",
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

Build your changes from the repository root with:

```
go build
terraform -chdir=examples/playground init
terraform -chdir=examples/playground plan
```

Generate docs for this provider by installing [`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs) and running `tfplugindocs` from the repository root.

## Testing

- `make test` runs the unit tests plus the full acceptance suite against an
  in-process fake Control API (`internal/provider/fake_control_api_test.go`). No
  credentials or network required, this is the loop to run on every change.
- `make testacc` runs the acceptance suite against a real Control API. Set
  `ABLY_ACCOUNT_TOKEN` (and optionally `ABLY_URL`). CI points it at staging.

## Code generation

Schema and model code is generated from the Control API spec. See
[`codegen/README.md`](codegen/README.md) for the pipeline. Regenerate with:

```sh
make generate
```

There are two tracks. Simple resources (`app`, `namespace`, `queue`) generate
from the vendored OpenAPI spec via `tfplugingen-openapi`. The integration-rule
families use an OpenAPI `oneOf` the generator can't read, so they are generated
from the in-repo `control` rule types by `codegen/ruletypesgen`, with field
descriptions sourced from the spec. Generated code lands under
`internal/provider/codegen/` and is committed.

Generation produces schema + model only. CRUD wiring to the `control` client
stays hand-written.

## Adding a new integration rule

1. Add the rule's control types to `control/rule_types_*.go` (create/patch/
   response bodies and target), if they don't already exist.
2. Add the rule to the `rules` list in `codegen/ruletypesgen/main.go`, mapping
   the resource name and its OpenAPI schema name (for descriptions).
3. Run `make generate`. This produces `internal/provider/codegen/resource_<name>/`.
4. Write the resource shim in `internal/provider/` (see "Porting" below for the
   pattern): a `Schema()` that adopts the generated schema, the CRUD methods
   delegating to the `control` client, and `Metadata`/`ImportState`.
   For a moderation or before-publish rule, take the shared plumbing in
   `internal/provider/before_publish_rules.go` instead of writing CRUD by hand:
   supply the model plus a create-body function and a response-mapping function,
   and delegate the four CRUD methods to `beforePublishCRUD`. Reference example:
   `resource_ably_rule_tisane.go`. Do **not** copy a webhook rule for these
   families; the generic `AblyRule` plumbing bakes in `source` and
   `request_mode`, which these rules do not have and the API rejects.
5. Register the resource in `internal/provider/provider.go`.
6. Add an example under `examples/resources/`, a template under
   `templates/resources/`, and run `tfplugindocs` to generate the doc.
7. Add an acceptance test and a unit test for any preserve-from-plan / write-only
   handling. Run `make test`.

## Adding a data source

The Control API has no fetch-by-ID endpoint for apps, keys, namespaces or queues,
only list endpoints, so every entity has a plural data source generated from its
list response and a singular one that lists and filters locally.

1. Add the read path to the `data_sources` block in `codegen/generator_config.yml`
   and run `make generate`. That produces
   `internal/provider/codegen/datasource_<name>/`.
2. Write the data source in `internal/provider/data_source_ably_<entity>.go`:
   a model mirroring the generated element attributes, the plural data source
   adopting the generated schema, and the singular one built with
   `elementAttributes` (see `data_sources.go`) so both serve the same generated
   attribute set. Use `findOne` for the lookup: it enforces exactly one of
   id/name and refuses to guess when a name matches more than one record.
3. Register both in `DataSources()` in `internal/provider/provider.go`.
4. Name them after the existing resource, not the spec. The key data sources are
   `ably_api_key`/`ably_api_keys` because the resource is `ably_api_key`, and the
   key permissions map is `capabilities` for the same reason.
5. Add an example under `examples/data-sources/`, a template under
   `templates/data-sources/`, run `tfplugindocs`, and add an acceptance test.

## Porting a resource onto generated code

The reference example is `ably_rule_bodyguard`
(`internal/provider/resource_ably_rule_bodyguard.go`). The pattern:

1. `Schema()` calls the generated `…ResourceSchema(ctx)` as its base. The
   generated schema already carries the attribute set, types, nesting,
   sensitivity, descriptions, and (sourced from the spec or the overrides table
   in `ruletypesgen`) enum validators, defaults and plan modifiers.
2. Strip the generated `CustomType` from any nested blocks (`attr.CustomType =
   nil`) so a plain-struct tfsdk model reflects cleanly. (Alternatively, adopt
   the generated model and its value types; the plain-struct approach keeps the
   CRUD simpler.)
3. Set the resource-level `MarkdownDescription`.
4. Leave the model and CRUD hand-written. Wiring to the `control` client is not
   generated.
5. `make test` must stay green; the fake exercises the full CRUD/import/diff.

If a rule needs metadata the spec doesn't carry (for example the `status` enum,
or a particular plan modifier), add it to the overrides table in
`codegen/ruletypesgen/main.go` rather than patching it in `Schema()`, so every
rule benefits and ports stay near-mechanical.
