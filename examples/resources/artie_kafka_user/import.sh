# Import a Managed Kafka user by using its cluster's UUID and its own UUID, separated by a slash. You can find them by:
# 1. Go to the Kafka clusters page in the Artie UI (https://app.artie.com) and click on the cluster
# 2. Copy the cluster's UUID, then copy the user's UUID from the cluster's users
terraform import artie_kafka_user.example <kafka_cluster_uuid>/<kafka_user_uuid>
