## 0.1.0 (Unreleased)

FEATURES:

* **New Resource:** `artie_column_hashing_salt` — manage salts used to hash column values. If `salt` is omitted, Artie generates a strong random value; the salt is sensitive and cannot be rotated in place, so changing it forces replacement. A salt still referenced by any pipeline cannot be deleted until the pipeline is updated to clear or change `column_hashing_salt_uuid`.
* **New Data Source:** `artie_column_hashing_salt` — look up an existing column hashing salt by UUID (useful when the salt was created outside Terraform).
* **Enhancement:** `artie_pipeline` gains an optional `column_hashing_salt_uuid` attribute. Set this (for example, to `artie_column_hashing_salt.main.uuid`) whenever a table has `columns_to_hash` configured.
