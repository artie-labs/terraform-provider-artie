package tfmodels

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type Source struct {
	Type           types.String     `tfsdk:"type"`
	Tables         map[string]Table `tfsdk:"tables"`
	MySQLConfig    *MySQLConfig     `tfsdk:"mysql_config"`
	PostgresConfig *PostgresConfig  `tfsdk:"postgresql_config"`
}

func (s Source) ToAPIModel() artieclient.Source {
	var sourceConfig artieclient.SourceConfig
	sourceType := artieclient.SourceTypeFromString(s.Type.ValueString())
	switch sourceType {
	case artieclient.MySQL:
		sourceConfig = s.MySQLConfig.ToAPIModel()
	case artieclient.PostgreSQL:
		sourceConfig = s.PostgresConfig.ToAPIModel()
	default:
		panic(fmt.Sprintf("invalid source type: %s", s.Type.ValueString()))
	}

	tables := []artieclient.Table{}
	for _, table := range s.Tables {
		tables = append(tables, table.ToAPIModel())
	}

	return artieclient.Source{
		Type:   sourceType,
		Config: sourceConfig,
		Tables: tables,
	}
}

func SourceFromAPIModel(apiModel artieclient.Source) Source {
	source := Source{
		Type:   types.StringValue(string(apiModel.Type)),
		Tables: TablesFromAPIModel(apiModel.Tables),
	}

	switch apiModel.Type {
	case artieclient.MySQL:
		source.MySQLConfig = MySQLConfigFromAPIModel(apiModel.Config)
	case artieclient.PostgreSQL:
		source.PostgresConfig = PostgresConfigFromAPIModel(apiModel.Config)
	default:
		panic(fmt.Sprintf("invalid source type: %s", apiModel.Type))
	}

	return source
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
