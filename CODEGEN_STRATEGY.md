# Generating the Ably Terraform Provider from the Control API spec: build vs buy

## 1. Problem and the constraint that drove hand-writing

We document the Control API with an OpenAPI 3.0.1 swagger spec (`swagger/v1/swagger.yaml` in the `website` repo, ~5,852 lines). We do not want to keep hand-maintaining three layers of code that all ultimately derive from that spec:

1. the in-repo `control` package (Go client: methods + types that talk to the API),
2. the Terraform models (tfsdk structs that mirror the control types),
3. the Terraform resources/data-sources/import wiring.

The constraint that made us hand-write the client in the first place still holds: when the team tried generators, the output was clunky. The research confirms exactly where "clunky" comes from, and it is not vague taste. The spec models 23 integration-rule variants as three parallel `oneOf` + `discriminator` unions (`rule_post`, `rule_patch`, `rule_response`) with discriminator values that are not valid identifiers (`http/ifttt`, `aws/lambda/before-publish`, `ingress-postgres-outbox`). The response union adds two synthetic variants (`webhook`, `unsupported`) with no request counterpart. Timestamps are `type:number` on rule responses but `type:integer` on app/key/namespace responses. Nullable fields are pervasive and many booleans carry the contradictory `default:false` + `nullable:true` pair. There is a multipart `pkcs12` upload, top-level array responses, and endpoints that return no body. On top of that the spec is not faithful to the shipped product: the client deliberately omits `/me/accounts` and `/help`, drops two fields the spec marks required on `MeToken`, and carries domain knowledge in comments that the spec does not have (Pulsar `authenticationMode: "token"` not `"jwt"`, write-once keys, delete cascades).

So the honest framing is: the spec is good enough to scaffold the simple resources and useless-to-dangerous for the rules subsystem. Any decision has to treat those two worlds separately.

## 2. Priority order: generator first, agent-readiness as the fallback

The decisions below are not a free-for-all between "generate" and "hand-write". There is a strict priority, applied at every layer and every stage:

1. **Generate it if we possibly can.** A generator is the preferred outcome everywhere, always. Even where this document concludes hand-writing wins today, that conclusion is provisional: when the blocking tool matures (e.g. `oneOf` support lands in `tfplugingen-openapi`, or we accept the oapi-codegen leaf-DTO hybrid for the client), we should move that surface back to generation. Generation is the goal; hand-writing is a concession we make only where generation is genuinely not feasible right now.
2. **Where generation genuinely isn't feasible, make the hand-written surface trivial for AI agents to extend.** This is the fallback, not a parallel option. The aim is to get as close to the leverage of a generator as possible (low-effort, low-error addition of new functionality) without one, so AI agents can do the development. That makes the hand-written rule and resource code a first-class agent-extensibility target.

Concretely, "agent-ready" means:

- **Clear, uniform examples.** Every rule/resource follows one identical pattern with low variance, so an agent copies the right thing. The odd-one-out is the enemy: one rule that does it differently teaches the agent the wrong lesson.
- **A fast, hermetic feedback loop.** One command that runs compile + lint + unit tests with no API credentials and returns unambiguous pass/fail. Agents are only as good as the signal they can get unattended, and today the acceptance tests need a real Control API (`internal/provider/e2e_test.go`), which an agent cannot run in-loop. This is the highest-value gap to close.
- **Loud tests on the footguns, not the happy path.** The silent-correctness traps this document already flags (PATCH clobbering server state, the sensitive-attribute "inconsistent value" diffs, the `bool` vs `*bool` namespace split) are exactly what agents get wrong, because they pattern-match the happy path and will not reason about partial-update semantics unless a test forces them to. "Let agents do the development" only works if those traps fail by default in tests. Agent-priority has to come bundled with footgun-priority, or we just ship the silent bugs faster.
- **Strong, task-shaped docs.** Not prose about the architecture, but a runbook: "to add a new rule, touch these N files in this order, here is the reference change." Better still, the scaffolding generator stamps the skeleton so the agent only fills the gaps.
- **A spec-drift signal.** A CI check that diffs the upstream swagger against what the provider implements, turning "not feature complete" from a vibe into a worklist an agent can pick up. Without it, an agent never knows there is work to do.

