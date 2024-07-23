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
	name = "new name"
}

# data "artie_deployments" "example" {}

# output "deployments" {
#   value = data.artie_deployments.example.deployments
# }
