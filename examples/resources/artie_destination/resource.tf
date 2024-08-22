variable "snowflake_password" {
  type      = string
  sensitive = true
}

resource "artie_destination" "snowflake" {
  type  = "snowflake"
  label = "Snowflake (Analytics)"
  snowflake_config = {
    account_url = "https://abc12345.snowflakecomputing.com"
    virtual_dwh = "compute_wh"
    username    = "user_abcd"
    password    = var.snowflake_password
  }
}
