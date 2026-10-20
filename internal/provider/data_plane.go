package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const dataPlaneReplaceNote = "Terraform deletes the old resource before the new data plane is validated, so a wrong name leaves it deleted; add `lifecycle { prevent_destroy = true }` to block replacements entirely. If an administrator moved this resource to another data plane, update `data_plane_name` to match before applying, or Terraform will recreate it in the old one."

// dataPlaneRequiresReplace replaces a resource that moves off its data plane, which the API rejects as an update.
func dataPlaneRequiresReplace() planmodifier.String {
	const desc = "Moving to a different data plane forces a new resource."
	return stringplanmodifier.RequiresReplaceIf(func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
		resp.RequiresReplace = movesDataPlane(req.StateValue, req.PlanValue)
	}, desc, desc)
}

// The API lets a resource without a data plane gain one in place, and treats "" as "use the default", so neither replaces.
func movesDataPlane(state, plan types.String) bool {
	if state.ValueString() == "" {
		return false
	}
	return plan.IsUnknown() || (plan.ValueString() != "" && plan.ValueString() != state.ValueString())
}
