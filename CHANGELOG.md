## 0.1.0 (Unreleased)

FEATURES:

BUG FIXES:

* `artie_pipeline`: Fixed a "provider produced an inconsistent result after apply"
  error when a table-level boolean setting (e.g. `encrypt_jsonb_columns`) was set
  explicitly to `false`. The API persists these "absent means off" toggles as null
  when false, so they now read back as `false` instead of `null`, allowing an
  explicit `false` to round-trip consistently.
