---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "artie Provider"
subcategory: ""
description: |-
  
---

# artie Provider



## Example Usage

```terraform
terraform {
  required_providers {
    artie = {
      source = "artie-labs/artie"
    }
  }
}

variable "artie_api_key" {
  type      = string
  sensitive = true
}

provider "artie" {
  api_key = var.artie_api_key
}

variable "snowflake_password" {
  type      = string
  sensitive = true
}

variable "postgres_password" {
  type      = string
  sensitive = true
}

resource "artie_destination" "snowflake" {
  type  = "snowflake"
  label = "Snowflake (Analytics)"
  snowflake_config = {
    account_url = "https://abc12345.snowflakecomputing.com"
    virtual_dwh = "compute_wh"
    username    = "user_abcd"
    password    = var.snowflake_password
  }
}

resource "artie_ssh_tunnel" "ssh_tunnel" {
  name     = "SSH Tunnel"
  host     = "1.2.3.4"
  port     = 22
  username = "artie"
}

resource "artie_deployment" "postgres_to_snowflake" {
  name = "PostgreSQL to Snowflake"
  source = {
    type = "postgresql"
    postgresql_config = {
      host     = "server.example.com"
      port     = 5432
      database = "customers"
      user     = "artie"
      password = var.postgres_password
    }
    tables = {
      "public.account" = {
        name                = "account"
        schema              = "public"
        enable_history_mode = true
      },
      "public.company" = {
        name   = "company"
        schema = "public"
      }
    }
  }
  destination_uuid = artie_destination.snowflake.uuid
  ssh_tunnel_uuid  = artie_ssh_tunnel.ssh_tunnel.uuid
  destination_config = {
    database = "ANALYTICS"
    schema   = "PUBLIC"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `api_key` (String, Sensitive) Artie API key to authenticate requests to the Artie API. Generate an API key in the Artie web app at https://app.artie.com/settings. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file.

### Optional

- `endpoint` (String) Artie API endpoint. This defaults to https://api.artie.com and should not need to be changed except when developing the provider.
