variable "postgres_password" {
  type      = string
  sensitive = true
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
