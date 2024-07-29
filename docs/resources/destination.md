---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "artie_destination Resource - artie"
subcategory: ""
description: |-
  Artie Destination resource
---

# artie_destination (Resource)

Artie Destination resource



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `config` (Attributes) (see [below for nested schema](#nestedatt--config))
- `name` (String)

### Optional

- `label` (String)

### Read-Only

- `company_uuid` (String)
- `last_updated_at` (String)
- `ssh_tunnel_uuid` (String)
- `uuid` (String)

<a id="nestedatt--config"></a>
### Nested Schema for `config`

Optional:

- `aws_access_key_id` (String)
- `aws_region` (String)
- `endpoint` (String)
- `gcp_location` (String)
- `gcp_project_id` (String)
- `host` (String)
- `port` (Number)
- `snowflake_account_url` (String)
- `snowflake_virtual_dwh` (String)
- `username` (String)