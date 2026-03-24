# Contributing to Ably Terraform provider

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Ensure you have added suitable tests and the test suite is passing
5. Ensure that `tfplugindocs` has been run and that the Ably Terraform provider documentation is up to date. Terraform Plugin Docs tool is available [HERE](https://github.com/hashicorp/terraform-plugin-docs)
6. Push the branch (`git push origin my-new-feature`)
7. Create a new Pull Request

## Testing

See [DEVELOPMENT.md](DEVELOPMENT.md) for guidance on testing locally.

## Release Process

This repo contains two independently versioned components:

- **Provider** — the Terraform provider itself
- **Control** — the Go client for the Ably Control API (under `control/`)

Both are released from the same workflow but use different tag conventions and produce different artifacts.

### Tag Conventions

| Component | Tag format | Example |
|-----------|------------|---------|
| Provider | `v<semver>` | `v1.2.3` |
| Control | `control/v<semver>` | `control/v1.0.0` |

Tags should conform to [SemVer](https://semver.org/) and must not be moved once pushed.

The `control/v<semver>` prefix is the Go module convention for subdirectory modules. The Go module proxy indexes new versions from this tag automatically.

### How to Release

1. Merge all pull requests for the release to `main`.
2. Create a release PR from a branch named `release/<date>` (e.g. `release/2025-04-10`). This PR should increment the version(s) and receive approval before merging to `main`.
3. Tag the release commit(s) on `main`. You can tag one or both components at once:
   ```bash
   git tag v1.2.3
   git tag control/v1.0.0
   git push origin v1.2.3 control/v1.0.0
   ```
4. Trigger the **Release** workflow (`workflow_dispatch`) from the Actions tab. No inputs are needed — the workflow discovers what to release from the tags.
5. Review the draft GitHub Release(s) and publish when ready.
6. Update the [Ably Changelog](https://changelog.ably.com/) (via [headwayapp](https://headwayapp.co/)) with these changes (you can just copy the notes you added to the CHANGELOG)

### What the Workflow Does

- **Detects** which tags are unreleased by checking for existing GitHub Releases.
- **Provider release**: runs GoReleaser to build signed binaries, checksums, and a manifest, then creates a draft GitHub Release with all artifacts. Publishing this release triggers the Terraform Registry webhook.
- **Control release**: generates release notes scoped to PRs that touch `control/` paths, then creates a draft GitHub Release (no binary artifacts).
- Both release jobs run **in parallel** when both components are tagged.
- The workflow is **idempotent** — running it again after releases exist is a safe no-op.

N.B. The provider release process differs from the [general release guidance for Ably SDKs](https://github.com/ably/engineering/blob/main/best-practices/releases.md) due to Terraform Registry requirements.
