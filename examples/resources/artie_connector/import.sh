# Import a connector by using its UUID, which you can find by:
# 1. Go to the pipeline overview page in the Artie UI
# 2. Click on the "View UUIDs" button to see all related resource UUIDs
terraform import artie_connector.my_connector <connector_uuid>

# Then print the state and copy it into your terraform config file
# (be sure to remove all read-only fields, like `uuid`):
terraform state show artie_connector.my_connector 