Note the convergence: every one of these (uniformity, scaffolding, hermetic tests, runbook) is also exactly the substrate a custom generator needs. The agent-readiness work is not a detour away from the generator, it is the groundwork for it. The two strategies reinforce each other rather than competing, which is why the fallback is safe to invest in even while generation stays the priority.

## 3. Build-vs-buy decision, per layer

### Layer 1: the `control` Go client — hand-written stays

Decision: keep `control/` hand-written. Do not adopt a generator wholesale.

The four candidates surveyed:

- **go-swagger** is disqualified immediately. It is OpenAPI 2.0 only and our spec is 3.0.1. `oneOf` does not even exist in its model. Down-converting to 2.0 would throw away the discriminator unions entirely. Do not pursue.
- **openapi-generator (Go target)** supports 3.x but produces the least idiomatic Go, mangles initialisms (`Id`/`Url`/`Apns` instead of `ID`/`URL`/`APNS`), and has the most fragile `oneOf`/discriminator handling of the lot, with known unmarshal panics about data matching more than one schema. That is precisely the shape this spec leans on hardest. It also has no clean seam for our `retryablehttp` transport. Poor fit.
- **oapi-codegen** is the strongest single candidate. It drops our existing `go-retryablehttp` client in cleanly via its `HttpRequestDoer` `Do(*http.Request)` interface, so the bespoke 5xx-only `retryPolicy` in `control/client.go` survives verbatim. The Post/Patch split is free because the spec already separates `http_rule_post` from `http_rule_patch`. But it mangles initialisms unless we maintain `x-go-name` overrides or a name-normalizer, its multipart support is the tool's rough edge (open issues into 2026), and its discriminated-union output is a `json.RawMessage` wrapper with `ValueByDiscriminator()` that is less type-safe and less ergonomic than the typed rule structs we have today.
- **ogen** models the hard features best: real Go sum types with native discriminator dispatch, the best initialism handling, solid multipart. The cost is rigidity. Its aggressive validation will likely force spec cleanup before it generates at all, and it bakes in its own error/tracing/client shape, so wedging in our retry policy means adopting ogen's model rather than keeping ours.

No single tool wins every axis, and the things generators do worst are exactly our load-bearing decisions: collapsing the 24-member `rule_response` `oneOf` into one `RuleResponse` with `Target interface{}` (`control/types.go:479-491`), `CreateRule`/`UpdateRule` taking `any` (`control/rules.go:22,38`), the per-resource `float64` vs `int64` timestamp split, the `NamespacePost` `bool` vs `NamespacePatch` `*bool` distinction tied to `default:false`+nullable, the hand-framed multipart pkcs12 upload, and the domain knowledge in comments. The first three and the comments are not realistically reproducible from the spec at all. The correctness-affecting ones (Post/Patch pointer+omitempty, the bool split, the stats nil-means-defaults builder, multipart) are reachable only with heavy config, and getting them wrong produces silent bugs (PATCH clobbering server state), not just ugly names.

The generator upside here is small because the one thing generators do for free (the Post/Patch split) we already get from the spec structure, and the boilerplate they would save is dwarfed by the hand-patching they would require forever after, plus owning a codegen pipeline in CI.

If we ever want to cut maintenance on this layer, the realistic move is a **hybrid, not a replacement**: use oapi-codegen to generate only the leaf variant DTOs (the 23 `rule_post`/`rule_patch`/`rule_response` concrete structs) with `x-go-name` annotations for initialisms, and keep the hand-written `client.go` transport, retry policy, and typed union dispatch. That captures most of the boilerplate without surrendering the parts generators do badly. But that is a later optimisation, not the recommendation now.

### Layer 2: the Terraform models (tfsdk schema + model structs) — off-the-shelf with post-processing, for the four simple resources only

Decision: use the HashiCorp chain (`tfplugingen-openapi` -> `tfplugingen-framework generate`) to produce schema + model Go for `ably_app`, `ably_key`, `ably_namespace`, `ably_queue` only. Generate nothing for rules from the OpenAPI front-end. Hand-write or script the rule models.

