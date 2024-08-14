# Local Development

terraform-provider-ably/internal/provider includes provider specific code

Update your ~/.terraformrc file with dev_overrides
E.G.
```
plugin_cache_dir   = "$HOME/.terraform.d/plugin-cache"
disable_checkpoint = true

provider_installation {

  dev_overrides {
      "github.com/ably/ably" = "/Users/grahamrussell/Documents/src/bin",
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

Build your changes from the repository root with
```
$ go install
$ cd examples/playground && terraform plan
```

NOTE: ensure GOBIN env var is set to the path configured in your dev_overrides section of ~/.terraformrc

Generate docs for this provider by installing tfplugindocs (https://github.com/hashicorp/terraform-plugin-docs) and running tfplugindocs from the repository root.
