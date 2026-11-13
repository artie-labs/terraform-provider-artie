package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// connectorConfig builds a connector config from the resource schema with the given attributes set; everything else is null.
func connectorConfig(t *testing.T, attributes map[string]attr.Value) tfsdk.Config {
	ctx := t.Context()
	var schemaResp resource.SchemaResponse
	(&ConnectorResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	// tfsdk.Config has no setter, so build its value as a State.
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}
	for name, value := range attributes {
		require.False(t, state.SetAttribute(ctx, path.Root(name), value).HasError(), name)
	}
	return tfsdk.Config{Schema: schemaResp.Schema, Raw: state.Raw}
}

func validateConnectorConfig(t *testing.T, attributes map[string]attr.Value) diag.Diagnostics {
	var resp resource.ValidateConfigResponse
	(&ConnectorResource{}).ValidateConfig(t.Context(), resource.ValidateConfigRequest{Config: connectorConfig(t, attributes)}, &resp)
	return resp.Diagnostics
}

var smbShareConfigAttrTypes = map[string]attr.Type{"path": types.StringType, "username": types.StringType, "password": types.StringType, "private_link_uuid": types.StringType}

func smbShareConfigValue(privateLinkUUID types.String) types.Object {
	return types.ObjectValueMust(smbShareConfigAttrTypes, map[string]attr.Value{
		"path":              types.StringValue(`\\share.example.com\data`),
		"username":          types.StringValue("artie"),
		"password":          types.StringValue("secret"),
		"private_link_uuid": privateLinkUUID,
	})
}

var mongoDBConfig = types.ObjectValueMust(
	map[string]attr.Type{"host": types.StringType, "username": types.StringType, "password": types.StringType},
	map[string]attr.Value{"host": types.StringValue("mongodb+srv://cluster.example.com"), "username": types.StringValue("artie"), "password": types.StringValue("secret")},
)

func TestConnectorValidateConfigPrivateLinks(t *testing.T) {
	const linkA = "11111111-1111-1111-1111-111111111111"
	const linkB = "22222222-2222-2222-2222-222222222222"
	{
		// mongodb rejects a PrivateLink: the API would replace the connection string with the link's DNS entry
		diags := validateConnectorConfig(t, map[string]attr.Value{
			"type":                       types.StringValue("mongodb"),
			"mongodb_config":             mongoDBConfig,
			"private_link_uuid":          types.StringValue(linkA),
			"snapshot_private_link_uuid": types.StringValue(linkB),
		})
		assert.Equal(t, 2, diags.ErrorsCount(), diags)
	}
	{
		// mongodb with the link removed or left out is fine
		diags := validateConnectorConfig(t, map[string]attr.Value{"type": types.StringValue("mongodb"), "mongodb_config": mongoDBConfig, "private_link_uuid": types.StringValue("")})
		assert.False(t, diags.HasError(), diags)
	}
	{
		// smb_share: different top-level and block links conflict
		diags := validateConnectorConfig(t, map[string]attr.Value{
			"type":              types.StringValue("smb_share"),
			"private_link_uuid": types.StringValue(linkA),
			"smb_share_config":  smbShareConfigValue(types.StringValue(linkB)),
		})
		assert.True(t, diags.HasError(), diags)
	}
	{
		// smb_share: an explicit "" at the top level conflicts with a block link
		diags := validateConnectorConfig(t, map[string]attr.Value{
			"type":              types.StringValue("smb_share"),
			"private_link_uuid": types.StringValue(""),
			"smb_share_config":  smbShareConfigValue(types.StringValue(linkB)),
		})
		assert.True(t, diags.HasError(), diags)
	}
	{
		// smb_share: the block link alone is fine
		diags := validateConnectorConfig(t, map[string]attr.Value{
			"type":             types.StringValue("smb_share"),
			"smb_share_config": smbShareConfigValue(types.StringValue(linkB)),
		})
		assert.False(t, diags.HasError(), diags)
	}
}

func TestSMBSharePrivateLinkPlanModifier(t *testing.T) {
	const link = "11111111-1111-1111-1111-111111111111"
	plan := func(attributes map[string]attr.Value, configValue types.String) (types.String, diag.Diagnostics) {
		req := planmodifier.StringRequest{Config: connectorConfig(t, attributes), ConfigValue: configValue, PlanValue: types.StringUnknown()}
		resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
		smbSharePrivateLinkPlanModifier{}.PlanModifyString(t.Context(), req, resp)
		return resp.PlanValue, resp.Diagnostics
	}
	{
		// another type with no smb_share_config: reading through the null block works and the plan is unchanged
		value, diags := plan(map[string]attr.Value{"type": types.StringValue("postgresql")}, types.StringNull())
		assert.False(t, diags.HasError(), diags)
		assert.True(t, value.IsUnknown())
	}
	{
		// smb_share leaving the top-level value out follows the block
		value, diags := plan(map[string]attr.Value{"type": types.StringValue("smb_share"), "smb_share_config": smbShareConfigValue(types.StringValue(link))}, types.StringNull())
		assert.False(t, diags.HasError(), diags)
		assert.Equal(t, types.StringValue(link), value)
	}
	{
		// smb_share with an unknown block plans unknown
		unknownBlock := types.ObjectUnknown(smbShareConfigAttrTypes)
		value, diags := plan(map[string]attr.Value{"type": types.StringValue("smb_share"), "smb_share_config": unknownBlock}, types.StringNull())
		assert.False(t, diags.HasError(), diags)
		assert.True(t, value.IsUnknown())
	}
	{
		// smb_share setting the top-level value keeps the planned value
		value, diags := plan(map[string]attr.Value{"type": types.StringValue("smb_share"), "smb_share_config": smbShareConfigValue(types.StringValue(link))}, types.StringValue(link))
		assert.False(t, diags.HasError(), diags)
		assert.True(t, value.IsUnknown())
	}
}
