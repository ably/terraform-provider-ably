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
5. Register the resource in `internal/provider/provider.go`.
6. Add an example under `examples/resources/`, a template under
   `templates/resources/`, and run `tfplugindocs` to generate the doc.
7. Add an acceptance test and a unit test for any preserve-from-plan / write-only
   handling. Run `make test`.

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
