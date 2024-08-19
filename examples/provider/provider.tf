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
  destination_config = {
    database = "ANALYTICS"
    schema   = "PUBLIC"
  }
}
