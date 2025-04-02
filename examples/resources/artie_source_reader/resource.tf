resource "artie_source_reader" "postgres_dev_reader" {
  name                               = "Postgres Dev Customers Reader"
  connector_uuid                     = artie_connector.postgres_dev.uuid
  database_name                      = "customers"
  postgres_replication_slot_override = "artie_reader"
}
