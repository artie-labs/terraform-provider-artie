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

This provider requires an Artie API key. To run it against your local Artie API instance, create a new API key through the dashboard (locally) and then set this env var:

```shell
export TF_VAR_artie_api_key=<yoursecretkey>
```

Create an `example.tf` file in the top level directory (it will be git-ignored) to hold the Terraform config you want to develop against. Ping Dana for an example of what to put in it.

To test managing an Artie deployment with this provider:

```shell
go install
terraform plan
terraform apply
```

You'll be prompted for any secrets that are specified in the config. You can avoid having to enter them each time by setting them as env vars, e.g.:

```shell
export TF_VAR_snowflake_password=...
export TF_VAR_postgres_password=...
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

_Note:_ Acceptance tests create real resources, and often cost money to run.
