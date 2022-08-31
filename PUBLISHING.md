# Ably Terraform Provider: Guidance on Releases and Publishing to Terraform Registry

Releasing and publishing Terraform provider follows a process that is different from the [general release guidance for Ably SDKs](https://github.com/ably/engineering/blob/main/sdk/releases.md), and this document outlines this process.

## Release Process

1. Merge all pull requests containing changes intended for this release to `main` branch
2. Prepare a [Release Branch](#release-branch) and a corresponding pull request, obtain approval from reviewers and then merge to `main` branch
3. Push Git [Version Tag](#version-tag)
4. Create GitHub release
5. New Github release will trigger the Release Workflow
6. It will also send a webhook to Terraform Registry, which will in turn ingest the new release


## Release Branch

Should:

- branch from the `main` branch
- merge to the `main` branch, once approved
- be named like `release/<version>`
- increment the version, conforming to [SemVer](https://semver.org/)
- add a change log entry (process to be documented under [#17](https://github.com/ably/engineering/issues/17))

## Version Tag

Should:

- have a `v` prefix - e.g. `v1.2.3` for the release of version `1.2.3`
- not be subsequently moved