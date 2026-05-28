# A column hashing salt (Artie generates the salt value automatically if omitted):
resource "artie_column_hashing_salt" "my_salt" {
  name        = "My Hashing Salt"
  description = "Used for hashing PII columns in the analytics pipeline"
}

# Reference the salt in a pipeline:
resource "artie_pipeline" "postgres_to_snowflake" {
  name                     = "PostgreSQL to Snowflake"
  source_reader_uuid       = artie_source_reader.postgres.uuid
  column_hashing_salt_uuid = artie_column_hashing_salt.my_salt.uuid
  tables = {
    "public.users" = {
      name            = "users"
      schema          = "public"
      columns_to_hash = ["email", "phone_number"]
    }
  }
  destination_connector_uuid = artie_connector.snowflake.uuid
  destination_config = {
    database = "ANALYTICS"
    schema   = "PUBLIC"
  }
}
