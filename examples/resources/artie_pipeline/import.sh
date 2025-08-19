# Import a pipeline by using its UUID, which you can find by:
# 1. Go to the pipeline overview page in the Artie UI
# 2. Open the dropdown in the top right corner
# 3. Select "View UUIDs" to see all related resource UUIDs
terraform import artie_pipeline.my_pipeline <pipeline_uuid>

# Then print the state and copy it into your terraform config file
# (be sure to remove all read-only fields, like `uuid`):
terraform state show artie_pipeline.my_pipeline
