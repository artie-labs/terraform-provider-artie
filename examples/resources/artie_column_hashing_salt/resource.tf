# Salts belong to the whole company, so manage them with a key that has access to All environments
# (full access, or the column_hashing_salts scopes).
provider "artie" {
  alias   = "workspace"
  api_key = var.artie_workspace_api_key
}

resource "artie_column_hashing_salt" "pii" {
  provider    = artie.workspace
  name        = "PII"
  description = "Hashes customer emails and phone numbers"
}

# Pipelines can stay on the default provider, even with an environment-bound key.
resource "artie_pipeline" "postgres_to_snowflake" {
  name               = "PostgreSQL to Snowflake"
  source_reader_uuid = artie_source_reader.postgres.uuid
  tables = {
    "public.customer" = {
      name            = "customer"
      schema          = "public"
      columns_to_hash = ["email", "phone"]
    }
  }
  column_hashing_salt_uuid   = artie_column_hashing_salt.pii.uuid
  destination_connector_uuid = artie_connector.snowflake.uuid
  destination_config = {
    database = "ANALYTICS"
    schema   = "PUBLIC"
  }
}
