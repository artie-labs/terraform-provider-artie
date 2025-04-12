# A source reader that can only be used by one pipeline:
resource "artie_source_reader" "postgres_dev_reader" {
  name                               = "Postgres Dev Customers Reader"
  connector_uuid                     = artie_connector.postgres_dev.uuid
  database_name                      = "customers"
  postgres_replication_slot_override = "artie_reader"
  is_shared                          = false
}

# A source reader that can be used by multiple pipelines:
resource "artie_source_reader" "postgres_dev_reader" {
  name                               = "Postgres Dev Customers Reader"
  connector_uuid                     = artie_connector.postgres_dev.uuid
  database_name                      = "customers"
  postgres_replication_slot_override = "artie_reader"
  is_shared                          = true
  tables = {
    "public.account" = {
      name               = "account"
      schema             = "public"
      columns_to_exclude = ["email"]
    },
    "public.company" = {
      name   = "company"
      schema = "public"
    }
  }
}
