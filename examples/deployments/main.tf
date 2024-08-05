terraform {
  required_providers {
    artie = {
      source = "artie.com/terraform/artie"
    }
  }
}

provider "artie" {
  endpoint = "http://0.0.0.0:8000"
}

import {
  to = artie_destination.snowflake
  id = "51b180a0-fbb9-49a2-ab45-cb46d913416d"
}

import {
  to = artie_deployment.dev_postgres_to_snowflake
  id = "c3dfa503-b6ae-48f3-a6b1-8491a506126d"
}

variable "snowflake_password" {
  type      = string
  sensitive = true
}
variable "postgres_password" {
  type      = string
  sensitive = true
}

resource "artie_destination" "snowflake" {
  type  = "Snowflake"
  label = "Snowflake (Partner Account)"
  snowflake_config = {
    account_url = "https://znb46775.snowflakecomputing.com"
    virtual_dwh = "compute_wh"
    username    = "tang8330"
    password    = var.snowflake_password
  }
}

resource "artie_deployment" "dev_postgres_to_snowflake" {
  name = "Dev PostgreSQL > Snowflake"
  source = {
    type = "PostgreSQL"
    postgres_config = {
      host     = "db-postgresql-sfo3-03243-do-user-13261354-0.c.db.ondigitalocean.com"
      port     = 25060
      database = "prod_dump_july_2024_4cvzb"
      user     = "doadmin"
      password = var.postgres_password
    }
    tables = [
      {
        name   = "invite"
        schema = "public"
      }
    ]
  }
  destination_uuid = artie_destination.snowflake.uuid
  destination_config = {
    database = "DEV_TEST"
    schema   = "PUBLIC"
  }
}

# import {
#   to = artie_destination.bigquery
#   id = "fa7d4efc-3957-41e5-b29c-66e2d49bffde"
# }

# variable "mongodb_password" {
#   type      = string
#   sensitive = true
# }

# variable "gcp_creds" {
#   type      = string
#   sensitive = true
# }

# resource "artie_destination" "bigquery" {
#   name  = "BigQuery"
#   label = "BigQuery"
#   config = {
#     gcp_location         = "us"
#     gcp_project_id       = "artie-labs"
#     gcp_credentials_data = var.gcp_creds
#   }
# }

# import {
#   to = artie_deployment.example
#   id = "38d5d2db-870a-4a38-a76c-9891b0e5122d"
# }

# resource "artie_deployment" "example" {
#   name = "MongoDB ➡️ BigQuery"
#   source = {
#     name = "MongoDB"
#     config = {
#       database = "myFirstDatabase"
#       host     = "mongodb+srv://cluster0.szddg49.mongodb.net/"
#       port     = 0
#       user     = "artie"
#       password = var.mongodb_password
#     }
#     tables = [
#       {
#         name   = "customers"
#         schema = ""
#       },
#       {
#         name   = "stock"
#         schema = ""
#       }
#     ]
#   }
#   destination_uuid = artie_destination.bigquery.uuid
#   destination_config = {
#     dataset = "customers"
#   }
# }
