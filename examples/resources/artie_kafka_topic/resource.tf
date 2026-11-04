resource "artie_kafka_topic" "example" {
  cluster_uuid       = artie_kafka_cluster.example.uuid
  topic              = "orders"
  partition_count    = 6
  replication_factor = 3
  retention_ms       = 604800000 # 7 days
  tiered_storage     = "on"

  config_entries = {
    "cleanup.policy" = "delete"
  }
}
