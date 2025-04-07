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

variable "postgres_password" {
  type      = string
  sensitive = true
}

variable "snowflake_password" {
  type      = string
  sensitive = true
}

resource "artie_ssh_tunnel" "ssh_tunnel" {
  name     = "SSH Tunnel"
  host     = "1.2.3.4"
  port     = 22
  username = "artie"
}

resource "artie_connector" "postgres_dev" {
  name = "Postgres Dev"
  type = "postgresql"
  postgresql_config = {
    host     = "server.example.com"
    port     = 5432
    user     = "artie"
    password = var.postgres_password
  }
  ssh_tunnel_uuid = artie_ssh_tunnel.ssh_tunnel.uuid
}

resource "artie_connector" "snowflake" {
  name = "Snowflake (Analytics)"
  type = "snowflake"
  snowflake_config = {
    account_url = "https://abc12345.snowflakecomputing.com"
    virtual_dwh = "compute_wh"
    username    = "user_abcd"
    password    = var.snowflake_password
  }
}

resource "artie_source_reader" "postgres_dev_reader" {
  name                               = "Postgres Dev Customers Reader"
  connector_uuid                     = artie_connector.postgres_dev.uuid
  database_name                      = "customers"
  postgres_replication_slot_override = "artie_reader"
}


resource "artie_pipeline" "postgres_to_snowflake" {
  name               = "PostgreSQL to Snowflake"
  source_reader_uuid = artie_source_reader.postgres_dev_reader.uuid
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
  destination_connector_uuid = artie_connector.snowflake.uuid
  destination_config = {
    database = "ANALYTICS"
    schema   = "PUBLIC"
  }
  soft_delete_rows                = true
  include_artie_updated_at_column = true
}
