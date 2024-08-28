package provider

import (
	"context"
	"fmt"
	"math"
	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/provider/tfmodels"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SSHTunnelResource{}
var _ resource.ResourceWithConfigure = &SSHTunnelResource{}
var _ resource.ResourceWithImportState = &SSHTunnelResource{}

func NewSSHTunnelResource() resource.Resource {
	return &SSHTunnelResource{}
}

type SSHTunnelResource struct {
	client artieclient.Client
}

func (r *SSHTunnelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_tunnel"
}

func (r *SSHTunnelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie SSH Tunnel resource. This resource allows you to create an SSH tunnel to connect to your source or destination databases.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Required: true, MarkdownDescription: "A human-readable label for this SSH tunnel."},
			"host": schema.StringAttribute{Required: true, MarkdownDescription: "The public hostname or IP address of your SSH server."},
			"port": schema.Int32Attribute{
				Required:            true,
				MarkdownDescription: "The port number of your SSH server.",
				Validators: []validator.Int32{
					int32validator.Between(22, math.MaxUint16),
				},
			},
			"username": schema.StringAttribute{Required: true, MarkdownDescription: "The username we should use when connecting to your SSH server."},
			"public_key": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "When you create an SSH tunnel in Artie, we generate a public/private key pair. Once generated, you'll need to add this public key to `~/.ssh/authorized_keys` on your SSH server.",
			},
		},
	}
}

func (r *SSHTunnelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SSHTunnelResource) GetUUIDFromState(ctx context.Context, state tfsdk.State, diagnostics diag.Diagnostics) (string, bool) {
	var stateData tfmodels.SSHTunnel
	diagnostics.Append(state.Get(ctx, &stateData)...)
	return stateData.UUID.ValueString(), diagnostics.HasError()
}

func (r *SSHTunnelResource) GetPlanData(ctx context.Context, plan tfsdk.Plan, diagnostics diag.Diagnostics) (tfmodels.SSHTunnel, bool) {
	var planData tfmodels.SSHTunnel
	diagnostics.Append(plan.Get(ctx, &planData)...)
	return planData, diagnostics.HasError()
}

func (r *SSHTunnelResource) SetStateData(ctx context.Context, state *tfsdk.State, diagnostics diag.Diagnostics, sshTunnel artieclient.SSHTunnel) {
	// Translate API response type into Terraform model and save it into state
	diagnostics.Append(state.Set(ctx, tfmodels.SSHTunnelFromAPIModel(sshTunnel))...)
}

func (r *SSHTunnelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, resp.Diagnostics)
	if hasError {
		return
	}

	sshTunnel, err := r.client.SSHTunnels().Create(ctx, planData.ToAPIBaseModel())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create SSH Tunnel", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, resp.Diagnostics, sshTunnel)
}

func (r *SSHTunnelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tunnelUUID, hasError := r.GetUUIDFromState(ctx, req.State, resp.Diagnostics)
	if hasError {
		return
	}

	sshTunnel, err := r.client.SSHTunnels().Get(ctx, tunnelUUID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read SSH Tunnel", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, resp.Diagnostics, sshTunnel)
}

func (r *SSHTunnelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, resp.Diagnostics)
	if hasError {
		return
	}

	sshTunnel, err := r.client.SSHTunnels().Update(ctx, planData.ToAPIModel())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update SSH Tunnel", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, resp.Diagnostics, sshTunnel)
}

func (r *SSHTunnelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tunnelUUID, hasError := r.GetUUIDFromState(ctx, req.State, resp.Diagnostics)
	if hasError {
		return
	}

	if err := r.client.SSHTunnels().Delete(ctx, tunnelUUID); err != nil {
		resp.Diagnostics.AddError("Unable to Delete SSH Tunnel", err.Error())
	}
}

func (r *SSHTunnelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
