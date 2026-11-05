resource "artie_kafka_user" "consumer" {
  cluster_uuid = artie_kafka_cluster.example.uuid
  username     = "my-app"
}

resource "artie_kafka_user" "admin" {
  cluster_uuid   = artie_kafka_cluster.example.uuid
  username       = "ops-admin"
  access_profile = "admin"
}

output "consumer_password" {
  value     = artie_kafka_user.consumer.password
  sensitive = true
}
