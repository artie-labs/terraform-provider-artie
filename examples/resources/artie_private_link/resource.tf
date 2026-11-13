# AWS PrivateLink
resource "artie_private_link" "example" {
  name             = "my-privatelink"
  vpc_service_name = "com.amazonaws.vpce.us-east-1.vpce-svc-1234567890abcdef0"
  az_ids           = ["use1-az1", "use1-az2"]
}

# Azure Private Link (on an Azure data plane)
resource "artie_private_link" "azure" {
  name                       = "my-azure-privatelink"
  private_link_service_alias = "my-pls.00000000-0000-0000-0000-000000000000.eastus.azure.privatelinkservice"
  data_plane_name            = "my-azure-data-plane"
}
