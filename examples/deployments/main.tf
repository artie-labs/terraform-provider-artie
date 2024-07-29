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
  to = artie_destination.bigquery
  id = "fa7d4efc-3957-41e5-b29c-66e2d49bffde"
}

resource "artie_destination" "bigquery" {
  name  = "BigQuery"
  label = "BigQuery"
  config = {
    gcp_location   = "us"
    gcp_project_id = "artie-labs"
  }
}

import {
  to = artie_deployment.example
  id = "38d5d2db-870a-4a38-a76c-9891b0e5122d"
}

resource "artie_deployment" "example" {
  name = "MongoDB ➡️ BigQuery"
  source = {
    name = "MongoDB"
    config = {
      database = "myFirstDatabase"
      host     = "mongodb+srv://cluster0.szddg49.mongodb.net/"
      port     = 0
      user     = "artie"
      dynamodb = {}
    }
    tables = [
      {
        name   = "customers"
        schema = ""
        advanced_settings = {
          skip_delete = false
        }
      },
      {
        name              = "stock"
        schema            = ""
        advanced_settings = {}
      }
    ]
  }
  destination_uuid = artie_destination.bigquery.uuid
  destination_config = {
    dataset = "customers"
  }
  advanced_settings = {
    enable_soft_delete     = true
    flush_interval_seconds = 60
  }
}

# resource "artie_destination" "bigquery2" {
#   name  = "BigQuery"
#   label = "BigQuery2 (to delete)"
#   config = {
#     gcp_location   = "us"
#     gcp_project_id = "artie-labs"
#   }
# }

# resource "artie_deployment" "newdeployment" {
#   name = "New deployment from tf"
#   source = {
#     name = "MongoDB"
#     config = {
#       database = "myFirstDatabase"
#       host     = "mongodb+srv://cluster0.szddg49.mongodb.net/"
#       port     = 0
#       user     = "artie"
#       dynamodb = {}
#     }
#     tables = [
#       {
#         name   = "customers"
#         schema = ""
#         advanced_settings = {
#           skip_delete = false
#         }
#       },
#     ]
#   }
#   destination_uuid = artie_destination.bigquery2.uuid
#   destination_config = {
#     dataset = "customers"
#   }
#   advanced_settings = {}
# }
