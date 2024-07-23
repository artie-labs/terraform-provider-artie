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

data "artie_deployments" "example" {}

output "deployments" {
  value = data.artie_deployments.example.deployments
}
