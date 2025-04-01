package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type SourceReader struct {
	UUID                            types.String `tfsdk:"uuid"`
	Name                            types.String `tfsdk:"name"`
	DataPlaneName                   types.String `tfsdk:"data_plane_name"`
	ConnectorUUID                   types.String `tfsdk:"connector_uuid"`
	DatabaseName                    types.String `tfsdk:"database_name"`
	OracleContainerName             types.String `tfsdk:"oracle_container"`
	OneTopicPerSchema               types.Bool   `tfsdk:"one_topic_per_schema"`
	PostgresPublicationNameOverride types.String `tfsdk:"postgres_publication_name_override"`
	PostgresReplicationSlotOverride types.String `tfsdk:"postgres_replication_slot_override"`
}

func (s SourceReader) ToAPIBaseModel() (artieclient.BaseSourceReader, diag.Diagnostics) {
	connectorUUID, diags := parseUUID(s.ConnectorUUID)
	if diags.HasError() {
		return artieclient.BaseSourceReader{}, diags
	}

	return artieclient.BaseSourceReader{
		Name:          s.Name.ValueString(),
		DataPlaneName: s.DataPlaneName.ValueString(),
		ConnectorUUID: connectorUUID,
		DatabaseName:  s.DatabaseName.ValueString(),
		ContainerName: s.OracleContainerName.ValueString(),
		Settings: artieclient.SourceReaderSettings{
			OneTopicPerSchema:               s.OneTopicPerSchema.ValueBool(),
			PostgresPublicationNameOverride: s.PostgresPublicationNameOverride.ValueString(),
			PostgresReplicationSlotOverride: s.PostgresReplicationSlotOverride.ValueString(),
		},
	}, nil
}

func (s SourceReader) ToAPIModel() (artieclient.SourceReader, diag.Diagnostics) {
	uuid, diags := parseUUID(s.UUID)
	if diags.HasError() {
		return artieclient.SourceReader{}, diags
	}

	baseSourceReader, diags := s.ToAPIBaseModel()
	if diags.HasError() {
		return artieclient.SourceReader{}, diags
	}

	return artieclient.SourceReader{
		UUID:             uuid,
		BaseSourceReader: baseSourceReader,
	}, nil
}

func SourceReaderFromAPIModel(apiModel artieclient.SourceReader) SourceReader {
	return SourceReader{
		UUID:                            types.StringValue(apiModel.UUID.String()),
		Name:                            types.StringValue(apiModel.Name),
		DataPlaneName:                   types.StringValue(apiModel.DataPlaneName),
		ConnectorUUID:                   types.StringValue(apiModel.ConnectorUUID.String()),
		DatabaseName:                    types.StringValue(apiModel.DatabaseName),
		OracleContainerName:             types.StringValue(apiModel.ContainerName),
		OneTopicPerSchema:               types.BoolValue(apiModel.Settings.OneTopicPerSchema),
		PostgresPublicationNameOverride: types.StringValue(apiModel.Settings.PostgresPublicationNameOverride),
		PostgresReplicationSlotOverride: types.StringValue(apiModel.Settings.PostgresReplicationSlotOverride),
	}
}
