variable "postgres_password" {
  type      = string
  sensitive = true
}

resource "artie_connector" "postgres_dev" {
  name = "Postgres Dev"
  type = "postgresql"
  postgresql_config = {
    host     = "server.example.com"
    port     = 5432
    user     = "artie"
    password = var.postgres_password
  }
}

variable "gcp_credentials" {
  type      = string
  sensitive = true
}

resource "artie_connector" "gcs_destination" {
  name = "GCS Destination"
  type = "gcs"
  gcs_config = {
    project_id       = "my-gcp-project"
    credentials_data = var.gcp_credentials
  }
}
