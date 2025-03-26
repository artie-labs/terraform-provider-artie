package tfmodels

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type Source struct {
	Type           types.String    `tfsdk:"type"`
	Tables         types.Map       `tfsdk:"tables"`
	DynamoDBConfig *DynamoDBConfig `tfsdk:"dynamodb_config"`
	MongoDBConfig  *MongoDBConfig  `tfsdk:"mongodb_config"`
	MySQLConfig    *MySQLConfig    `tfsdk:"mysql_config"`
	MSSQLConfig    *MSSQLConfig    `tfsdk:"mssql_config"`
	OracleConfig   *OracleConfig   `tfsdk:"oracle_config"`
	PostgresConfig *PostgresConfig `tfsdk:"postgresql_config"`
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
	case artieclient.DynamoDB:
		sourceConfig = s.DynamoDBConfig.ToAPIModel()
	case artieclient.MongoDB:
		sourceConfig = s.MongoDBConfig.ToAPIModel()
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

	tables := map[string]Table{}
	diags := s.Tables.ElementsAs(ctx, &tables, false)
	apiTables := []artieclient.Table{}
	for _, table := range tables {
		apiTable, tableDiags := table.ToAPIModel(ctx)
		diags.Append(tableDiags...)
		if diags.HasError() {
			return artieclient.Source{}, diags
		}
		apiTables = append(apiTables, apiTable)
	}

	return artieclient.Source{
		Type:   sourceType,
		Config: sourceConfig,
		Tables: apiTables,
	}, diags
}

