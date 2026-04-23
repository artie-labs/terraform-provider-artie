package provider

import (
	"context"
	"fmt"

	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/provider/tfmodels"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ resource.Resource = &ColumnHashingSaltResource{}
var _ resource.ResourceWithConfigure = &ColumnHashingSaltResource{}
var _ resource.ResourceWithImportState = &ColumnHashingSaltResource{}

func NewColumnHashingSaltResource() resource.Resource {
	return &ColumnHashingSaltResource{}
}

type ColumnHashingSaltResource struct {
	client artieclient.Client
}

func (r *ColumnHashingSaltResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_column_hashing_salt"
}

func (r *ColumnHashingSaltResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Column Hashing Salt resource. This resource manages salts used when hashing column values in pipelines. Reference a salt's `uuid` from an `artie_pipeline`'s `column_hashing_salt_uuid` to apply it to any table that has `columns_to_hash` set.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A human-readable name for the column hashing salt.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "A description of the salt's purpose.",
			},
			"salt": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The salt value used when hashing column values. If omitted, Artie will generate a strong random salt. This value is sensitive and cannot be rotated in place; changing it forces replacement.",
			},
		},
	}
}

func (r *ColumnHashingSaltResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(ArtieProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected ArtieProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}

	client, err := providerData.NewClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to build Artie client", err.Error())
		return
	}

	r.client = client
}

func (r *ColumnHashingSaltResource) GetUUIDFromState(ctx context.Context, state tfsdk.State, diagnostics *diag.Diagnostics) (string, bool) {
	var stateData tfmodels.ColumnHashingSalt
	diagnostics.Append(state.Get(ctx, &stateData)...)
	return stateData.UUID.ValueString(), diagnostics.HasError()
}

func (r *ColumnHashingSaltResource) GetPlanData(ctx context.Context, plan tfsdk.Plan, diagnostics *diag.Diagnostics) (tfmodels.ColumnHashingSalt, bool) {
	var planData tfmodels.ColumnHashingSalt
	diagnostics.Append(plan.Get(ctx, &planData)...)
	return planData, diagnostics.HasError()
}

func (r *ColumnHashingSaltResource) SetStateData(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, apiModel artieclient.ColumnHashingSalt) {
	diagnostics.Append(state.Set(ctx, tfmodels.ColumnHashingSaltFromAPIModel(apiModel))...)
}

func (r *ColumnHashingSaltResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	salt, err := r.client.ColumnHashingSalts().Create(ctx, planData.ToAPIBaseModel())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Column Hashing Salt", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, salt)
}

func (r *ColumnHashingSaltResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateData tfmodels.ColumnHashingSalt
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	salt, err := r.client.ColumnHashingSalts().Get(ctx, stateData.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Column Hashing Salt", err.Error())
		return
	}

	// The API only returns the salt value on Create, so preserve the value from state.
	if salt.Salt == "" {
		salt.Salt = stateData.Salt.ValueString()
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, salt)
}

func (r *ColumnHashingSaltResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	var stateData tfmodels.ColumnHashingSalt
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	saltUUID := planData.UUID.ValueString()
	salt, err := r.client.ColumnHashingSalts().Update(ctx, saltUUID, artieclient.UpdateColumnHashingSaltRequest{
		Name:        planData.Name.ValueString(),
		Description: planData.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Column Hashing Salt", err.Error())
		return
	}

	// The API only returns the salt value on Create, so preserve the value from state.
	if salt.Salt == "" {
		salt.Salt = stateData.Salt.ValueString()
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, salt)
}

func (r *ColumnHashingSaltResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	saltUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	if err := r.client.ColumnHashingSalts().Delete(ctx, saltUUID); err != nil {
		resp.Diagnostics.AddError("Unable to delete Column Hashing Salt", err.Error())
	}
}

func (r *ColumnHashingSaltResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
