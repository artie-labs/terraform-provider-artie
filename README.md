# Artie Terraform Provider

Currently in development. TODO: [publish it on the Terraform Registry](https://developer.hashicorp.com/terraform/registry/providers/publishing)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Developing the Provider

Create a `~/.terraformrc` file containing the following:

```
provider_installation {

  dev_overrides {
      "artie.com/terraform/artie" = "/Users/<your-username>/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

**You'll need to run `go install` any time you make changes to the provider code before running it again locally.**

This provider requires an Artie API key. To run it against your local Artie API instance, create a new API key through the dashboard (locally) and then set the following env vars:

```shell
export ARTIE_API_KEY=<yoursecretkey>
export ARTIE_ENDPOINT=https://0.0.0.0:8000/api
```

The `examples/` directory contains example Terraform config files. To test managing an Artie deployment with this provider:

```shell
cd examples/deployments
terraform plan
terraform apply
```

You'll be prompted for any secrets that are specified in the config. You can avoid having to enter them each time by setting them as env vars, e.g.:

```shell
export TF_VAR_snowflake_password=...
export TF_VAR_mongodb_password=...
```

To run with a particular log level:
```shell
TF_LOG=INFO terraform plan
```

### Documentation

To generate or update documentation, run `go generate`. If you make changes to a resource's schema and don't run this, a CI check will fail until you run it and commit the result.

### Testing

TODO: add acceptance tests before publishing.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.
