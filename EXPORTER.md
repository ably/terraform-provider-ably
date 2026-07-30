# Exporting an existing Ably account

`cmd/ably-exporter` generates Terraform configuration for the resources already
in an Ably account, so an account built by hand can be brought under Terraform
without writing the config yourself.

It needs an account token and nothing else. Every Control API call it makes is a
read.

```sh
export ABLY_ACCOUNT_TOKEN=...
make export-account                 # writes ./ably-export
cd ably-export
terraform init
terraform plan                      # should report only the imports
```

`imports.tf` adopts the resources into state. Delete it once `terraform apply`
has run the imports.

## Getting it

Each provider release carries an `ably-exporter_<version>_<os>_<arch>.zip` asset
for Linux, macOS and Windows on amd64 and arm64, so the exporter matches the
provider version it generates config for.

The exporter is not listed in the release's `SHA256SUMS`. That file is what the
Terraform Registry verifies the provider against.

To build from a checkout:

```sh
make exporter                       # writes ./bin/ably-exporter
```

## What it writes

```
ably-export/
  provider.tf          required_providers and the provider block
  app_<name>.tf        one file per app: the app and everything inside it
  variables.tf         only with -secrets=vars
  imports.tf           import blocks, unless -imports=false
```

Labels come from Ably names, so `ably_app.chat_service` rather than an opaque ID.
Attributes holding another resource's ID are written as references
(`app_id = ably_app.chat_service.id`), so the config carries its own dependencies
and is portable between accounts.

Files holding credentials are written `0600`.

## Flags

| Flag | Default | What it does |
| --- | --- | --- |
| `-token` | `$ABLY_ACCOUNT_TOKEN` | Account token. |
| `-out` | `./ably-export` | Output directory. |
| `-app` | all apps | Only export this app, by ID or name. Repeatable, or comma-separated. |
| `-secrets` | `inline` | `inline`, `vars` or `omit`. See below. |
| `-imports` | `true` | Generate `import` blocks. |
| `-single-file` | `false` | Write everything to `main.tf`. |
| `-provider-version` | `~> 1.0` | Version constraint in `required_providers`. Empty omits it. |
| `-force` | `false` | Write into a directory that already holds `.tf` files. Clears files a previous export wrote; leaves hand-written ones alone. |

## Secrets

- `-secrets=inline` (default) writes what the API returned. Accurate and plans
  clean, but the output holds credentials.
- `-secrets=vars` replaces sensitive values with references to sensitive
  variables declared in `variables.tf`. Safe to commit, but you supply the values
  before it will apply.
- `-secrets=omit` leaves sensitive attributes out, with a comment where each one
  would have gone.

Whichever you pick, the exporter reports how many sensitive values it touched.

## Before you apply

- **Some values cannot be exported.** The Control API accepts them on write and
  never returns them: `ably_app.fcm_key` and the APNs credentials, AWS
  `secret_access_key`, Kafka SASL credentials, Pulsar `tls_trust_certs`. Where
  such a value is required, the exporter writes a `# TODO` (or a variable, under
  `-secrets=vars`) and lists it in the summary, rather than writing the empty
  string the API hands back and wiping the live credential on apply. Optional ones
  are simply absent and will be cleared unless you fill them in. Read the plan.
- **Revoked API keys are skipped.** The provider reads them as gone, so exporting
  one gives config for a resource that isn't there.
- **Unsupported rule types are skipped with a warning**, named by rule ID and
  type rather than dropped silently.
- **`# TODO` means something needs you.** Either a value the API withholds, or
  config the provider rejected. Both are listed in the summary.

## How it keeps up with the provider

The exporter carries no copy of the schema. It runs the provider in-process and
drives it over the protocol-v6 RPCs Terraform uses:

1. `GetProviderSchema` for the resource schemas.
2. `ImportResourceState` and `ReadResource` per resource, which is what
   `terraform import` does, so the state rendered is the state Terraform would
   hold.
3. `ValidateResourceConfig` on the config it is about to write.

So **a new attribute on an existing resource needs no change to the exporter**,
and what it writes is what the provider says it will accept.

Two things do need a change, and both fail loudly if missed:

- **A new integration rule** needs a line in `RuleTypeResources`
  (`internal/provider/rule_types.go`) mapping its `ruleType` to its resource type.
  Every rule arrives from the same `GET /apps/{id}/rules`, so that mapping is the
  only way to tell which resource owns one. `TestRuleTypeResources` catches a
  missing entry.
- **A new family of resources** (not an app, key, namespace, queue or rule) needs
  a lister in `internal/exporter/discover.go` and its type in
  `SupportedResourceTypes`. `TestSupportedResourceTypesCoverage` catches a gap.

### The validate-and-repair pass

Validators are not part of the protocol schema, so the exporter asks about them
instead. `ably_namespace` is the live example: it has an `identified` attribute
and a deprecated `authenticated` alias that conflict in config, and a read
returns both.

The exporter validates what it generated and, when the provider objects, removes
the attribute and asks again. It only ever removes **deprecated** attributes,
which have a modern equivalent carrying the same value. Anything else is left as
the API returned it and reported with a `TODO`: a rejected value is a real
disagreement between provider and API, and deleting config to satisfy a validator
would hide it. Every removal is reported.

That handles new conflicts and new deprecations without any change here.

## Testing

`make test` covers the exporter against a read-only in-process fake Control API,
with no credentials and no network. It checks the generated HCL parses, that
computed attributes never appear (Terraform rejects those), that references and
import IDs are right, all three secrets modes, withheld required values, stale
file cleanup, file permissions, and the repair pass.

For an end-to-end check, export a real account and run `terraform plan`: a
correct export reports imports and no other changes.
