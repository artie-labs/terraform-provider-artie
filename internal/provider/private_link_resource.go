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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PrivateLinkResource{}
var _ resource.ResourceWithConfigure = &PrivateLinkResource{}
var _ resource.ResourceWithImportState = &PrivateLinkResource{}

func NewPrivateLinkResource() resource.Resource {
	return &PrivateLinkResource{}
}

type PrivateLinkResource struct {
	client artieclient.Client
}

func (r *PrivateLinkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_link"
}

func (r *PrivateLinkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie PrivateLink connection resource. This resource represents an AWS PrivateLink connection to Artie.",
		Attributes: map[string]schema.Attribute{
			"uuid":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name":            schema.StringAttribute{Required: true, MarkdownDescription: "A human-readable label for this PrivateLink connection."},
			"aws_account_id":  schema.StringAttribute{Required: true, MarkdownDescription: "Your AWS Account ID that owns the VPC endpoint."},
			"aws_region":      schema.StringAttribute{Required: true, MarkdownDescription: "The AWS region of the VPC endpoint."},
			"vpc_endpoint_id": schema.StringAttribute{Required: true, MarkdownDescription: "The VPC Endpoint ID (e.g., vpce-xxxxxxxxxxxxxxxxx) that connects to Artie's endpoint service."},
			"status":          schema.StringAttribute{Computed: true, MarkdownDescription: "The status of the PrivateLink connection."},
			"service_name":    schema.StringAttribute{Computed: true, MarkdownDescription: "The endpoint service name in AWS for Artie."},
		},
	}
}

func (r *PrivateLinkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(ArtieProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected ArtieProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	client, err := providerData.NewClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to build Artie client", err.Error())
		return
	}

	r.client = client
}

func (r *PrivateLinkResource) GetUUIDFromState(ctx context.Context, state tfsdk.State, diagnostics *diag.Diagnostics) (string, bool) {
	var stateData tfmodels.PrivateLink
	diagnostics.Append(state.Get(ctx, &stateData)...)
	return stateData.UUID.ValueString(), diagnostics.HasError()
}

func (r *PrivateLinkResource) GetPlanData(ctx context.Context, plan tfsdk.Plan, diagnostics *diag.Diagnostics) (tfmodels.PrivateLink, bool) {
	var planData tfmodels.PrivateLink
	diagnostics.Append(plan.Get(ctx, &planData)...)
	return planData, diagnostics.HasError()
}

func (r *PrivateLinkResource) SetStateData(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, pl artieclient.PrivateLinkConnection) {
	diagnostics.Append(state.Set(ctx, tfmodels.PrivateLinkFromAPIModel(pl))...)
}

func (r *PrivateLinkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	conn, err := r.client.PrivateLinks().Create(ctx, planData.ToAPIBaseModel())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create PrivateLink connection", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, conn)
}

func (r *PrivateLinkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	uuid, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	conn, err := r.client.PrivateLinks().Get(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read PrivateLink connection", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, conn)
}

func (r *PrivateLinkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	apiModel, diags := planData.ToAPIModel()
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	conn, err := r.client.PrivateLinks().Update(ctx, apiModel)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update PrivateLink connection", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, conn)
}

func (r *PrivateLinkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	uuid, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	if err := r.client.PrivateLinks().Delete(ctx, uuid); err != nil {
		resp.Diagnostics.AddError("Unable to delete PrivateLink connection", err.Error())
		return
	}
}

func (r *PrivateLinkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
