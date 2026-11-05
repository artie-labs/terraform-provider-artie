# Import a Managed Kafka admin user by using its cluster's UUID and its own UUID, separated by a slash.
# You can find the user's UUID with the List Kafka users API endpoint (GET /kafka-clusters/{uuid}/users).
terraform import artie_kafka_admin_user.example <kafka_cluster_uuid>/<kafka_user_uuid>
