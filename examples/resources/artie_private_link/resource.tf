resource "artie_private_link" "example" {
  name            = "My PrivateLink Connection"
  aws_account_id  = "123456789012"
  aws_region      = "us-east-1"
  vpc_endpoint_id = "vpce-1234567890abcdef0"
}

