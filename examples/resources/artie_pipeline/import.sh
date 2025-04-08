# Import a pipeline by using its UUID, which you can find in
# the URL of the deployment overview page in the Artie UI, e.g.:
# https://app.artie.com/deployments/<pipeline_uuid>/overview
terraform import artie_pipeline.my_pipeline <pipeline_uuid>

# Then print the state and copy it into your terraform config file
# (be sure to remove all read-only fields, like `uuid`, `status`, etc.):
terraform state show artie_pipeline.my_pipeline

# Then you can find the UUIDs of the objects your pipeline depends on and import those too:
terraform import artie_source_reader.my_source_reader <my_pipeline.source_reader_uuid>
terraform import artie_connector.my_source_connector <my_source_reader.connector_uuid>
terraform import artie_connector.my_destination_connector <my_pipeline.destination_connector_uuid>
