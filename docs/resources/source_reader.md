---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "artie_source_reader Resource - terraform-provider-artie"
subcategory: ""
description: |-
  Artie Source Reader resource. This represents a process that reads data from a source connector and inserts it info Kafka. A Source Reader can be used by multiple Pipelines, e.g. to read from a single PostgreSQL replication slot and copy the data to multiple destinations.
---

# artie_source_reader (Resource)

Artie Source Reader resource. This represents a process that reads data from a source connector and inserts it info Kafka. A Source Reader can be used by multiple Pipelines, e.g. to read from a single PostgreSQL replication slot and copy the data to multiple destinations.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `connector_uuid` (String) The source connector that we should read data from.

### Optional

- `backfill_batch_size` (Number) The number of rows to read from the source database in each batch while backfilling. Maximum allowed value is 50,000. Default is 5,000.
- `data_plane_name` (String) The name of the data plane to deploy this source reader in. If this is not set, we will use the default data plane for your account. To see the full list of supported data planes on your account, click on 'New pipeline' in our UI.
- `database_name` (String) The name of the database we should read data from in the source connector. This should be specified if the source connector's type is DocumentDB, MongoDB, MySQL, MS SQL, Oracle (this maps to the service name), or PostgreSQL.
- `databases_to_unify` (List of String) If `enable_unify_across_databases` is set to true, this should be a list of databases within your Microsoft SQL Server that we should sync data from. All tables that you opt into being unified should exist in each of these databases. This is only applicable if the source type is Microsoft SQL Server.
- `enable_heartbeats` (Boolean) If the source database is a very low-traffic PostgreSQL database (e.g., a dev database) and is running on Amazon RDS, we recommend setting this to true to prevent WAL growth issues. This is only applicable if the source type is PostgreSQL.
- `enable_unify_across_databases` (Boolean) If set to true, you can specify multiple databases within your Microsoft SQL Server that we should sync data from, and we will unify tables with the same name and schema into a single destination table. This is useful if you have multiple identical databases and want to fan-in the data. This is only applicable if the source type is Microsoft SQL Server and `mssql_replication_method` is set to `fn_dblog`.
- `enable_unify_across_schemas` (Boolean) If set to true, you can specify tables that should be generalized to all schemas, meaning we will sync all tables with the same name into the same destination table. This is useful if you have multiple identical schemas and want to fan-in the data. This is only applicable if the source type is PostgreSQL.
- `is_shared` (Boolean) If set to true, this source reader can be used by multiple pipelines.
- `mssql_replication_method` (String) If unset, we will use the default replication method (Capture Instances). If set to `fn_dblog`, we will stream data from transaction logs via SQL access. This is only applicable if the source type is Microsoft SQL Server.
- `name` (String) An optional human-readable label for this source reader.
- `one_topic_per_schema` (Boolean) If set to true, Artie will write all incoming CDC events into a single Kafka topic per schema. This is currently only supported if your source is Oracle and your account has this feature enabled.
- `oracle_container_name` (String) The name of the container (pluggable database) if the source type is Oracle and you are using a container database.
- `partition_suffix_regex_pattern` (String) If this source reader is reading any partitioned tables, this regex pattern should describe the expected suffix of each partition's name so that we can consume data from all partitions. If not set, this defaults to `_((default)|([0-9]{4})_(0[1-9]|1[012]))$` - meaning that for a table called `my_table` that's partitioned by month, we will detect partitions such as `my_table_default`, `my_table_2025_01`, `my_table_2025_02`, etc.
- `postgres_publication_mode` (String) This should be set to `filtered` if the PostgreSQL publication in the source database is not set to include `ALL TABLES`. If that's the case, you will need to explicitly add tables to the publication.
- `postgres_publication_name_override` (String) If set, this will override the name of the PostgreSQL publication. Otherwise, we will use our default value, `dbz_publication`. This is only applicable if the source type is PostgreSQL.
- `postgres_replication_slot_override` (String) If set, this will override the name of the PostgreSQL replication slot. Otherwise, we will use our default value, `artie`. This is only applicable if the source type is PostgreSQL.
- `tables` (Attributes Map) A map of tables from the source database that you want this source reader to include CDC events for. This should be specified if (and only if) the source reader has `is_shared` set to true, and it must include all tables that are specified in the `tables` attribute of any pipeline that uses this source reader. The key for each table should be formatted as `schema_name.table_name` if your source database uses schemas, otherwise just `table_name`. (see [below for nested schema](#nestedatt--tables))

### Read-Only

- `uuid` (String)

<a id="nestedatt--tables"></a>
### Nested Schema for `tables`

Required:

- `name` (String) The name of the table in the source database.

Optional:

- `child_partition_schema_name` (String) If the source table is partitioned and its child partitions are in a different schema, this should specify the name of that schema.
- `columns_to_exclude` (List of String) An optional list of columns to exclude from CDC events. This cannot be used if `columns_to_include` is also specified.
- `columns_to_include` (List of String) An optional list of columns to include in CDC events. If not provided, all columns will be included. This cannot be used if `columns_to_exclude` is also specified.
- `is_partitioned` (Boolean) If the source table is partitioned, set this to true and we will ingest data from all of its partitions. You may also need to customize `partition_suffix_regex_pattern` on the source reader.
- `schema` (String) The name of the schema the table belongs to in the source database. This must be specified if your source database uses schemas (such as PostgreSQL), e.g. `public`.
- `unify_across_databases` (Boolean) This should be set to true for any tables that you intend to unify across databases in any pipeline that uses this source reader. This is only applicable if the source reader has `enable_unify_across_databases` set to true and `databases_to_unify` filled.
- `unify_across_schemas` (Boolean) This should be set to true for any tables that you intend to unify across schemas in any pipeline that uses this source reader. This is only applicable if the source reader has `enable_unify_across_schemas` set to true.

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
# Import a source reader by using its UUID, which you can find by:
# 1. Go to the pipeline overview page in the Artie UI
# 2. Click on the "View UUIDs" button to see all related resource UUIDs
terraform import artie_source_reader.my_source_reader <source_reader_uuid>

# Then print the state and copy it into your terraform config file
# (be sure to remove all read-only fields, like `uuid`):
terraform state show artie_source_reader.my_source_reader
```
