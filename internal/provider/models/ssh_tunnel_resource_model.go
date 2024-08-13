package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type SSHTunnelResourceModel struct {
	UUID      types.String `tfsdk:"uuid"`
	Name      types.String `tfsdk:"name"`
	Host      types.String `tfsdk:"host"`
	Port      types.Int32  `tfsdk:"port"`
	Username  types.String `tfsdk:"username"`
	PublicKey types.String `tfsdk:"public_key"`
}
