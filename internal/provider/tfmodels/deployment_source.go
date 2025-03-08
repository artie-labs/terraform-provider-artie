package tfmodels

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type Source struct {
	Type           types.String     `tfsdk:"type"`
	Tables         map[string]Table `tfsdk:"tables"`
	MySQLConfig    *MySQLConfig     `tfsdk:"mysql_config"`
	MSSQLConfig    *MSSQLConfig     `tfsdk:"mssql_config"`
	OracleConfig   *OracleConfig    `tfsdk:"oracle_config"`
	PostgresConfig *PostgresConfig  `tfsdk:"postgresql_config"`
}

func (s Source) ToAPIModel(ctx context.Context) (artieclient.Source, diag.Diagnostics) {
	var sourceConfig artieclient.SourceConfig
	sourceType, err := artieclient.ConnectorTypeFromString(s.Type.ValueString())
	if err != nil {
		return artieclient.Source{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert Source to API model", err.Error(),
		)}
	}

	switch sourceType {
	case artieclient.MySQL:
		sourceConfig = s.MySQLConfig.ToAPIModel()
	case artieclient.MSSQL:
		sourceConfig = s.MSSQLConfig.ToAPIModel()
	case artieclient.Oracle:
		sourceConfig = s.OracleConfig.ToAPIModel()
	case artieclient.PostgreSQL:
		sourceConfig = s.PostgresConfig.ToAPIModel()
	default:
		return artieclient.Source{}, []diag.Diagnostic{diag.NewErrorDiagnostic(
			"Unable to convert Source to API model", fmt.Sprintf("unhandled source type: %s", s.Type.ValueString()),
		)}
	}

	tables := []artieclient.Table{}
	diags := diag.Diagnostics{}
	for _, table := range s.Tables {
		apiTable, tableDiags := table.ToAPIModel(ctx)
		diags.Append(tableDiags...)
		if diags.HasError() {
			return artieclient.Source{}, diags
		}
		tables = append(tables, apiTable)
	}

	return artieclient.Source{
		Type:   sourceType,
		Config: sourceConfig,
		Tables: tables,
	}, diags
}

func SourceFromAPIModel(ctx context.Context, apiModel artieclient.Source) (Source, diag.Diagnostics) {
	tables, diags := TablesFromAPIModel(ctx, apiModel.Tables)
	if diags.HasError() {
		return Source{}, diags
	}

	source := Source{
		Type:   types.StringValue(string(apiModel.Type)),
		Tables: tables,
	}

	switch apiModel.Type {
	case artieclient.MySQL:
		source.MySQLConfig = MySQLConfigFromAPIModel(apiModel.Config)
	case artieclient.MSSQL:
		source.MSSQLConfig = MSSQLConfigFromAPIModel(apiModel.Config)
	case artieclient.Oracle:
		source.OracleConfig = OracleConfigFromAPIModel(apiModel.Config)
	case artieclient.PostgreSQL:
		source.PostgresConfig = PostgresConfigFromAPIModel(apiModel.Config)
	default:
		diags.AddError("Unable to convert API model to Source", fmt.Sprintf("invalid source type: %s", apiModel.Type))
		return Source{}, diags
	}

	return source, diags
}

type MySQLConfig struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int32  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Database types.String `tfsdk:"database"`
	Password types.String `tfsdk:"password"`
}

func (m MySQLConfig) ToAPIModel() artieclient.SourceConfig {
	return artieclient.SourceConfig{
		Host:     m.Host.ValueString(),
		Port:     m.Port.ValueInt32(),
		User:     m.User.ValueString(),
		Password: m.Password.ValueString(),
		Database: m.Database.ValueString(),
	}
}

func MySQLConfigFromAPIModel(apiModel artieclient.SourceConfig) *MySQLConfig {
	return &MySQLConfig{
		Host:     types.StringValue(apiModel.Host),
		Port:     types.Int32Value(apiModel.Port),
		User:     types.StringValue(apiModel.User),
		Password: types.StringValue(apiModel.Password),
		Database: types.StringValue(apiModel.Database),
	}
}

type MSSQLConfig struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int32  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Database types.String `tfsdk:"database"`
	Password types.String `tfsdk:"password"`
}

func (m MSSQLConfig) ToAPIModel() artieclient.SourceConfig {
	return artieclient.SourceConfig{
		Host:     m.Host.ValueString(),
		Port:     m.Port.ValueInt32(),
		User:     m.User.ValueString(),
		Password: m.Password.ValueString(),
		Database: m.Database.ValueString(),
	}
}

func MSSQLConfigFromAPIModel(apiModel artieclient.SourceConfig) *MSSQLConfig {
	return &MSSQLConfig{
		Host:     types.StringValue(apiModel.Host),
		Port:     types.Int32Value(apiModel.Port),
		User:     types.StringValue(apiModel.User),
		Password: types.StringValue(apiModel.Password),
		Database: types.StringValue(apiModel.Database),
	}
}

type OracleConfig struct {
	Host      types.String `tfsdk:"host"`
	Port      types.Int32  `tfsdk:"port"`
	User      types.String `tfsdk:"user"`
	Password  types.String `tfsdk:"password"`
	Service   types.String `tfsdk:"service"`
	Container types.String `tfsdk:"container"`
}

func (o OracleConfig) ToAPIModel() artieclient.SourceConfig {
	return artieclient.SourceConfig{
		Host:      o.Host.ValueString(),
		Port:      o.Port.ValueInt32(),
		User:      o.User.ValueString(),
		Password:  o.Password.ValueString(),
		Database:  o.Service.ValueString(),
		Container: o.Container.ValueString(),
	}
}

func OracleConfigFromAPIModel(apiModel artieclient.SourceConfig) *OracleConfig {
	return &OracleConfig{
		Host:      types.StringValue(apiModel.Host),
		Port:      types.Int32Value(apiModel.Port),
		User:      types.StringValue(apiModel.User),
		Password:  types.StringValue(apiModel.Password),
		Service:   types.StringValue(apiModel.Database),
		Container: types.StringValue(apiModel.Container),
	}
}

type PostgresConfig struct {
	Host     types.String `tfsdk:"host"`
	Port     types.Int32  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Database types.String `tfsdk:"database"`
	Password types.String `tfsdk:"password"`
}

func (p PostgresConfig) ToAPIModel() artieclient.SourceConfig {
	return artieclient.SourceConfig{
		Host:     p.Host.ValueString(),
		Port:     p.Port.ValueInt32(),
		User:     p.User.ValueString(),
		Password: p.Password.ValueString(),
		Database: p.Database.ValueString(),
	}
}

func PostgresConfigFromAPIModel(apiModel artieclient.SourceConfig) *PostgresConfig {
	return &PostgresConfig{
		Host:     types.StringValue(apiModel.Host),
		Port:     types.Int32Value(apiModel.Port),
		User:     types.StringValue(apiModel.User),
		Password: types.StringValue(apiModel.Password),
		Database: types.StringValue(apiModel.Database),
	}
}
