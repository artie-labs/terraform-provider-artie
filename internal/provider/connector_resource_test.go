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

	"terraform-provider-artie/internal/provider/tfmodels"
)

func TestValidateKafkaConfig(t *testing.T) {
	kafkaConfig := func(mechanism, username, password, clientCert, clientKey string) tfmodels.KafkaSharedConfig {
		return tfmodels.KafkaSharedConfig{
			SASLMechanism: types.StringValue(mechanism),
			Username:      types.StringValue(username),
			Password:      types.StringValue(password),
			ClientCert:    types.StringValue(clientCert),
			ClientKey:     types.StringValue(clientKey),
		}
	}
	{
		// SCRAM with a username and password
		assert.False(t, validateKafkaConfig(kafkaConfig("SCRAM-SHA-512", "artie", "secret", "", "")).HasError())
	}
	{
		// No SASL and no credentials
		assert.False(t, validateKafkaConfig(kafkaConfig("", "", "", "", "")).HasError())
	}
	{
		// A password without a mechanism would be dropped by the API
		assert.True(t, validateKafkaConfig(kafkaConfig("", "", "secret", "", "")).HasError())
	}
	{
		// A username without a mechanism is rejected by the API
		assert.True(t, validateKafkaConfig(kafkaConfig("", "artie", "", "", "")).HasError())
	}
	{
		// AWS-MSK-IAM doesn't use a username or password, so the API drops them
		assert.True(t, validateKafkaConfig(kafkaConfig("AWS-MSK-IAM", "artie", "secret", "", "")).HasError())
	}
	{
		// A client key without a client cert would be dropped by the API
		assert.True(t, validateKafkaConfig(kafkaConfig("", "", "", "", "key")).HasError())
	}
	{
		// Mutual TLS
		assert.False(t, validateKafkaConfig(kafkaConfig("", "", "", "cert", "key")).HasError())
	}
	{
		// An unknown mechanism, e.g. from another resource, is checked once it's known
		config := kafkaConfig("", "artie", "secret", "", "")
		config.SASLMechanism = types.StringUnknown()
		assert.False(t, validateKafkaConfig(config).HasError())
	}
}

func TestConnectorValidateConfigUnknownClickHouseAddresses(t *testing.T) {
	ctx := t.Context()
	r := &ConnectorResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	// tfsdk.Config has no setter, so build its value as a State.
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}
	require.False(t, state.SetAttribute(ctx, path.Root("type"), types.StringValue("clickhouse")).HasError())
	// addresses comes from another resource's output, so it's unknown until apply
	require.False(t, state.SetAttribute(ctx, path.Root("clickhouse_config").AtName("addresses"), types.ListUnknown(types.StringType)).HasError())

	req := resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: state.Raw}}
	var resp resource.ValidateConfigResponse
	r.ValidateConfig(ctx, req, &resp)
	assert.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)
}

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
		// documentdb is rejected too: the API would replace its host with the link's DNS entry
		diags := validateConnectorConfig(t, map[string]attr.Value{"type": types.StringValue("documentdb"), "private_link_uuid": types.StringValue(linkA)})
		assert.Equal(t, []string{"`private_link_uuid` isn't supported for `documentdb` connectors."}, privateLinkRejections(diags))
	}
	{
		// planetscale is rejected for both links: it shares MySQL's host and snapshotHost
		diags := validateConnectorConfig(t, map[string]attr.Value{"type": types.StringValue("planetscale"), "snapshot_private_link_uuid": types.StringValue(linkB)})
		assert.Equal(t, []string{"`snapshot_private_link_uuid` isn't supported for `planetscale` connectors."}, privateLinkRejections(diags))
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
		return runStringPlanModifier(t, smbSharePrivateLinkPlanModifier{}, attributes, configValue, types.StringUnknown())
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
		const other = "22222222-2222-2222-2222-222222222222"
		value, diags := runStringPlanModifier(t, smbSharePrivateLinkPlanModifier{}, map[string]attr.Value{"type": types.StringValue("smb_share"), "smb_share_config": smbShareConfigValue(types.StringValue(link))}, types.StringValue(other), types.StringValue(other))
		assert.False(t, diags.HasError(), diags)
		assert.Equal(t, types.StringValue(other), value)
	}
}

