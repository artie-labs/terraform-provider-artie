resource "artie_pipeline" "postgres_to_snowflake" {
  name               = "PostgreSQL to Snowflake"
  source_reader_uuid = artie_source_reader.postgres.uuid
  destination_uuid   = artie_connector.snowflake.uuid
  destination_config = {
    database = "ANALYTICS"
    schema   = "PUBLIC"
  }
  soft_delete_rows                = true
  include_artie_updated_at_column = true
}
