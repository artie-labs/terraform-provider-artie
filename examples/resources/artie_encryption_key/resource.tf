# A passphrase-based encryption key (Artie generates the key material automatically):
resource "artie_encryption_key" "passphrase_key" {
  name        = "My Encryption Key"
  description = "Used for encrypting PII columns in the analytics pipeline"
}

# Reference the key in a pipeline:
resource "artie_pipeline" "postgres_to_snowflake" {
  name                = "PostgreSQL to Snowflake"
  source_reader_uuid  = artie_source_reader.postgres.uuid
  encryption_key_uuid = artie_encryption_key.passphrase_key.uuid
  tables = {
    "public.users" = {
      name               = "users"
      schema             = "public"
      columns_to_encrypt = ["ssn", "credit_card_number"]
    }
  }
  destination_connector_uuid = artie_connector.snowflake.uuid
  destination_config = {
    database = "ANALYTICS"
    schema   = "PUBLIC"
  }
}