func TestConnectorValidateConfigSSHTunnelAndPrivateLink(t *testing.T) {
	const link = "11111111-1111-1111-1111-111111111111"
	const tunnel = "33333333-3333-3333-3333-333333333333"
	conflicts := func(diags diag.Diagnostics) []string {
		var paths []string
		for _, d := range diags.Errors() {
			withPath, ok := d.(diag.DiagnosticWithPath)
			if !ok {
				continue
			}
			if d.Summary() == "Conflicting connection settings" {
				paths = append(paths, withPath.Path().String())
			}
		}
		return paths
	}
	{
		// A tunnel and a link together: rejected, since the API refuses the pair
		diags := validateConnectorConfig(t, map[string]attr.Value{
			"type":                       types.StringValue("postgresql"),
			"ssh_tunnel_uuid":            types.StringValue(tunnel),
			"snapshot_private_link_uuid": types.StringValue(link),
		})
		assert.Equal(t, []string{"snapshot_private_link_uuid"}, conflicts(diags))
	}
	{
		// A tunnel with a link in the SMB share block: rejected too
		diags := validateConnectorConfig(t, map[string]attr.Value{
			"type":             types.StringValue("smb_share"),
			"ssh_tunnel_uuid":  types.StringValue(tunnel),
			"smb_share_config": smbShareConfigValue(types.StringValue(link)),
		})
		assert.Equal(t, []string{"smb_share_config.private_link_uuid"}, conflicts(diags))
	}
	{
		// A tunnel with the link removed: fine
		diags := validateConnectorConfig(t, map[string]attr.Value{
			"type":              types.StringValue("postgresql"),
			"ssh_tunnel_uuid":   types.StringValue(tunnel),
			"private_link_uuid": types.StringValue(""),
		})
		assert.Empty(t, conflicts(diags))
	}
}

func TestPrivateLinkSSHTunnelPlan(t *testing.T) {
	const tunnel = "33333333-3333-3333-3333-333333333333"
	{
		// Switching to a PrivateLink (top level or in the SMB share block): a tunnel left out of the config is removed
		assert.Equal(t, types.StringValue(""), privateLinkSSHTunnelPlan([]privateLinkAttribute{{path.Root("private_link_uuid"), types.StringValue("11111111-1111-1111-1111-111111111111")}}, types.StringValue(tunnel)))
		assert.Equal(t, types.StringValue(""), privateLinkSSHTunnelPlan([]privateLinkAttribute{{path.Root("smb_share_config").AtName("private_link_uuid"), types.StringValue("11111111-1111-1111-1111-111111111111")}}, types.StringValue(tunnel)))
	}
	{
		// A link only known at apply: unknown
		assert.True(t, privateLinkSSHTunnelPlan([]privateLinkAttribute{{path.Root("private_link_uuid"), types.StringUnknown()}}, types.StringValue(tunnel)).IsUnknown())
	}
	{
		// No link (left out or removed): the tunnel from state is kept
		assert.Equal(t, types.StringValue(tunnel), privateLinkSSHTunnelPlan([]privateLinkAttribute{{path.Root("private_link_uuid"), types.StringNull()}, {path.Root("snapshot_private_link_uuid"), types.StringValue("")}}, types.StringValue(tunnel)))
	}
}

