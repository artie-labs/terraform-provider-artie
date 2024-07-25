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
        name              = "new_table"
        schema            = ""
        advanced_settings = {}
      },
      {
        name              = "stock"
        schema            = ""
        advanced_settings = {}
      }
    ]
  }
  destination_uuid = "fa7d4efc-3957-41e5-b29c-66e2d49bffde"
  destination_config = {
    dataset = "customers"
  }
  advanced_settings = {
    enable_soft_delete     = true
    flush_interval_seconds = 60
  }
}

# data "artie_deployments" "example" {}

# output "deployments" {
#   value = data.artie_deployments.example.deployments
# }
