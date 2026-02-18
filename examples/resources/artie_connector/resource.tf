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
    username = "artie"
    password = var.postgres_password
  }
}

variable "cockroach_password" {
  type      = string
  sensitive = true
}

resource "artie_connector" "cockroach_source" {
  name = "CockroachDB Source"
  type = "cockroach"
  cockroach_config = {
    host     = "my-cockroach-cluster.example.com"
    port     = 26257
    username = "artie"
    password = var.cockroach_password
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

variable "aws_access_key_id" {
  type = string
}

variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}

resource "artie_connector" "iceberg_s3tables" {
  name = "Iceberg S3 Tables"
  type = "iceberg"
  iceberg_config = {
    provider          = "s3tables"
    access_key_id     = var.aws_access_key_id
    secret_access_key = var.aws_secret_access_key
    bucket_arn        = "arn:aws:s3tables:us-east-1:123456789012:bucket/my-iceberg-bucket"
  }
}

variable "iceberg_rest_token" {
  type      = string
  sensitive = true
}

resource "artie_connector" "iceberg_rest_catalog" {
  name = "Iceberg REST Catalog"
  type = "iceberg"
  iceberg_config = {
    provider  = "rest"
    uri       = "https://workspace.cloud.databricks.com/api/2.1/unity-catalog/iceberg"
    token     = var.iceberg_rest_token
    warehouse = "my-warehouse"
  }
}

variable "iceberg_rest_client_id" {
  type = string
}

variable "iceberg_rest_client_secret" {
  type      = string
  sensitive = true
}

resource "artie_connector" "iceberg_rest_catalog_oauth2" {
  name = "Iceberg REST Catalog (OAuth2)"
  type = "iceberg"
  iceberg_config = {
    provider   = "rest"
    uri        = "https://workspace.cloud.databricks.com/api/2.1/unity-catalog/iceberg"
    credential = "${var.iceberg_rest_client_id}:${var.iceberg_rest_client_secret}"
    auth_uri   = "https://workspace.cloud.databricks.com/oidc/v1/token"
    warehouse  = "my-warehouse"
    scope      = "catalog"
  }
}
