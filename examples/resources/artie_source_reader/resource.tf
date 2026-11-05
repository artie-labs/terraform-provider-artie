# A source reader that can only be used by one pipeline:
resource "artie_source_reader" "postgres_dev_reader" {
  name                               = "Postgres Dev Customers Reader"
  connector_uuid                     = artie_connector.postgres_dev.uuid
  database_name                      = "customers"
  postgres_replication_slot_override = "artie_reader"
  is_shared                          = false
}

# A source reader that can be used by multiple pipelines (must specify tables):
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

# A Microsoft SQL Server source reader that reads transaction log backups from an SMB share connector:
resource "artie_source_reader" "mssql_trn_reader" {
  name                     = "MSSQL Orders Reader"
  connector_uuid           = artie_connector.mssql_prod.uuid
  database_name            = "orders"
  mssql_replication_method = "trn_log_backup"
  archive_connector_uuid   = artie_connector.fsx_log_backups.uuid
}
