package tfmodels

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type Destination struct {
	UUID            types.String           `tfsdk:"uuid"`
	SSHTunnelUUID   types.String           `tfsdk:"ssh_tunnel_uuid"`
	Type            types.String           `tfsdk:"type"`
	Label           types.String           `tfsdk:"label"`
	SnowflakeConfig *SnowflakeSharedConfig `tfsdk:"snowflake_config"`
	BigQueryConfig  *BigQuerySharedConfig  `tfsdk:"bigquery_config"`
	RedshiftConfig  *RedshiftSharedConfig  `tfsdk:"redshift_config"`
}

type SnowflakeSharedConfig struct {
	AccountURL types.String `tfsdk:"account_url"`
	VirtualDWH types.String `tfsdk:"virtual_dwh"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	PrivateKey types.String `tfsdk:"private_key"`
}

type BigQuerySharedConfig struct {
	ProjectID       types.String `tfsdk:"project_id"`
	Location        types.String `tfsdk:"location"`
	CredentialsData types.String `tfsdk:"credentials_data"`
}

type RedshiftSharedConfig struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (d *Destination) UpdateFromAPIModel(apiModel artieclient.Destination) {
	d.UUID = types.StringValue(apiModel.UUID.String())
	d.Type = types.StringValue(string(apiModel.Type))
	d.Label = types.StringValue(apiModel.Label)
	d.SSHTunnelUUID = optionalUUIDToStringValue(apiModel.SSHTunnelUUID)
	d.SnowflakeConfig = nil
	d.BigQueryConfig = nil
	d.RedshiftConfig = nil

	switch apiModel.Type {
	case artieclient.Snowflake:
		d.SnowflakeConfig = &SnowflakeSharedConfig{
			AccountURL: types.StringValue(apiModel.Config.SnowflakeAccountURL),
			VirtualDWH: types.StringValue(apiModel.Config.SnowflakeVirtualDWH),
			PrivateKey: types.StringValue(apiModel.Config.SnowflakePrivateKey),
			Username:   types.StringValue(apiModel.Config.Username),
			Password:   types.StringValue(apiModel.Config.Password),
		}
	case artieclient.BigQuery:
		d.BigQueryConfig = &BigQuerySharedConfig{
			ProjectID:       types.StringValue(apiModel.Config.GCPProjectID),
			Location:        types.StringValue(apiModel.Config.GCPLocation),
			CredentialsData: types.StringValue(apiModel.Config.GCPCredentialsData),
		}
	case artieclient.Redshift:
		d.RedshiftConfig = &RedshiftSharedConfig{
			Endpoint: types.StringValue(apiModel.Config.Endpoint),
			Username: types.StringValue(apiModel.Config.Username),
			Password: types.StringValue(apiModel.Config.Password),
		}
	default:
		panic(fmt.Sprintf("invalid destination type: %s", apiModel.Type))
	}
}

func (d Destination) ToAPIBaseModel() artieclient.BaseDestination {
	var sharedConfig artieclient.DestinationSharedConfig
	destinationType := artieclient.DestinationTypeFromString(d.Type.ValueString())
	switch destinationType {
	case artieclient.Snowflake:
		sharedConfig = artieclient.DestinationSharedConfig{
			SnowflakeAccountURL: d.SnowflakeConfig.AccountURL.ValueString(),
			SnowflakeVirtualDWH: d.SnowflakeConfig.VirtualDWH.ValueString(),
			SnowflakePrivateKey: d.SnowflakeConfig.PrivateKey.ValueString(),
			Username:            d.SnowflakeConfig.Username.ValueString(),
			Password:            d.SnowflakeConfig.Password.ValueString(),
		}
	case artieclient.BigQuery:
		sharedConfig = artieclient.DestinationSharedConfig{
			GCPProjectID:       d.BigQueryConfig.ProjectID.ValueString(),
			GCPLocation:        d.BigQueryConfig.Location.ValueString(),
			GCPCredentialsData: d.BigQueryConfig.CredentialsData.ValueString(),
		}
	case artieclient.Redshift:
		sharedConfig = artieclient.DestinationSharedConfig{
			Endpoint: d.RedshiftConfig.Endpoint.ValueString(),
			Username: d.RedshiftConfig.Username.ValueString(),
			Password: d.RedshiftConfig.Password.ValueString(),
		}
	default:
		panic(fmt.Sprintf("invalid destination type: %s", d.Type.ValueString()))
	}

	return artieclient.BaseDestination{
		Type:          destinationType,
		Label:         d.Label.ValueString(),
		Config:        sharedConfig,
		SSHTunnelUUID: ParseOptionalUUID(d.SSHTunnelUUID),
	}
}

func (d Destination) ToAPIModel() artieclient.Destination {
	return artieclient.Destination{
		UUID:            parseUUID(d.UUID),
		BaseDestination: d.ToAPIBaseModel(),
	}
}
