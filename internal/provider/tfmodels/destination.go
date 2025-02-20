package tfmodels

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type Destination struct {
	UUID            types.String           `tfsdk:"uuid"`
	SSHTunnelUUID   types.String           `tfsdk:"ssh_tunnel_uuid"`
	Type            types.String           `tfsdk:"type"`
	Label           types.String           `tfsdk:"label"`
	BigQueryConfig  *BigQuerySharedConfig  `tfsdk:"bigquery_config"`
	RedshiftConfig  *RedshiftSharedConfig  `tfsdk:"redshift_config"`
	S3Config        *S3SharedConfig        `tfsdk:"s3_config"`
	SnowflakeConfig *SnowflakeSharedConfig `tfsdk:"snowflake_config"`
}

func (d Destination) ToAPIBaseModel() (artieclient.BaseDestination, diag.Diagnostics) {
	var sharedConfig artieclient.DestinationSharedConfig
	destinationType, err := artieclient.DestinationTypeFromString(d.Type.ValueString())
	if err != nil {
		return artieclient.BaseDestination{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert Destination to API model", err.Error(),
		)}
	}

	switch destinationType {
	case artieclient.BigQuery:
		sharedConfig = d.BigQueryConfig.ToAPIModel()
	case artieclient.Redshift:
		sharedConfig = d.RedshiftConfig.ToAPIModel()
	case artieclient.S3:
		sharedConfig = d.S3Config.ToAPIModel()
	case artieclient.Snowflake:
		sharedConfig = d.SnowflakeConfig.ToAPIModel()
	default:
		return artieclient.BaseDestination{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert Destination to API model", fmt.Sprintf("unhandled destination type: %s", d.Type.ValueString()),
		)}
	}

	sshTunnelUUID, diags := parseOptionalUUID(d.SSHTunnelUUID)
	if diags.HasError() {
		return artieclient.BaseDestination{}, diags
	}

	return artieclient.BaseDestination{
		Type:          destinationType,
		Label:         d.Label.ValueString(),
		Config:        sharedConfig,
		SSHTunnelUUID: sshTunnelUUID,
	}, diags
}

func (d Destination) ToAPIModel() (artieclient.Destination, diag.Diagnostics) {
	baseModel, diags := d.ToAPIBaseModel()
	if diags.HasError() {
		return artieclient.Destination{}, diags
	}

	uuid, uuidDiags := parseUUID(d.UUID)
	diags.Append(uuidDiags...)
	if diags.HasError() {
		return artieclient.Destination{}, diags
	}

	return artieclient.Destination{
		UUID:            uuid,
		BaseDestination: baseModel,
	}, diags
}

func DestinationFromAPIModel(apiModel artieclient.Destination) (Destination, diag.Diagnostics) {
	destination := Destination{
		UUID:          types.StringValue(apiModel.UUID.String()),
		Type:          types.StringValue(string(apiModel.Type)),
		Label:         types.StringValue(apiModel.Label),
		SSHTunnelUUID: optionalUUIDToStringValue(apiModel.SSHTunnelUUID),
	}

	switch apiModel.Type {
	case artieclient.BigQuery:
		destination.BigQueryConfig = BigQuerySharedConfigFromAPIModel(apiModel.Config)
	case artieclient.Redshift:
		destination.RedshiftConfig = RedshiftSharedConfigFromAPIModel(apiModel.Config)
	case artieclient.S3:
		destination.S3Config = S3SharedConfigFromAPIModel(apiModel.Config)
	case artieclient.Snowflake:
		destination.SnowflakeConfig = SnowflakeSharedConfigFromAPIModel(apiModel.Config)
	default:
		return Destination{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert API model to Destination", fmt.Sprintf("invalid destination type: %s", apiModel.Type),
		)}
	}

	return destination, nil
}

type BigQuerySharedConfig struct {
	ProjectID       types.String `tfsdk:"project_id"`
	Location        types.String `tfsdk:"location"`
	CredentialsData types.String `tfsdk:"credentials_data"`
}

func (b BigQuerySharedConfig) ToAPIModel() artieclient.DestinationSharedConfig {
	return artieclient.DestinationSharedConfig{
		GCPProjectID:       b.ProjectID.ValueString(),
		GCPLocation:        b.Location.ValueString(),
		GCPCredentialsData: b.CredentialsData.ValueString(),
	}
}

func BigQuerySharedConfigFromAPIModel(apiModel artieclient.DestinationSharedConfig) *BigQuerySharedConfig {
	return &BigQuerySharedConfig{
		ProjectID:       types.StringValue(apiModel.GCPProjectID),
		Location:        types.StringValue(apiModel.GCPLocation),
		CredentialsData: types.StringValue(apiModel.GCPCredentialsData),
	}
}

type RedshiftSharedConfig struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (r RedshiftSharedConfig) ToAPIModel() artieclient.DestinationSharedConfig {
	return artieclient.DestinationSharedConfig{
		Endpoint: r.Endpoint.ValueString(),
		Username: r.Username.ValueString(),
		Password: r.Password.ValueString(),
	}
}

func RedshiftSharedConfigFromAPIModel(apiModel artieclient.DestinationSharedConfig) *RedshiftSharedConfig {
	return &RedshiftSharedConfig{
		Endpoint: types.StringValue(apiModel.Endpoint),
		Username: types.StringValue(apiModel.Username),
		Password: types.StringValue(apiModel.Password),
	}
}

type S3SharedConfig struct {
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	Region          types.String `tfsdk:"region"`
}

func (s S3SharedConfig) ToAPIModel() artieclient.DestinationSharedConfig {
	return artieclient.DestinationSharedConfig{
		AWSAccessKeyID:     s.AccessKeyID.ValueString(),
		AWSSecretAccessKey: s.SecretAccessKey.ValueString(),
		AWSRegion:          s.Region.ValueString(),
	}
}

func S3SharedConfigFromAPIModel(apiModel artieclient.DestinationSharedConfig) *S3SharedConfig {
	return &S3SharedConfig{
		AccessKeyID:     types.StringValue(apiModel.AWSAccessKeyID),
		SecretAccessKey: types.StringValue(apiModel.AWSSecretAccessKey),
		Region:          types.StringValue(apiModel.AWSRegion),
	}
}

type SnowflakeSharedConfig struct {
	AccountURL types.String `tfsdk:"account_url"`
	VirtualDWH types.String `tfsdk:"virtual_dwh"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	PrivateKey types.String `tfsdk:"private_key"`
}

func (s SnowflakeSharedConfig) ToAPIModel() artieclient.DestinationSharedConfig {
	return artieclient.DestinationSharedConfig{
		SnowflakeAccountURL: s.AccountURL.ValueString(),
		SnowflakeVirtualDWH: s.VirtualDWH.ValueString(),
		SnowflakePrivateKey: s.PrivateKey.ValueString(),
		Username:            s.Username.ValueString(),
		Password:            s.Password.ValueString(),
	}
}

func SnowflakeSharedConfigFromAPIModel(apiModel artieclient.DestinationSharedConfig) *SnowflakeSharedConfig {
	return &SnowflakeSharedConfig{
		AccountURL: types.StringValue(apiModel.SnowflakeAccountURL),
		VirtualDWH: types.StringValue(apiModel.SnowflakeVirtualDWH),
		PrivateKey: types.StringValue(apiModel.SnowflakePrivateKey),
		Username:   types.StringValue(apiModel.Username),
		Password:   types.StringValue(apiModel.Password),
	}
}