func TestSSHTunnelPrivateLinkPlan(t *testing.T) {
	const link = "11111111-1111-1111-1111-111111111111"
	{
		// Switching to an SSH tunnel: a link left out of the config is removed, the way main always sent null
		assert.Equal(t, types.StringValue(""), sshTunnelPrivateLinkPlan(types.StringValue("33333333-3333-3333-3333-333333333333"), types.StringValue(link)))
	}
	{
		// Tunnel only known at apply: unknown
		assert.True(t, sshTunnelPrivateLinkPlan(types.StringUnknown(), types.StringValue(link)).IsUnknown())
	}
	{
		// No tunnel (left out or removed): the link from state is kept
		assert.Equal(t, types.StringValue(link), sshTunnelPrivateLinkPlan(types.StringNull(), types.StringValue(link)))
		assert.Equal(t, types.StringValue(link), sshTunnelPrivateLinkPlan(types.StringValue(""), types.StringValue(link)))
	}
}

// runStringPlanModifier runs a string plan modifier against a connector config with the given attributes set.
func runStringPlanModifier(t *testing.T, modifier planmodifier.String, attributes map[string]attr.Value, configValue, planValue types.String) (types.String, diag.Diagnostics) {
	req := planmodifier.StringRequest{Config: connectorConfig(t, attributes), ConfigValue: configValue, PlanValue: planValue}
	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	modifier.PlanModifyString(t.Context(), req, resp)
	return resp.PlanValue, resp.Diagnostics
}

func TestSSHTunnelAndPrivateLinkPlanModifiers(t *testing.T) {
	const link = "11111111-1111-1111-1111-111111111111"
	const tunnel = "33333333-3333-3333-3333-333333333333"
	{
		// A link left out while the config sets a tunnel is planned as removed
		value, diags := runStringPlanModifier(t, sshTunnelDetachesPrivateLinkPlanModifier{}, map[string]attr.Value{"type": types.StringValue("postgresql"), "ssh_tunnel_uuid": types.StringValue(tunnel)}, types.StringNull(), types.StringValue(link))
		assert.False(t, diags.HasError(), diags)
		assert.Equal(t, types.StringValue(""), value)
	}
	{
		// A link set in the config is left alone
		value, diags := runStringPlanModifier(t, sshTunnelDetachesPrivateLinkPlanModifier{}, map[string]attr.Value{"type": types.StringValue("postgresql"), "ssh_tunnel_uuid": types.StringValue(tunnel)}, types.StringValue(link), types.StringValue(link))
		assert.False(t, diags.HasError(), diags)
		assert.Equal(t, types.StringValue(link), value)
	}
	{
		// A tunnel left out while the config sets a link is planned as removed
		value, diags := runStringPlanModifier(t, privateLinkDetachesSSHTunnelPlanModifier{}, map[string]attr.Value{"type": types.StringValue("postgresql"), "private_link_uuid": types.StringValue(link)}, types.StringNull(), types.StringValue(tunnel))
		assert.False(t, diags.HasError(), diags)
		assert.Equal(t, types.StringValue(""), value)
	}
	{
		// A stray smb_share_config link on another type isn't sent, so it doesn't remove the tunnel
		value, diags := runStringPlanModifier(t, privateLinkDetachesSSHTunnelPlanModifier{}, map[string]attr.Value{"type": types.StringValue("postgresql"), "smb_share_config": smbShareConfigValue(types.StringValue(link))}, types.StringNull(), types.StringValue(tunnel))
		assert.False(t, diags.HasError(), diags)
		assert.Equal(t, types.StringValue(tunnel), value)
	}
}

func TestDetachPlansKeepARemovedValue(t *testing.T) {
	{
		// Already removed: stays "", even while the other side is only known at apply
		assert.Equal(t, types.StringValue(""), sshTunnelPrivateLinkPlan(types.StringUnknown(), types.StringValue("")))
		assert.Equal(t, types.StringValue(""), privateLinkSSHTunnelPlan([]privateLinkAttribute{{path.Root("private_link_uuid"), types.StringUnknown()}}, types.StringValue("")))
	}
}

// privateLinkRejections returns the details of the "PrivateLink not supported" errors.
func privateLinkRejections(diags diag.Diagnostics) []string {
	var rejected []string
	for _, d := range diags.Errors() {
		if d.Summary() == "PrivateLink not supported" {
			rejected = append(rejected, d.Detail())
		}
	}
	return rejected
}