func SourceFromAPIModel(ctx context.Context, apiModel artieclient.Source) (Source, diag.Diagnostics) {
	tables, diags := TablesFromAPIModel(ctx, apiModel.Tables)
	if diags.HasError() {
		return Source{}, diags
	}

	tablesMap, mapDiags := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: TableAttrTypes}, tables)
	diags.Append(mapDiags...)

	source := Source{
		Type:   types.StringValue(string(apiModel.Type)),
		Tables: tablesMap,
	}

	switch apiModel.Type {
	case artieclient.DynamoDB:
		if apiModel.Config.DynamoDB == nil {
			diags.AddError("DynamoDB config is missing", "")
			return Source{}, diags
		}
		source.DynamoDBConfig = DynamoDBConfigFromAPIModel(*apiModel.Config.DynamoDB)
	case artieclient.MongoDB:
		source.MongoDBConfig = MongoDBConfigFromAPIModel(apiModel.Config)
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

type DynamoDBConfig struct {
	StreamArn          types.String `tfsdk:"stream_arn"`
	AwsAccessKeyID     types.String `tfsdk:"access_key_id"`
	AwsSecretAccessKey types.String `tfsdk:"secret_access_key"`
	Backfill           types.Bool   `tfsdk:"backfill"`
	BackfillBucket     types.String `tfsdk:"backfill_bucket"`
	BackfillFolder     types.String `tfsdk:"backfill_folder"`
}

func (d DynamoDBConfig) ToAPIModel() artieclient.SourceConfig {
	return artieclient.SourceConfig{
		DynamoDB: &artieclient.DynamoDBConfig{
			StreamsArn:         d.StreamArn.ValueString(),
			AwsAccessKeyID:     d.AwsAccessKeyID.ValueString(),
			AwsSecretAccessKey: d.AwsSecretAccessKey.ValueString(),
			SnapshotConfig: artieclient.DynamoDBSnapshotConfig{
				Enabled:        d.Backfill.ValueBool(),
				Bucket:         d.BackfillBucket.ValueString(),
				OptionalFolder: d.BackfillFolder.ValueString(),
			},
		},
	}
}

func DynamoDBConfigFromAPIModel(apiDynamoCfg artieclient.DynamoDBConfig) *DynamoDBConfig {
	return &DynamoDBConfig{
		StreamArn:          types.StringValue(apiDynamoCfg.StreamsArn),
		AwsAccessKeyID:     types.StringValue(apiDynamoCfg.AwsAccessKeyID),
		AwsSecretAccessKey: types.StringValue(apiDynamoCfg.AwsSecretAccessKey),
		Backfill:           types.BoolValue(apiDynamoCfg.SnapshotConfig.Enabled),
		BackfillBucket:     types.StringValue(apiDynamoCfg.SnapshotConfig.Bucket),
		BackfillFolder:     types.StringValue(apiDynamoCfg.SnapshotConfig.OptionalFolder),
	}
}

type MongoDBConfig struct {
	Host     types.String `tfsdk:"host"`
	User     types.String `tfsdk:"user"`
	Database types.String `tfsdk:"database"`
	Password types.String `tfsdk:"password"`
}

func (m MongoDBConfig) ToAPIModel() artieclient.SourceConfig {
	return artieclient.SourceConfig{
		Host:     m.Host.ValueString(),
		User:     m.User.ValueString(),
		Password: m.Password.ValueString(),
		Database: m.Database.ValueString(),
	}
}

func MongoDBConfigFromAPIModel(apiModel artieclient.SourceConfig) *MongoDBConfig {
	return &MongoDBConfig{
		Host:     types.StringValue(apiModel.Host),
		User:     types.StringValue(apiModel.User),
		Password: types.StringValue(apiModel.Password),
		Database: types.StringValue(apiModel.Database),
	}
}

type MySQLConfig struct {
	Host         types.String `tfsdk:"host"`
	SnapshotHost types.String `tfsdk:"snapshot_host"`
	Port         types.Int32  `tfsdk:"port"`
	User         types.String `tfsdk:"user"`
	Database     types.String `tfsdk:"database"`
	Password     types.String `tfsdk:"password"`
}

func (m MySQLConfig) ToAPIModel() artieclient.SourceConfig {
	return artieclient.SourceConfig{
		Host:         m.Host.ValueString(),
		SnapshotHost: m.SnapshotHost.ValueString(),
		Port:         m.Port.ValueInt32(),
		User:         m.User.ValueString(),
		Password:     m.Password.ValueString(),
		Database:     m.Database.ValueString(),
	}
}

func MySQLConfigFromAPIModel(apiModel artieclient.SourceConfig) *MySQLConfig {
	return &MySQLConfig{
		Host:         types.StringValue(apiModel.Host),
		SnapshotHost: types.StringValue(apiModel.SnapshotHost),
		Port:         types.Int32Value(apiModel.Port),
		User:         types.StringValue(apiModel.User),
		Password:     types.StringValue(apiModel.Password),
		Database:     types.StringValue(apiModel.Database),
	}
}

type MSSQLConfig struct {
	Host         types.String `tfsdk:"host"`
	SnapshotHost types.String `tfsdk:"snapshot_host"`
	Port         types.Int32  `tfsdk:"port"`
	User         types.String `tfsdk:"user"`
	Database     types.String `tfsdk:"database"`
	Password     types.String `tfsdk:"password"`
}

func (m MSSQLConfig) ToAPIModel() artieclient.SourceConfig {
	return artieclient.SourceConfig{
		Host:         m.Host.ValueString(),
		SnapshotHost: m.SnapshotHost.ValueString(),
		Port:         m.Port.ValueInt32(),
		User:         m.User.ValueString(),
		Password:     m.Password.ValueString(),
		Database:     m.Database.ValueString(),
	}
}

func MSSQLConfigFromAPIModel(apiModel artieclient.SourceConfig) *MSSQLConfig {
	return &MSSQLConfig{
		Host:         types.StringValue(apiModel.Host),
		SnapshotHost: types.StringValue(apiModel.SnapshotHost),
		Port:         types.Int32Value(apiModel.Port),
		User:         types.StringValue(apiModel.User),
		Password:     types.StringValue(apiModel.Password),
		Database:     types.StringValue(apiModel.Database),
	}
}

type OracleConfig struct {
	Host         types.String `tfsdk:"host"`
	SnapshotHost types.String `tfsdk:"snapshot_host"`
	Port         types.Int32  `tfsdk:"port"`
	User         types.String `tfsdk:"user"`
	Password     types.String `tfsdk:"password"`
	Service      types.String `tfsdk:"service"`
	Container    types.String `tfsdk:"container"`
}

func (o OracleConfig) ToAPIModel() artieclient.SourceConfig {
	return artieclient.SourceConfig{
		Host:         o.Host.ValueString(),
		SnapshotHost: o.SnapshotHost.ValueString(),
		Port:         o.Port.ValueInt32(),
		User:         o.User.ValueString(),
		Password:     o.Password.ValueString(),
		Database:     o.Service.ValueString(),
		Container:    o.Container.ValueString(),
	}
}

func OracleConfigFromAPIModel(apiModel artieclient.SourceConfig) *OracleConfig {
	return &OracleConfig{
		Host:         types.StringValue(apiModel.Host),
		SnapshotHost: types.StringValue(apiModel.SnapshotHost),
		Port:         types.Int32Value(apiModel.Port),
		User:         types.StringValue(apiModel.User),
		Password:     types.StringValue(apiModel.Password),
		Service:      types.StringValue(apiModel.Database),
		Container:    types.StringValue(apiModel.Container),
	}
}

type PostgresConfig struct {
	Host         types.String `tfsdk:"host"`
	SnapshotHost types.String `tfsdk:"snapshot_host"`
	Port         types.Int32  `tfsdk:"port"`
	User         types.String `tfsdk:"user"`
	Database     types.String `tfsdk:"database"`
	Password     types.String `tfsdk:"password"`
}

func (p PostgresConfig) ToAPIModel() artieclient.SourceConfig {
	return artieclient.SourceConfig{
		Host:         p.Host.ValueString(),
		SnapshotHost: p.SnapshotHost.ValueString(),
		Port:         p.Port.ValueInt32(),
		User:         p.User.ValueString(),
		Password:     p.Password.ValueString(),
		Database:     p.Database.ValueString(),
	}
}

func PostgresConfigFromAPIModel(apiModel artieclient.SourceConfig) *PostgresConfig {
	return &PostgresConfig{
		Host:         types.StringValue(apiModel.Host),
		SnapshotHost: types.StringValue(apiModel.SnapshotHost),
		Port:         types.Int32Value(apiModel.Port),
		User:         types.StringValue(apiModel.User),
		Password:     types.StringValue(apiModel.Password),
		Database:     types.StringValue(apiModel.Database),
	}
}
