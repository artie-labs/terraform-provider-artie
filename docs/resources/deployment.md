---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "artie_deployment Resource - terraform-provider-artie"
subcategory: ""
description: |-
  Artie Deployment resource. This represents a connection that syncs data from a single source (e.g., Postgres) to a single destination (e.g., Snowflake).
---

# artie_deployment (Resource)

Artie Deployment resource. This represents a connection that syncs data from a single source (e.g., Postgres) to a single destination (e.g., Snowflake).

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `destination_config` (Attributes) This contains configuration that pertains to the destination database but is specific to this deployment. The basic connection settings for the destination, which can be shared by multiple deployments, are stored in the corresponding `artie_destination` resource. (see [below for nested schema](#nestedatt--destination_config))
- `destination_uuid` (String) This must point to an `artie_destination` resource.
- `name` (String) The human-readable name of the deployment. This is used only as a label and can contain any characters.
- `source` (Attributes) This contains configuration for this deployment's source database. (see [below for nested schema](#nestedatt--source))

### Optional

- `drop_deleted_columns` (Boolean) If set to true, when a column is dropped from the source it will also be dropped in the destination.
- `include_artie_updated_at_column` (Boolean) If set to true, Artie will add a new column to your dataset called __artie_updated_at.
- `include_database_updated_at_column` (Boolean) If set to true, Artie will add a new column to your dataset called __artie_db_updated_at.
- `one_topic_per_schema` (Boolean) If set to true, Artie will write all incoming CDC events into a single Kafka topic per schema. This only works if your source is Oracle and your account has this feature enabled.
- `soft_delete_rows` (Boolean) If set to true, a new boolean column called __artie_delete will be added to your destination to indicate if the row has been deleted.
- `ssh_tunnel_uuid` (String) This can point to an `artie_ssh_tunnel` resource if you need us to use an SSH tunnel to connect to your source database.

### Read-Only

- `snowflake_eco_schedule_uuid` (String)
- `status` (String)
- `uuid` (String)

<a id="nestedatt--destination_config"></a>
### Nested Schema for `destination_config`

Optional:

- `bucket` (String) The name of the S3 bucket that data should be synced to. This should be filled if the destination is S3.
- `database` (String) The name of the database that data should be synced to in the destination. This should be filled if the destination is MS SQL or Snowflake, unless `use_same_schema_as_source` is set to true.
- `dataset` (String) The name of the dataset that data should be synced to in the destination. This should be filled if the destination is BigQuery.
- `folder` (String) If provided, all files will be stored under this folder inside the S3 bucket. This is optional and only applies if the destination is S3.
- `schema` (String) The name of the schema that data should be synced to in the destination. This should be filled if the destination is MS SQL, Redshift, or Snowflake (unless `use_same_schema_as_source` is set to true).
- `schema_name_prefix` (String) If `use_same_schema_as_source` is enabled, this prefix will be added to each schema name in the destination. This is useful if you want to namespace all of this deployment's schemas in the destination.
- `use_same_schema_as_source` (Boolean) If set to true, each table from the source database will be synced to a schema with the same name as its source schema. This can only be used if both the source and destination support multiple schemas (e.g. PostgreSQL, Redshift, Snowflake, etc).


<a id="nestedatt--source"></a>
### Nested Schema for `source`

Required:

- `tables` (Attributes Map) A map of tables from the source database that you want to replicate to the destination. The key for each table should be formatted as `schema_name.table_name` if your source database uses schemas, otherwise just `table_name`. (see [below for nested schema](#nestedatt--source--tables))
- `type` (String) The type of source database. This must be one of the following: `mysql`, `mssql`, `oracle`, `postgresql`.

Optional:

- `mssql_config` (Attributes) This should be filled out if the source type is `mssql`. (see [below for nested schema](#nestedatt--source--mssql_config))
- `mysql_config` (Attributes) This should be filled out if the source type is `mysql`. (see [below for nested schema](#nestedatt--source--mysql_config))
- `oracle_config` (Attributes) This should be filled out if the source type is `oracle`. (see [below for nested schema](#nestedatt--source--oracle_config))
- `postgresql_config` (Attributes) This should be filled out if the source type is `postgresql`. (see [below for nested schema](#nestedatt--source--postgresql_config))

<a id="nestedatt--source--tables"></a>
### Nested Schema for `source.tables`

Required:

- `name` (String) The name of the table in the source database.

Optional:

- `alias` (String) An optional alias for the table. If set, this will be the name of the destination table.
- `columns_to_exclude` (List of String) An optional list of columns to exclude from syncing to the destination.
- `columns_to_hash` (List of String) An optional list of columns to hash in the destination. Values for these columns will be obscured with a one-way hash.
- `enable_history_mode` (Boolean) If set to true, we will create an additional table in the destination (suffixed with `__history`) to store all changes to the source table over time.
- `individual_deployment` (Boolean) If set to true, we will spin up a separate Artie Transfer deployment to handle this table. This should only be used if this table has extremely high throughput (over 1M+ per hour) and has much higher throughput than other tables.
- `merge_predicates` (Attributes List) Optional: if the destination table is partitioned, specify the column(s) it's partitioned by. This will help with merge performance and currently only applies to Snowflake and BigQuery. For BigQuery, only one column can be specified and it must be a time column partitioned by day. (see [below for nested schema](#nestedatt--source--tables--merge_predicates))
- `schema` (String) The name of the schema the table belongs to in the source database. This must be specified if your source database uses schemas (such as PostgreSQL), e.g. `public`.
- `skip_deletes` (Boolean) If set to true, we will skip delete events for this table and only process insert and update events.

Read-Only:

- `is_partitioned` (Boolean)
- `uuid` (String)

<a id="nestedatt--source--tables--merge_predicates"></a>
### Nested Schema for `source.tables.merge_predicates`

Required:

- `partition_field` (String) The name of the column the destination table is partitioned by.



<a id="nestedatt--source--mssql_config"></a>
### Nested Schema for `source.mssql_config`

Required:

- `database` (String) The name of the database in Microsoft SQL Server.
- `host` (String) The hostname of the Microsoft SQL Server. This must point to the primary host, not a read replica.
- `password` (String, Sensitive) The password of the service account. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file.
- `port` (Number) The default port for Microsoft SQL Server is 1433.
- `user` (String) The username of the service account we will use to connect to the database.


<a id="nestedatt--source--mysql_config"></a>
### Nested Schema for `source.mysql_config`

Required:

- `database` (String) The name of the database in the MySQL server.
- `host` (String) The hostname of the MySQL database. This must point to the primary host, not a read replica.
- `password` (String, Sensitive) The password of the service account. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file.
- `port` (Number) The default port for MySQL is 3306.
- `user` (String) The username of the service account we will use to connect to the MySQL database. This service account needs enough permissions to read from the server binlogs.


<a id="nestedatt--source--oracle_config"></a>
### Nested Schema for `source.oracle_config`

Required:

- `host` (String) The hostname of the Oracle database. This must point to the primary host, not a read replica. This database must also have `ARCHIVELOG` mode and supplemental logging enabled.
- `password` (String, Sensitive) The password of the service account. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file.
- `port` (Number) The default port for Oracle is 1521.
- `service` (String) The name of the service in the Oracle server.
- `user` (String) The username of the service account we will use to connect to the Oracle database.

Optional:

- `container` (String) The name of the container (pluggable database). Required if you are using a container database; otherwise this should be omitted.


<a id="nestedatt--source--postgresql_config"></a>
### Nested Schema for `source.postgresql_config`

Required:

- `database` (String) The name of the database in the PostgreSQL server.
- `host` (String) The hostname of the PostgreSQL database. This must point to the primary host, not a read replica. This database must also have its `WAL_LEVEL` set to `logical`.
- `password` (String, Sensitive) The password of the service account. We recommend storing this in a secret manager and referencing it via a *sensitive* Terraform variable, instead of putting it in plaintext in your Terraform config file.
- `port` (Number) The default port for PostgreSQL is 5432.
- `user` (String) The username of the service account we will use to connect to the PostgreSQL database. This service account needs enough permissions to create and read from the replication slot.

## Import

Import is supported using the following syntax:

```shell
# Import an Artie deployment by using the uuid.
# https://app.artie.com/deployments/123e4567-e89b-12d3-a456-426614174000/overview

terraform import artie_deployment.default 123e4567-e89b-12d3-a456-426614174000
```
