terraform {
  required_providers {
    artie = {
      source = "artie.com/terraform/artie"
    }
  }
}

variable "artie_api_key" {
  type      = string
  sensitive = true
}

provider "artie" {
  api_key = var.artie_api_key
}
