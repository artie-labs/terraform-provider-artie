package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

// Runs the resource's data_plane_name plan modifiers in order, the way the framework does for an update.
func planDataPlane(t *testing.T, r resource.Resource, config, state types.String) (types.String, bool) {
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	attr, ok := schemaResp.Schema.Attributes["data_plane_name"].(schema.StringAttribute)
	assert.True(t, ok)

	raw := tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})
	plan := config
	if config.IsNull() {
		// Unset optional+computed attributes start out unknown.
		plan = types.StringUnknown()
	}
	var replace bool
	for _, m := range attr.PlanModifiers {
		req := planmodifier.StringRequest{ConfigValue: config, StateValue: state, PlanValue: plan, State: tfsdk.State{Raw: raw}, Plan: tfsdk.Plan{Raw: raw}}
		resp := &planmodifier.StringResponse{PlanValue: plan}
		m.PlanModifyString(ctx, req, resp)
		assert.False(t, resp.Diagnostics.HasError())
		plan = resp.PlanValue
		replace = replace || resp.RequiresReplace
	}
	return plan, replace
}

func assertDataPlanePlan(t *testing.T, r resource.Resource) {
	{
		// Left out of the config: keeps the server's value, no replacement.
		plan, replace := planDataPlane(t, r, types.StringNull(), types.StringValue("a"))
		assert.Equal(t, types.StringValue("a"), plan)
		assert.False(t, replace)
	}
	{
		// Unchanged: no replacement.
		_, replace := planDataPlane(t, r, types.StringValue("a"), types.StringValue("a"))
		assert.False(t, replace)
	}
	{
		// Moved to another data plane, or to one only known at apply time: replaced.
		_, replace := planDataPlane(t, r, types.StringValue("b"), types.StringValue("a"))
		assert.True(t, replace)
		_, replace = planDataPlane(t, r, types.StringUnknown(), types.StringValue("a"))
		assert.True(t, replace)
	}
	{
		// No data plane yet, or set to "" (the default): updated in place.
		_, replace := planDataPlane(t, r, types.StringValue("b"), types.StringValue(""))
		assert.False(t, replace)
		_, replace = planDataPlane(t, r, types.StringValue(""), types.StringValue("a"))
		assert.False(t, replace)
	}
}

func TestDataPlanePlan(t *testing.T) {
	{
		// Connector
		assertDataPlanePlan(t, &ConnectorResource{})
	}
	{
		// Pipeline
		assertDataPlanePlan(t, &PipelineResource{})
	}
	{
		// Source reader
		assertDataPlanePlan(t, &SourceReaderResource{})
	}
}
