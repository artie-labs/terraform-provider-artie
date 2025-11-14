# Import a PrivateLink connection by using its UUID, which you can find by:
# 1. Go to the PrivateLink connections page in the Artie UI
# 2. Click on the PrivateLink connection you want to import
# 3. Open the drop down in the top right corner and select "View UUID"
terraform import artie_private_link.example <privatelink_uuid>

# Then print the state and copy it into your terraform config file (be sure to remove all read-only/computed fields like `uuid`, `status`, and `dns_entry`):
terraform state show artie_private_link.example
