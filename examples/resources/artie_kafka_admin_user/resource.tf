resource "artie_kafka_admin_user" "example" {
  cluster_uuid = artie_kafka_cluster.example.uuid
  username     = "ops-admin"
}

output "kafka_admin_password" {
  value     = artie_kafka_admin_user.example.password
  sensitive = true
}
