package models

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type DestinationType string

const (
	Snowflake DestinationType = "snowflake"
	BigQuery  DestinationType = "bigquery"
	Redshift  DestinationType = "redshift"
)

type DestinationResourceModel struct {
	UUID            types.String                `tfsdk:"uuid"`
	SSHTunnelUUID   types.String                `tfsdk:"ssh_tunnel_uuid"`
	Type            types.String                `tfsdk:"type"`
	Label           types.String                `tfsdk:"label"`
	SnowflakeConfig *SnowflakeSharedConfigModel `tfsdk:"snowflake_config"`
	BigQueryConfig  *BigQuerySharedConfigModel  `tfsdk:"bigquery_config"`
	RedshiftConfig  *RedshiftSharedConfigModel  `tfsdk:"redshift_config"`
}

type SnowflakeSharedConfigModel struct {
	AccountURL types.String `tfsdk:"account_url"`
	VirtualDWH types.String `tfsdk:"virtual_dwh"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	PrivateKey types.String `tfsdk:"private_key"`
}

type BigQuerySharedConfigModel struct {
	ProjectID       types.String `tfsdk:"project_id"`
	Location        types.String `tfsdk:"location"`
	CredentialsData types.String `tfsdk:"credentials_data"`
}

type RedshiftSharedConfigModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (d *DestinationResourceModel) UpdateFromAPIModel(apiModel artieclient.Destination) {
	d.UUID = types.StringValue(apiModel.UUID.String())
	d.Type = types.StringValue(apiModel.Type)
	d.Label = types.StringValue(apiModel.Label)
	d.SSHTunnelUUID = optionalUUIDToStringValue(apiModel.SSHTunnelUUID)

	switch strings.ToLower(d.Type.ValueString()) {
	case string(Snowflake):
		d.SnowflakeConfig = &SnowflakeSharedConfigModel{
			AccountURL: types.StringValue(apiModel.Config.SnowflakeAccountURL),
			VirtualDWH: types.StringValue(apiModel.Config.SnowflakeVirtualDWH),
			PrivateKey: types.StringValue(apiModel.Config.SnowflakePrivateKey),
			Username:   types.StringValue(apiModel.Config.Username),
			Password:   types.StringValue(apiModel.Config.Password),
		}
	case string(BigQuery):
		d.BigQueryConfig = &BigQuerySharedConfigModel{
			ProjectID:       types.StringValue(apiModel.Config.GCPProjectID),
			Location:        types.StringValue(apiModel.Config.GCPLocation),
			CredentialsData: types.StringValue(apiModel.Config.GCPCredentialsData),
		}
	case string(Redshift):
		d.RedshiftConfig = &RedshiftSharedConfigModel{
			Endpoint: types.StringValue(apiModel.Config.Endpoint),
			Username: types.StringValue(apiModel.Config.Username),
			Password: types.StringValue(apiModel.Config.Password),
		}
	}
}

func (d DestinationResourceModel) ToAPIBaseModel() artieclient.BaseDestination {
	var sharedConfig artieclient.DestinationSharedConfig
	switch strings.ToLower(d.Type.ValueString()) {
	case string(Snowflake):
		sharedConfig = artieclient.DestinationSharedConfig{
			SnowflakeAccountURL: d.SnowflakeConfig.AccountURL.ValueString(),
			SnowflakeVirtualDWH: d.SnowflakeConfig.VirtualDWH.ValueString(),
			SnowflakePrivateKey: d.SnowflakeConfig.PrivateKey.ValueString(),
			Username:            d.SnowflakeConfig.Username.ValueString(),
			Password:            d.SnowflakeConfig.Password.ValueString(),
		}
	case string(BigQuery):
		sharedConfig = artieclient.DestinationSharedConfig{
			GCPProjectID:       d.BigQueryConfig.ProjectID.ValueString(),
			GCPLocation:        d.BigQueryConfig.Location.ValueString(),
			GCPCredentialsData: d.BigQueryConfig.CredentialsData.ValueString(),
		}
	case string(Redshift):
		sharedConfig = artieclient.DestinationSharedConfig{
			Endpoint: d.RedshiftConfig.Endpoint.ValueString(),
			Username: d.RedshiftConfig.Username.ValueString(),
			Password: d.RedshiftConfig.Password.ValueString(),
		}
	default:
		sharedConfig = artieclient.DestinationSharedConfig{}
	}

	return artieclient.BaseDestination{
		Type:          d.Type.ValueString(),
		Label:         d.Label.ValueString(),
		Config:        sharedConfig,
		SSHTunnelUUID: ParseOptionalUUID(d.SSHTunnelUUID),
	}
}

func (d DestinationResourceModel) ToAPIModel() artieclient.Destination {
	return artieclient.Destination{
		UUID:            parseUUID(d.UUID),
		BaseDestination: d.ToAPIBaseModel(),
	}
}
