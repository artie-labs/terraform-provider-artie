package tfmodels

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/openapi"
)

func TestSnowflakeEcoScheduleRoundTrip(t *testing.T) {
	apiModel := openapi.PayloadsSnowflakeEcoSchedule{
		Uuid: uuid.New(),
		Name: "Business hours",
		Settings: openapi.PayloadsSnowflakeEcoScheduleSettings{
			Timezone:                    "America/New_York",
			StartTimeOfDay:              "09:00",
			EndTimeOfDay:                "18:00",
			Days:                        []string{"Monday", "Friday"},
			OffPeakFlushIntervalSeconds: 1800,
			OffPeakBufferRows:           500000,
			OffPeakFlushSizeKb:          250000,
		},
	}

	tfModel := SnowflakeEcoScheduleFromAPIModel(apiModel)
	assert.Equal(t, "09:00", tfModel.PeakStartTime.ValueString())
	assert.Equal(t, openapi.RouterSnowflakeEcoScheduleCreateRequest{Name: apiModel.Name, Settings: apiModel.Settings}, tfModel.ToAPICreateRequest())
	assert.Equal(t, openapi.RouterSnowflakeEcoScheduleUpdateRequest{Name: apiModel.Name, Settings: apiModel.Settings}, tfModel.ToAPIUpdateRequest())
}
