resource "artie_kafka_cluster" "example" {
  name          = "my-kafka-cluster"
  region        = "us-east-1"
  instance_type = "m8g.xlarge"
  brokers       = 3
  storage_gib   = 100
  allowed_cidrs = ["10.0.0.0/16"]

  tiered_storage = {
    local_retention_ms = 604800000 # 7 days
  }

  storage_autoscale = {
    max_gib = 500
  }
}
