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
