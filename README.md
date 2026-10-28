# Terraform Provider for FCO

Manage resources on the fi-ts.io platform using Terraform.

> [!IMPORTANT]
> This provider is in **alpha** stage. The schema, API surface, and behavior
> are still being shaped and may change without notice. Expect breaking
> changes and use at your own risk.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.27

## Building the provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

## Using the provider

First, let's add the provider to your project:

```terraform
terraform {
  required_providers {
    fco = {
      source = "fi-ts/fco"
    }
  }
}
```

The provider is configured with an API authentication token (a JWT), the base
URL of the API server, the project to operate in, and the tenant login. Any of
these may also be supplied via the `FCO_*` environment variables, in which case
values set in the configuration take precedence.

```terraform
provider "fco" {
  token        = "<YOUR_TOKEN>"        # or set env FCO_TOKEN
  base_url     = "https://api.fits.de" # or set env FCO_BASE_URL
  project_uuid = "<PROJECT_UUID>"      # or set env FCO_PROJECT_UUID
  tenant       = "<TENANT>"            # or set env FCO_TENANT
}
```

Now you are ready to go! :tada:

## Developing the provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
