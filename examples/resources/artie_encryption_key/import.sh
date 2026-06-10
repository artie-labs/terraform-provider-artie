# Import an encryption key by using its UUID, which you can find by:
# 1. Go to the pipeline overview page in the Artie UI (for a pipeline that uses this encryption key)
# 2. Open the dropdown in the top right corner
# 3. Select "View UUIDs" to see all related resource UUIDs
terraform import artie_encryption_key.my_encryption_key <encryption_key_uuid>

# Then print the state and copy it into your terraform config file
# (be sure to remove all read-only fields, like `uuid`, `key`, and `type`):
terraform state show artie_encryption_key.my_encryption_key
