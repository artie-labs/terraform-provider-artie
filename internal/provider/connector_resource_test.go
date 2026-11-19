package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
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

// connectorConfig builds a connector config from the resource schema with the given attributes set; everything else is
// null. A name like "oracle_config.password" sets a nested attribute.
func connectorConfig(t *testing.T, attributes map[string]attr.Value) tfsdk.Config {
	ctx := t.Context()
	var schemaResp resource.SchemaResponse
	(&ConnectorResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	// tfsdk.Config has no setter, so build its value as a State.
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)}
	for name, value := range attributes {
		parts := strings.Split(name, ".")
		attributePath := path.Root(parts[0])
		for _, part := range parts[1:] {
			attributePath = attributePath.AtName(part)
		}
		require.False(t, state.SetAttribute(ctx, attributePath, value).HasError(), name)
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

// validateConnectorThroughServer runs the provider server's config validation, which unlike [validateConnectorConfig]
// includes attribute validators.
func validateConnectorThroughServer(t *testing.T, attributes map[string]attr.Value, writeOnlyAllowed bool) []*tfprotov6.Diagnostic {
	server := providerserver.NewProtocol6(New("test")())()
	raw := connectorConfig(t, attributes).Raw
	config, err := tfprotov6.NewDynamicValue(raw.Type(), raw)
	require.NoError(t, err)
	resp, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{
		TypeName:           "artie_connector",
		Config:             &config,
		ClientCapabilities: &tfprotov6.ValidateResourceConfigClientCapabilities{WriteOnlyAttributesAllowed: writeOnlyAllowed},
	})
	require.NoError(t, err)
	return resp.Diagnostics
}

func TestConnectorValidateConfigWriteOnlySecrets(t *testing.T) {
	oracle := func(password, passwordWO types.String) map[string]attr.Value {
		return map[string]attr.Value{
			"type":                      types.StringValue("oracle"),
			"oracle_config.host":        types.StringValue("oracle.example.com"),
			"oracle_config.port":        types.Int32Value(1521),
			"oracle_config.username":    types.StringValue("artie"),
			"oracle_config.password":    password,
			"oracle_config.password_wo": passwordWO,
		}
	}
	{
		// A config from before write-only secrets still validates
		assert.Empty(t, validateConnectorThroughServer(t, oracle(types.StringValue("secret"), types.StringNull()), true))
	}
	{
		// password_wo alone validates
		assert.Empty(t, validateConnectorThroughServer(t, oracle(types.StringNull(), types.StringValue("secret")), true))
	}
	{
		// Both is an error
		assert.NotEmpty(t, validateConnectorThroughServer(t, oracle(types.StringValue("secret"), types.StringValue("secret")), true))
	}
	{
		// Neither is an error, as when password was required
		assert.NotEmpty(t, validateConnectorThroughServer(t, oracle(types.StringNull(), types.StringNull()), true))
	}
	{
		// An empty write-only secret is an error
		assert.NotEmpty(t, validateConnectorThroughServer(t, oracle(types.StringNull(), types.StringValue("")), true))
	}
	{
		// Terraform before 1.11 can't send write-only values
		assert.NotEmpty(t, validateConnectorThroughServer(t, oracle(types.StringNull(), types.StringValue("secret")), false))
	}
	{
		// A version left behind after moving back to password is an error, so the change shows in the plan
		config := oracle(types.StringValue("secret"), types.StringNull())
		config["oracle_config.password_wo_version"] = types.Int64Value(1)
		assert.NotEmpty(t, validateConnectorThroughServer(t, config, true))
	}
	{
		config := oracle(types.StringNull(), types.StringValue("secret"))
		config["oracle_config.password_wo_version"] = types.Int64Value(1)
		assert.Empty(t, validateConnectorThroughServer(t, config, true))
	}
}

func TestConnectorValidateConfigWriteOnlyOptionalSecret(t *testing.T) {
	clickHouse := func(password, passwordWO types.String) map[string]attr.Value {
		return map[string]attr.Value{
			"type":                          types.StringValue("clickhouse"),
			"clickhouse_config.addresses":   types.ListValueMust(types.StringType, []attr.Value{types.StringValue("ch.example.com:9440")}),
			"clickhouse_config.password":    password,
			"clickhouse_config.password_wo": passwordWO,
		}
	}
	{
		// A config from before write-only secrets still validates
		assert.Empty(t, validateConnectorThroughServer(t, clickHouse(types.StringValue("secret"), types.StringNull()), true))
	}
	{
		// password_wo alone validates
		assert.Empty(t, validateConnectorThroughServer(t, clickHouse(types.StringNull(), types.StringValue("secret")), true))
	}
	{
		// The password stays optional, so neither validates
		assert.Empty(t, validateConnectorThroughServer(t, clickHouse(types.StringNull(), types.StringNull()), true))
	}
	{
		// Both is an error
		assert.NotEmpty(t, validateConnectorThroughServer(t, clickHouse(types.StringValue("secret"), types.StringValue("secret")), true))
	}
	{
		// A version left behind after removing password_wo is an error, so the removal shows in the plan
		config := clickHouse(types.StringNull(), types.StringNull())
		config["clickhouse_config.password_wo_version"] = types.Int64Value(1)
		assert.NotEmpty(t, validateConnectorThroughServer(t, config, true))
	}
}
