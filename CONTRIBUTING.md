# Contributing to Ably Terraform provider

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Ensure you have added suitable tests and the test suite is passing
5. Ensure that `tfplugindocs` has been run and that the Ably Terraform provider documentation is up to date. Terraform Plugin Docs tool is available [HERE](https://github.com/hashicorp/terraform-plugin-docs)
6. Push the branch (`git push origin my-new-feature`)
7. Create a new Pull Request

## Release Process

1. Merge all pull requests containing changes intended for this release to `main` branch
2. Prepare a [Release Branch](#release-branch) and a corresponding pull request, obtain approval from reviewers and then merge to `main` branch
3. Push Git [Version Tag](#version-tag)
4. Trigger the Release Workflow to create a draft Github Release
5. Review the draft GitHub release and publish it if everything is ok
6. Publishing a Github release will send a webhook to Terraform Registry, which will in turn ingest the new release

N.B. Releasing and publishing Terraform provider follows a process that is different from the [general release guidance for Ably SDKs](https://github.com/ably/engineering/blob/main/sdk/releases.md) due to the requirements of Terraform Registry.

### Release Branch

Should:

- branch from the `main` branch
- merge to the `main` branch, once approved
- be named like `release/<version>`
- increment the version, conforming to [SemVer](https://semver.org/)
- add a change log entry (process to be documented under [#17](https://github.com/ably/engineering/issues/17))

### Version Tag

Should:

- have a `v` prefix - e.g. `v1.2.3` for the release of version `1.2.3`
- not be subsequently moved
