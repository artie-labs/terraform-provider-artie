package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type DestinationResourceModel struct {
	UUID          types.String                  `tfsdk:"uuid"`
	CompanyUUID   types.String                  `tfsdk:"company_uuid"`
	SSHTunnelUUID types.String                  `tfsdk:"ssh_tunnel_uuid"`
	Name          types.String                  `tfsdk:"name"`
	Label         types.String                  `tfsdk:"label"`
	LastUpdatedAt types.String                  `tfsdk:"last_updated_at"`
	Config        *DestinationSharedConfigModel `tfsdk:"config"`
}

type DestinationSharedConfigModel struct {
	Host                types.String `tfsdk:"host"`
	Port                types.Int64  `tfsdk:"port"`
	Endpoint            types.String `tfsdk:"endpoint"`
	Username            types.String `tfsdk:"username"`
	GCPProjectID        types.String `tfsdk:"gcp_project_id"`
	GCPLocation         types.String `tfsdk:"gcp_location"`
	AWSAccessKeyID      types.String `tfsdk:"aws_access_key_id"`
	AWSRegion           types.String `tfsdk:"aws_region"`
	SnowflakeAccountURL types.String `tfsdk:"snowflake_account_url"`
	SnowflakeVirtualDWH types.String `tfsdk:"snowflake_virtual_dwh"`
	// TODO sensitive fields
	// Password           types.String `tfsdk:"password"`
	// GCPCredentialsData types.String `tfsdk:"gcp_credentials_data"`
	// AWSSecretAccessKey types.String `tfsdk:"aws_secret_access_key"`
	// SnowflakePrivateKey types.String `tfsdk:"snowflake_private_key"`
}
