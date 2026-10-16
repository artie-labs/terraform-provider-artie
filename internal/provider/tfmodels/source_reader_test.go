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
