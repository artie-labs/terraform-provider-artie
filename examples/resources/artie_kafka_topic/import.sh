# Import a Kafka topic by using its cluster's UUID and its name, separated by a slash. You can find the cluster's UUID by:
# 1. Go to the Kafka clusters page in the Artie UI (https://app.artie.com)
# 2. Click on the Kafka cluster the topic is on and copy the UUID
terraform import artie_kafka_topic.example <kafka_cluster_uuid>/<topic>

# Then print the state and copy it into your terraform config file. `config_entries` isn't read back from the API, so add
# the ones you want Terraform to track; they aren't applied to the existing topic.
terraform state show artie_kafka_topic.example
