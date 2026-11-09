resource "artie_snowflake_eco_schedule" "business_hours" {
  name            = "Business hours"
  timezone        = "America/New_York"
  peak_days       = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]
  peak_start_time = "09:00"
  peak_end_time   = "18:00"

  # Outside peak hours (weeknights and all weekend), flush every 30 minutes.
  off_peak_flush_interval_seconds = 1800
  off_peak_buffer_rows            = 500000
  off_peak_flush_size_kb          = 250000
}

resource "artie_pipeline" "my_pipeline" {
  # ...
  snowflake_eco_schedule_uuid = artie_snowflake_eco_schedule.business_hours.uuid
}
