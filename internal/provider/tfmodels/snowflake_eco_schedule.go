package tfmodels

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/openapi"
)

type SnowflakeEcoSchedule struct {
	UUID                        types.String   `tfsdk:"uuid"`
	Name                        types.String   `tfsdk:"name"`
	Timezone                    types.String   `tfsdk:"timezone"`
	PeakStartTime               types.String   `tfsdk:"peak_start_time"`
	PeakEndTime                 types.String   `tfsdk:"peak_end_time"`
	PeakDays                    []types.String `tfsdk:"peak_days"`
	OffPeakFlushIntervalSeconds types.Int64    `tfsdk:"off_peak_flush_interval_seconds"`
	OffPeakBufferRows           types.Int64    `tfsdk:"off_peak_buffer_rows"`
	OffPeakFlushSizeKb          types.Int64    `tfsdk:"off_peak_flush_size_kb"`
}

func (s SnowflakeEcoSchedule) settings() openapi.OrmSnowflakeEcoScheduleSettings {
	days := make([]string, len(s.PeakDays))
	for i, day := range s.PeakDays {
		days[i] = day.ValueString()
	}
	return openapi.OrmSnowflakeEcoScheduleSettings{
		Timezone:                    s.Timezone.ValueString(),
		StartTimeOfDay:              s.PeakStartTime.ValueString(),
		EndTimeOfDay:                s.PeakEndTime.ValueString(),
		Days:                        days,
		OffPeakFlushIntervalSeconds: int(s.OffPeakFlushIntervalSeconds.ValueInt64()),
		OffPeakBufferRows:           int(s.OffPeakBufferRows.ValueInt64()),
		OffPeakFlushSizeKb:          int(s.OffPeakFlushSizeKb.ValueInt64()),
	}
}

func (s SnowflakeEcoSchedule) ToAPICreateRequest() openapi.RouterSnowflakeEcoScheduleCreateRequest {
	return openapi.RouterSnowflakeEcoScheduleCreateRequest{Name: s.Name.ValueString(), Settings: s.settings()}
}

func (s SnowflakeEcoSchedule) ToAPIUpdateRequest() openapi.RouterSnowflakeEcoScheduleUpdateRequest {
	return openapi.RouterSnowflakeEcoScheduleUpdateRequest{Name: s.Name.ValueString(), Settings: s.settings()}
}

func SnowflakeEcoScheduleFromAPIModel(apiModel openapi.PayloadsSnowflakeEcoSchedule) SnowflakeEcoSchedule {
	days := make([]types.String, len(apiModel.Settings.Days))
	for i, day := range apiModel.Settings.Days {
		days[i] = types.StringValue(day)
	}
	return SnowflakeEcoSchedule{
		UUID:                        types.StringValue(apiModel.Uuid.String()),
		Name:                        types.StringValue(apiModel.Name),
		Timezone:                    types.StringValue(apiModel.Settings.Timezone),
		PeakStartTime:               types.StringValue(apiModel.Settings.StartTimeOfDay),
		PeakEndTime:                 types.StringValue(apiModel.Settings.EndTimeOfDay),
		PeakDays:                    days,
		OffPeakFlushIntervalSeconds: types.Int64Value(int64(apiModel.Settings.OffPeakFlushIntervalSeconds)),
		OffPeakBufferRows:           types.Int64Value(int64(apiModel.Settings.OffPeakBufferRows)),
		OffPeakFlushSizeKb:          types.Int64Value(int64(apiModel.Settings.OffPeakFlushSizeKb)),
	}
}