Why this split: `tfplugingen-openapi` flatly cannot handle `oneOf` + discriminator. HashiCorp issue #94 has been open since Nov 2023 with no maintainer commitment, and issue #82 confirms resources containing `oneOf`/`anyOf` children get silently skipped with warnings. The rules are 17 of 19 resources. Pointing the OpenAPI generator at them gets us nothing. But the simple resources are plain object schemas, and `ably_app` alone is ~25 attributes with sensitive flags and defaults, so generating its schema/model from the spec is genuine value that then tracks the spec automatically.

Carry the caveat honestly: every tool in this chain is **tech preview**. `tfplugingen-openapi` has had no release since v0.3.0 (Jan 2024), `tfplugingen-framework` is at v0.4.1 (Sep 2024), `terraform-plugin-codegen-spec` at v0.2.0 (Sep 2024). I am not betting the rule pipeline on a near-future `oneOf` fix.

For the rule models, the escape hatch is `terraform-plugin-codegen-spec`'s Go bindings: a small Go program walks the in-repo `control` rule types (`control/rule_types_*.go`) and emits one Provider Code Spec entry per rule variant, which then feeds the same `tfplugingen-framework generate` stage. This is more reliable than waiting on #94 and mirrors the existing hand-flattened `GetRuleSchema` approach.

I am explicitly rejecting **Speakeasy** as the default here. It is the only tool that claims `oneOf` + full CRUD, but it generates and owns its own SDK runtime, which directly fights the in-repo `control` client we just migrated to (commit `47f5ba8`), and it wants `x-speakeasy-entity` annotations on a spec owned by another repo. It is worth a time-boxed spike, not an adopt.

### Layer 3: resources / data-sources / CRUD / import — custom generator for the rule fan-out, hand-written stays for the rest

Decision: no off-the-shelf tool produces this layer for us. Build a small custom generator for the per-rule resource files; leave everything else hand-written.

The HashiCorp chain emits schema + model only. `generate` does not wire CRUD; `scaffold` emits unwired stubs. So all wiring to the `control` client is manual regardless of tool. The provider-layer audit makes the shape clear: the 15 thin per-rule `.go` files (`resource_ably_rule_http.go` etc., ~70-105 lines each, ~1,200 lines total) are near-pure boilerplate. Each is a struct, interface assertions, a `Schema()` calling `GetRuleSchema` with a literal target map, `Metadata`, `Provider()`/`Name()`, and five one-line CRUD/Import delegations to the already-written generic `CreateRule[T]`/`ReadRule[T]`/etc. over `AblyRuleDecoder[T]`. That is the clean generation win, and a small templated generator keyed off the control rule types produces it.

The value-add code that a generator must **not** try to own lives in three places per rule: the tfsdk target model struct in `models.go`, and the `GetPlanRule` / `GetRuleResponse` switch cases. Those carry per-field type/helper choices, the `RuleType` string literals (`http/cloudflare-worker`, `ingress/mongodb`), and roughly seven non-generatable special cases: AWS auth mode branching with `secret_access_key` pulled from prior plan, `webhookEnveloped` forcing `enveloped=false` in batch mode, the kinesis/sqs `request_mode != single` guard, Pulsar `tls_trust_certs` write-only preservation, and amqp/external `url`/`exchange`/`message_ttl` preservation to dodge the "inconsistent values for sensitive attribute" error. A generator must take a per-field mapping table as input and leave explicit escape hatches for these, or it will reintroduce perpetual diffs.

For the four non-rule resources, the load-bearing logic is hand-written and field-specific: read-back-via-List because there is no get-by-id, write-only secret preservation (`fcm_key`, `apns_*`), the capability map<->set conversion with `SortSetsInMap`, namespace conditional batching/conflation with a Create-vs-Update difference, queue immutability, RFC3339-vs-Int64 timestamp handling, the provider `Configure` bootstrap, and `TypeName` mismatches (`ably_api_key`, `ably_ingress_rule_mongodb`). Generation here would reproduce skeletons and leave all the load-bearing code by hand. Not worth a generator. Keep it hand-written.

## 4. Recommended end-to-end pipeline

Two source-of-truth tracks feeding one code-generation back end, with a curated intermediate model to absorb the spec-fidelity gaps.

