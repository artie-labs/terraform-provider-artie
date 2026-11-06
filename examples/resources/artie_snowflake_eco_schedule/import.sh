# Import a Snowflake Eco Mode schedule by its UUID. List your schedules' UUIDs with
# GET https://api.artie.com/snowflake-eco-schedules, or read snowflake_eco_schedule_uuid
# from an artie_pipeline that uses the schedule.
terraform import artie_snowflake_eco_schedule.my_schedule <snowflake_eco_schedule_uuid>

# Then print the state and copy it into your terraform config file
# (be sure to remove all read-only fields, like `uuid`):
terraform state show artie_snowflake_eco_schedule.my_schedule
