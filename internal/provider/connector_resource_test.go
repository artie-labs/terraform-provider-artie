package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
