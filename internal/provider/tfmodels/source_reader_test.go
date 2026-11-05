package tfmodels

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/openapi"
)

func TestSourceReaderFailoverReplicationSlotRoundTrip(t *testing.T) {
	model := SourceReader{
		UUID:                       types.StringValue(uuid.NewString()),
		ConnectorUUID:              types.StringValue(uuid.NewString()),
		UseFailoverReplicationSlot: types.BoolValue(true),
	}

	payload, diags := model.ToAPIModel(t.Context())
	assert.False(t, diags.HasError())
	assert.True(t, *payload.Settings.UseFailoverReplicationSlot)

	fromAPI, diags := SourceReaderFromAPIModel(t.Context(), openapi.PayloadsSourceReader{
		Uuid:          payload.Uuid,
		ConnectorUUID: payload.ConnectorUUID,
		Settings:      payload.Settings,
	})
	assert.False(t, diags.HasError())
	assert.True(t, fromAPI.UseFailoverReplicationSlot.ValueBool())
}

func TestSourceReaderToAPIModel_UnknownBackfillBatchSizeIsOmitted(t *testing.T) {
	// Left out of the config, backfill_batch_size is unknown at create. It must be omitted so the API applies its
	// default (25,000), not sent as 0, which the reader treats as its own fallback of 5,000.
	model := SourceReader{
		UUID:              types.StringValue(uuid.NewString()),
		ConnectorUUID:     types.StringValue(uuid.NewString()),
		BackfillBatchSize: types.Int64Unknown(),
	}

	payload, diags := model.ToAPIModel(t.Context())
	assert.False(t, diags.HasError())
	assert.Nil(t, payload.Settings.BackfillBatchSize)
}

func TestSourceReaderMongoDBDeploymentScopedMode(t *testing.T) {
	{
		// Set: sent to the API and read back.
		model := SourceReader{
			UUID:                        types.StringValue(uuid.NewString()),
			ConnectorUUID:               types.StringValue(uuid.NewString()),
			MongoDBDeploymentScopedMode: types.BoolValue(true),
		}

		payload, diags := model.ToAPIModel(t.Context())
		assert.False(t, diags.HasError())
		assert.True(t, *payload.Settings.DeploymentScopedMode)

		fromAPI, diags := SourceReaderFromAPIModel(t.Context(), openapi.PayloadsSourceReader{
			Uuid:          payload.Uuid,
			ConnectorUUID: payload.ConnectorUUID,
			Settings:      payload.Settings,
		})
		assert.False(t, diags.HasError())
		assert.True(t, fromAPI.MongoDBDeploymentScopedMode.ValueBool())
	}
	{
		// Left out of the config: omitted so the API keeps the stored value.
		model := SourceReader{
			UUID:                        types.StringValue(uuid.NewString()),
			ConnectorUUID:               types.StringValue(uuid.NewString()),
			MongoDBDeploymentScopedMode: types.BoolUnknown(),
		}

		payload, diags := model.ToAPIModel(t.Context())
		assert.False(t, diags.HasError())
		assert.Nil(t, payload.Settings.DeploymentScopedMode)
	}
	{
		// Unset in the API: read as false.
		fromAPI, diags := SourceReaderFromAPIModel(t.Context(), openapi.PayloadsSourceReader{
			Uuid:          uuid.New(),
			ConnectorUUID: uuid.New(),
		})
		assert.False(t, diags.HasError())
		assert.False(t, fromAPI.MongoDBDeploymentScopedMode.ValueBool())
	}
}

func TestSourceReaderArchiveConnectorRoundTrip(t *testing.T) {
	archiveConnectorUUID := uuid.New()
	model := SourceReader{
		UUID:                   types.StringValue(uuid.NewString()),
		ConnectorUUID:          types.StringValue(uuid.NewString()),
		MSSQLReplicationMethod: types.StringValue("trn_log_backup"),
		ArchiveConnectorUUID:   types.StringValue(archiveConnectorUUID.String()),
	}

	payload, diags := model.ToAPIModel(t.Context())
	assert.False(t, diags.HasError())
	assert.Equal(t, archiveConnectorUUID, *payload.ArchiveConnectorUUID)
	// The API only keeps the archive connector with this location type
	assert.Equal(t, openapi.EnumsMSSQLArchiveLocationTypeSmbShare, *payload.Settings.ArchiveLocationType)
	assert.Equal(t, archiveConnectorUUID, *SourceReaderCreateRequestFromAPIModel(payload).ArchiveConnectorUUID)

	fromAPI, diags := SourceReaderFromAPIModel(t.Context(), payload)
	assert.False(t, diags.HasError())
	assert.Equal(t, archiveConnectorUUID.String(), fromAPI.ArchiveConnectorUUID.ValueString())

	// Without one, neither is sent
	model.ArchiveConnectorUUID = types.StringNull()
	payload, diags = model.ToAPIModel(t.Context())
	assert.False(t, diags.HasError())
	assert.Nil(t, payload.ArchiveConnectorUUID)
	assert.Nil(t, payload.Settings.ArchiveLocationType)
	fromAPI, diags = SourceReaderFromAPIModel(t.Context(), payload)
	assert.False(t, diags.HasError())
	assert.Equal(t, "", fromAPI.ArchiveConnectorUUID.ValueString())
}
