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
