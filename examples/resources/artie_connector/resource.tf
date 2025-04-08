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
