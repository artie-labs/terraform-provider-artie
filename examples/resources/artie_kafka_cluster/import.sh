# Import a Managed Kafka cluster by using its UUID, which you can find by:
# 1. Go to the Kafka clusters page in the Artie UI (https://app.artie.com)
# 2. Click on the Kafka cluster you want to import and copy the UUID
terraform import artie_kafka_cluster.example <kafka_cluster_uuid>

# Then print the state and copy it into your terraform config file (be sure to remove all read-only/computed fields like `uuid`, `status`, and `external_bootstrap_url`):
terraform state show artie_kafka_cluster.example