**Spec hygiene (shared, up front).** The swagger lives in the website repo, so we do not edit it in place. We maintain an **overlay/patch file in this repo** that the pipeline applies before any generation: add `x-go-name` for initialisms, drop the `/me/accounts` and `/help` endpoints we do not ship, and set `ignores` in the `tfplugingen-openapi` config to suppress the rule `oneOf` so the simple resources generate without aborting. The overlay is our seam for keeping the website spec untouched while the generators see a spec they can swallow.

**Track A: simple resources (app, key, namespace, queue).**
1. `swagger.yaml` + overlay + a hand-written `generator_config.yml` (mapping each resource to its create/read/update/delete path+method) -> `tfplugingen-openapi generate` -> `provider_code_spec.json`.
2. `provider_code_spec.json` -> `tfplugingen-framework generate resources|data-sources` -> Go schema + model files.
3. Hand-written (stays): CRUD wiring to `control`, timestamp transforms, write-only field preservation, read-back-via-List, plan modifiers and defaults, `Configure` bootstrap.

**Track B: rules (15 resources).**
1. A small Go program using `terraform-plugin-codegen-spec` Go bindings walks the in-repo `control/rule_types_*.go` types and emits one Provider Code Spec entry per rule variant. This is where we solve the `oneOf` problem: we never ask any OpenAPI tool to read the discriminated union. The in-repo control types are already the curated, per-family-correct model (ingress/before-publish/moderation families correctly drop `requestMode`/`source`), so they are a better source of truth for the TF schema than the spec is.
2. That spec JSON -> `tfplugingen-framework generate` -> rule schema + model Go.
3. A **custom templated generator** (also keyed off the control rule types plus a per-field mapping table) emits the 15 thin per-rule resource files that delegate to the existing generic `CreateRule[T]` plumbing, with explicit hook points for the ~7 special cases.

