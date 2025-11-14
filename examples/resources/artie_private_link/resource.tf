resource "artie_private_link" "example" {
  vpc_service_name = "com.amazonaws.vpce.us-east-1.vpce-svc-1234567890abcdef0"
  region           = "us-east-1"
  vpc_endpoint_id  = "vpce-1234567890abcdef0"
}