**Handling the spec-fidelity gaps:** the overlay deletes the unshipped endpoints and the response-only `webhook`/`unsupported` variants are simply never modelled (the control client's `Target interface{}` already absorbs them; our TF rule generation is driven from control types, not the spec, so they never appear). The `MeToken` required-field divergence and the timestamp-type split stay encoded in the hand-written control types, which are the source of truth for Track B. In short: the spec drives the four simple resources, the curated control package drives the rules.

## 5. Phased implementation plan

**Phase 0 — Spike to de-risk (1-2 days). Do this first.** Two narrow spikes, both cheap and both gating:
- Run `tfplugingen-openapi` over the overlaid spec for `ably_app` only, with rules ignored, and confirm it produces a spec JSON, then `tfplugingen-framework generate` produces compilable schema/model that matches the existing `ably_app` contract. This validates the whole Track A chain on the hardest simple resource before we commit.
- Write a throwaway Go program that emits a Provider Code Spec entry for one rule variant (say `http`) from its control type, and run `tfplugingen-framework generate` on it. This validates that the codegen-spec bindings are a viable front end for rules.

If either spike fails badly, we fall back to fully hand-written for that track and lose little.

**Phase 0c — Agent-extensibility spike (1-2 days). The most important spike.** Take the codebase exactly as it is today, with no special preparation, and have an AI agent add one genuinely-missing resource end-to-end: schema, model, CRUD wiring, import, docs, and acceptance test. Watch precisely where it stumbles, missing examples, weak docs, or the inability to verify itself without real API credentials. That is direct evidence for what the agent-readiness backlog actually is, rather than us guessing, and it tells us how close the hand-written fallback can get to generator-level leverage.

Concrete target: the `control` package already fully supports before-publish rules (`BeforePublishWebhook*`, `BeforePublishAWSLambda*` in `control/rule_types_before_publish.go`) and the five moderation rules (`HiveTextModelOnly`, `HiveDashboard`, `BodyguardTextModeration`, `TisaneTextModeration`, `AzureTextModeration` in `control/rule_types_moderation.go`), but NONE of them are exposed as Terraform resources, and the provider currently ships ZERO data sources (`DataSources()` returns an empty slice in `provider.go:181`). So the agent's task is pure Terraform-layer work over an already-written client type, following the established per-rule pattern. Pick one moderation rule as the target: these drop `requestMode`/`source`, so they are a representative variant rather than a vanilla `http` clone, which makes the spike honest about whether the pattern holds for non-identical cases. Whatever the agent struggles with becomes the agent-readiness backlog (and, not coincidentally, the requirements for the Phase 3 custom rule generator).

If Phase 0/0a/0b show generation is blocked for a track, the agent-readiness work for that track becomes the funded path per the priority order in Section 2, while we keep generation on the table for when the blocking tool matures.

**Phase 0c results (run 2026-06-08).** We ran this spike. An AI agent, given no coaching beyond "add the Bodyguard text moderation rule following the repo's conventions", produced a compiling, unit-tested, documented, registered `ably_rule_bodyguard` resource. `go build`, `go vet`, `gofmt`, the provider unit tests, `golangci-lint`, and `tfplugindocs generate` all ran green unattended. The acceptance test could not run: it needs a real Control API token, and that is the finding that matters most.

Verified against the code, not just the agent's self-report:

- **The generic rule plumbing is silently webhook-only, and this is the keystone problem.** `GetRuleSchema` (`internal/provider/rules.go:704`) hardcodes `source` as a `Required` nested attribute and adds `request_mode`, and the generic `AblyRule` decoder (`models.go:203-204`) bakes in `RequestMode`/`Source`. `BodyguardTextModerationRulePost` has neither; it has `BeforePublishConfig`/`InvocationMode`/`ChatRoomFilter`/`Target`. So the dominant, most-copyable pattern (13 of the rule resources delegate to `CreateRule[T]`) is a trap for the entire moderation/before-publish family: copying it yields a resource with a bogus required `source` block, a `request_mode` the API rejects, and no `before_publish_config`, that still compiles and looks right. The agent avoided it only by reading the control types closely; a cheaper agent would not. `internal/provider/ingress_rules.go` is the existing precedent for a sourceless family, but nothing signposts it. Fixing this (a second generic family, or at minimum loud "webhook-only" guards) is the single highest-value change, and it is the same refactor a Phase 3 generator would need.
- **`RuleResponse` drops the moderation fields on read.** `control/types.go` `RuleResponse` carries no `beforePublishConfig`/`invocationMode`/`chatRoomFilter`, and `apiKey` is write-only. A naive read maps them to null and Terraform throws "inconsistent result after apply" plus sensitive-attribute diffs. The need is documented only in amqp/pulsar comments in `rules.go`. The agent wrote a preserve-from-plan block and a loud unit test for it, exactly the footgun-test pattern Section 2 calls for.
- **No hermetic end-to-end loop is the real blocker.** Everything except the live round-trip verified unattended. The one test that proves the resource actually works (CRUD, import, whether the API returns the moderation fields) needs a credential the agent cannot have, so its read logic is inferred, not proven. This is the gap that stops us trusting agent-authored resources without a human in the loop, and it motivates the local end-to-end test environment in Section 7.
- Minor: resource naming and the `bodyguard/text-moderation` discriminator were guessed from comments (no importable constant, no written naming rule); `provider.go` registration has no checklist, so an agent can ship an unregistered resource.

Net: the thesis holds, an agent can extend this codebase, but unsupervised correctness depends on two things the spike shows are missing: the structural fix to the rule plumbing, and a credential-free round-trip test.

**Phase 1 — Track A simple resources (3-5 days).** Build the overlay, the `generator_config.yml`, and wire `tfplugingen-openapi` -> `tfplugingen-framework generate` for app/key/namespace/queue into a `make generate` target. Diff the generated schema/model against the current hand-written ones and reconcile. CRUD wiring stays hand-written but now sits on top of generated schema/model. Effort is dominated by reconciling generated attribute metadata against our existing defaults/sensitive flags.

**Phase 2 — Track B rule schema/model generation (4-6 days).** Promote the Phase 0 rule spike into a real generator: walk all `control/rule_types_*.go`, emit spec entries for all 15 variants, generate schema/model. Carry the per-field type/helper mapping table as explicit generator input.

**Phase 3 — Track B custom per-rule resource generator (3-5 days).** Template the 15 thin resource files delegating to the existing generic CRUD, with declared escape hatches for the ~7 special cases. This is the largest line-count win and the lowest-risk generation because the target files are already uniform.

**Phase 4 — Pipeline, CI, and agent-readiness (2-3 days).** Single `make generate`, mark generated files clearly, add a CI check that regeneration produces no diff, document the overlay, and decide and document the regeneration cadence. Land the agent-readiness deliverables alongside, since they fall out of the same work: a hermetic `make test` (compile + lint + unit, no credentials) that an agent can run in-loop; a spec-drift CI check that diffs the upstream swagger against implemented resources and fails with the list of what is missing; and a short "add a new rule / resource" runbook pointing at the canonical example and the files to touch. If the agent spike surfaced specific footguns, add the loud tests for them here.

**Layer 1 (the control client) gets no phase.** It stays hand-written. Revisit the oapi-codegen leaf-DTO hybrid only if rule-variant churn becomes a real maintenance cost.

Rough total: ~3 weeks of focused work, front-loaded by a 1-2 day spike that can kill either track early.

## 6. Key risks and open questions

- **Everything in Track A is tech preview.** `tfplugingen-openapi` has not shipped since Jan 2024. If it stays stalled or the `tfplugingen-framework` spec format moves under us, we own a pipeline built on unmaintained tools. Mitigation: the generated output is plain framework Go we can fork and keep if the tools die. We are not locked in, but we would lose the regeneration benefit.
- **The overlay is cross-repo coupling.** The swagger lives in the website repo. If it changes shape (new rule family, renamed schema) our overlay and `generator_config.yml` can break silently. The CI no-diff check catches drift on our side but not upstream intent. Open question: do we want a contract test that fails when the upstream spec changes in a way the overlay does not cover?
- **The rule special cases are where regressions hide.** The seven hand-written carve-outs (AWS auth, batch `enveloped`, kinesis/sqs guard, write-only preservation) are exactly the bugs that produce perpetual diffs or sensitive-attribute errors. If the generator ever overwrites a hook point, behaviour regresses quietly. Mitigation: generated rule files must call out to hand-written hooks, never inline the special logic.
- **`tfplugingen-openapi`'s `allOf` behaviour is undocumented.** `DESIGN.md` does not cover it. If the simple-resource schemas use composition we may hit unspecified behaviour in Phase 1. The spike should check for `allOf` usage in the four simple schemas.
- **Confirm `ogen`/`oapi-codegen` claims by generating, not from docs.** The research flagged `ogen`'s exact `WithClient` signature and APNS-initialism output as medium-confidence. This only matters if we ever pursue the Layer 1 hybrid, but do not commit to oapi-codegen leaf-DTO generation without first generating and eyeballing the union and initialism output.
- **The generator-first commitment needs a trigger, or it rots into permanent hand-writing.** The priority order in Section 2 says we move surfaces back to generation as the blocking tools mature, but "mature" is not self-announcing. Open question: do we want a tracked check on `tfplugingen-openapi` issue #94 (`oneOf` support) and a periodic re-spike, so the agent-readiness fallback does not quietly become the permanent answer by default? The risk is that agents make hand-extension feel cheap enough that we stop reaching for the generator even once it becomes viable.
- **Open question: is the codegen-spec front end for rules worth it over staying hand-flattened?** The per-rule files are the win; the schema/model generation in Phase 2 is more marginal because `GetRuleSchema` already abstracts it well. If Phase 0's rule spike is awkward, we can ship Phase 3 (per-rule file generation) on top of hand-written schema/model and skip Phase 2 entirely.

---

## 7. Local end-to-end test environment

Phase 0c showed the real blocker to unsupervised agent development is not docs or examples, it is that nothing lets an agent prove a resource actually works without a production-grade credential. An agent can write a resource that compiles, lints, and unit-tests green, and still be confidently wrong about what the API returns on read. We need a credential-free way to close that loop, and a heavier way to prove the provider and the real service actually agree.

### What "the Control API locally" actually involves

It is not one service. The local Control API is two systems wired together:

- **ably/website** serves the Control API HTTP surface (the `/v1` routes), does the JWT bearer-token auth, and keeps a Postgres mirror of accounts/apps/keys. Locally, with no `CONTROL_API_HOST` set, it mounts the Control API under `http://localhost:<rails-port>/api/v1`. It does not enforce HTTPS for localhost.
- **ably/realtime** (the "farm") is what actually provisions resources. The website's provisioner calls an Admin API via ActiveResource, and with `REALTIME_ENV=local` (already the default in `apps/website/.env`) that Admin API is the farm's, at `http://localhost:8090` with basic auth `admin`/`admin`. The farm is backed by Cassandra.

So every resource-creating Control API call (apps, keys, namespaces, queues, rules, which is the entire provider) flows website -> Admin API -> farm. A website on its own cannot provision; the farm is a hard dependency. There is no standalone or mock Admin API in either repo, and nothing today wires the farm's datastore to a Control API stand-in. This is the heavy part, and it is undocumented as an end-to-end path.

### The recommendation: three tiers, not one

Making the full website+realtime stack the loop an agent runs on every change is the wrong target. It is heavy and, as below, partly human-gated. Split the need into three tiers:

**Tier 1, the hermetic fake (the agent's inner loop, build this first).** The `control/` package already has exactly the right pattern: `control/testutil_test.go` stands up an `httptest` server, points a real `control.Client` at it, and runs fully offline with no token. The provider's acceptance tests don't do this yet, but they don't need a source change to: `provider.go` already takes `ABLY_URL`, the client does a plain `BaseURL + path` with no host allow-listing, and the account ID is discovered at runtime via `/me`. So we add a small stateful in-process fake Control API in `internal/provider/` (implements `/me` plus CRUD for the resources), set `ABLY_URL=srv.URL` and a dummy token, and the existing `resource.Test` flows run unchanged.

- Catches: schema validation, plan/diff stability, full CRUD wiring, import ID parsing, attribute mapping, the negative `ExpectError` cases, and crucially the footguns, including the moderation read-preservation bug from Phase 0c (the fake returns what the real API returns, so a missing preserve-from-plan shows up as an inconsistent-result failure). All credential-free, offline, runnable on forks and by an agent unattended.
- Misses: whether the real API agrees. A fake that mis-models the API passes happily, so it must be kept honest (see below).
- This is the single change that unblocks unsupervised agent development, and it needs no live stack.

**Tier 2, staging-backed acceptance (the real-behaviour gate, already exists).** CI already runs the full acceptance suite with `TF_ACC=1` against `https://staging-control.ably-dev.net/v1` using a repo-secret token. Every PR provisions and destroys real resources in staging. Keep this. It is the proof that the provider and a real Control API agree, and it is what keeps the Tier 1 fake honest.

**Tier 3, the full local stack (heavyweight, for deep debugging and offline real-behaviour).** Stand up website+realtime locally when you genuinely need a real Control API offline. The runbook is below. The key realism: an agent can use this stack once it is warm by pointing `ABLY_URL` at it, but it cannot reliably cold-start it unattended (see blockers). So the model is: a human or a session-level setup brings it up, and the agent then drives it.

### Tier 3 runbook (and its blockers)

Order matters: the farm must be up before the website seeds, because seeding provisions through the Admin API.

1. **Resolve the Redis collision first.** Both repos run Redis via Docker on host port 6379. In `apps/website/.env` set `REDIS_PORT=6380` and (re)run the website's `docker-compose up -d`. The website README documents exactly this for running against a local realtime service.
2. **Start the realtime farm** (from `ably/realtime`): `mise run shared-up` (Cassandra, Redis, RabbitMQ), `mise run cassandra-setup` (one-time), then `mise run farm -- --daemon --stop-existing --daemon-timeout=60`. Verify with `mise run farm-status` (expects `{"running":true,"healthy":true}`). Admin API lands on `localhost:8090`.
3. **Start the website** (from `ably/website/apps/website`): ensure `.env` has `REALTIME_ENV=local` (default) and add `JWT_SECRET` (a working value is in `.env.sample`; it is currently missing from the live `.env` and token auth fails without it). `docker-compose up -d`, then `bin/dev` for the first (seeding) boot and `bin/dev --quick` thereafter. Note the Rails port it prints.
4. **Mint a local Control API token**: `bin/rake "control_api:generate_tokens[1,1,/tmp/ably_tokens.json]"` writes `{account_id, tokens}`. The token is a self-contained HS256 JWT signed with `JWT_SECRET`, no external service involved. This task also provisions an account, so the farm must be up.
5. **Point the provider/tests at it**: `ABLY_URL=http://localhost:<rails-port>/api/v1`, `ABLY_ACCOUNT_TOKEN=<minted token>`, then `TF_ACC=1 make testacc` (or `go test`).
6. **Region caveat**: the queue resource validator only accepts `us-east-1-a` and `eu-west-1-a`, and the e2e tests use `us-east-1-a`; the local stack must accept that region or queue creation fails.

Blockers that stop this being a cold-start-unattended path today, and why Tier 1 is the agent's real loop:

- The farm pulls images from a private ECR registry whose login expires every 12 hours and needs AWS creds for the Ably prod account.
- The farm builds private Go modules over SSH and needs an SSH key with `ably-labs` org access.
- First farm start compiles Go roles and routinely exceeds the default 30s health timeout (hence `--daemon-timeout=60`).
- The website needs `mise`/`asdf` toolchains, the `ably-env` CLI, assorted `.env` secrets, and the private `cartography` gem (skippable).
- Only one farm can run at a time.

None of these are things an agent can clear on its own, which is why Tier 1 is the loop we hand the agent and Tier 3 is a warm environment a human sets up.

### Keeping the fake honest

The risk with Tier 1 is a fake that drifts from reality and passes anyway. Two mitigations, in order of preference: derive the fake's response shapes from real recorded staging responses (record once, hand-curate into the fake), and add a periodic contract check that diffs the fake's responses against the Tier 2 staging suite. The fake proves the provider is internally consistent; only staging proves it matches production. The two tiers together are the credible loop.

### What to build

- An `httptest`-backed fake Control API in `internal/provider/`, plus a `make test` target that runs the hermetic suite with zero env. This is the agent's inner loop and the highest-value item from this whole exercise.
- A documented, scripted Tier 3 bring-up (a `make local-stack` that sequences farm-then-website with the Redis-port and ordering caveats baked in), accepting it still needs the human-gated prerequisites cleared once.
- A contract check tying the fake to staging so it cannot quietly lie.

### Status: Tier 1 is built (2026-06-09)

The hermetic fake exists at `internal/provider/fake_control_api_test.go` and `make test` now runs the entire provider acceptance suite against it with no credentials and no network, green. `make testacc` is unchanged and still hits a real Control API when `TF_ACC` is set. The fake is ~500 lines of stateful in-memory CRUD over the endpoints in `control/*.go`.

Two things the build surfaced that are worth carrying forward:

- **The `dev_overrides` trap is real and it bit us immediately.** The repo pins the provider source to the `ably/ably` namespace, which the in-process reattach factory (keyed by the bare type `ably`) cannot satisfy, so the suite relies on `dev_overrides`. A stale `dev_overrides` in `~/.terraformrc` silently ran an old installed binary instead of the code under test, so edits appeared to do nothing. `TestMain` now defends against this: it builds the provider from current source into a temp dir and writes its own clean `dev_overrides` config, guaranteeing the tests exercise the current code. This is the same class of silent-staleness failure the strategy warns about, and an agent would have been badly misled by it.
- **The fake's honesty ledger has started.** Two real-API behaviours had to be encoded so the provider's computed attributes did not drift: namespaces always return `batchingEnabled`/`conflationEnabled` (default false), and HTTP-family rule targets default `format` to `json`. These are exactly the "keep the fake honest" items; each is a small, documented divergence the staging suite (Tier 2) should eventually be diffed against.

### Open questions

- Is it worth investing to make Tier 3 cold-startable in CI (caching ECR creds, vendoring the private modules), or is Tier 1 plus the existing staging Tier 2 enough? My instinct is the latter: the staging suite already gives real-behaviour proof in CI, so Tier 3's value is local debugging, not the automated loop.
- Could we issue a long-lived staging account token for local use, so a developer or agent gets a real round-trip by pointing at staging with no local stack at all? That may beat Tier 3 for most cases.
- Who owns the fake's fidelity, and how often does the contract check run?

## 8. Reference: load-bearing files

- `swagger/v1/swagger.yaml` (website repo): `oneOf`/discriminator ~1331-1481, pkcs12 multipart ~266-314
- `control/client.go`: retryablehttp transport, `retryPolicy`, `WithHTTPClient` ~44-132
- `control/types.go`: `RuleResponse` 479-491, `NamespacePost`/`NamespacePatch` 139-172
- `control/rules.go`: `CreateRule`/`UpdateRule` taking `any` 22, 38
- `internal/provider/rules.go`: `GetPlanRule`/`GetRuleResponse` switch cases, plus `modifiers.go` and `models.go`
