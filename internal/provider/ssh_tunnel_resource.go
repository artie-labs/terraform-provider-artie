package provider

import (
	"context"
	"fmt"
	"math"
	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/provider/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
		MarkdownDescription: "Artie SSH Tunnel resource",
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

func (r *SSHTunnelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data models.SSHTunnelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating SSH Tunnel via API")
	sshTunnel, err := r.client.SSHTunnels().Create(ctx, data.Name.ValueString(), data.Host.ValueString(), data.Username.ValueString(), data.Port.ValueInt32())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create SSH Tunnel", err.Error())
		return
	}

	tflog.Info(ctx, "Created SSH Tunnel")

	// Translate API response into Terraform state
	models.SSHTunnelAPIToResourceModel(sshTunnel, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SSHTunnelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var data models.SSHTunnelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading SSH Tunnel from API")
	sshTunnel, err := r.client.SSHTunnels().Get(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read SSH Tunnel", err.Error())
		return
	}

	// Translate API response into Terraform state
	models.SSHTunnelAPIToResourceModel(sshTunnel, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SSHTunnelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data models.SSHTunnelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sshTunnel := models.SSHTunnelResourceToAPIModel(data)
	tflog.Info(ctx, "Updating SSH Tunnel via API")
	sshTunnel, err := r.client.SSHTunnels().Update(ctx, sshTunnel)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update SSH Tunnel", err.Error())
		return
	}

	// Translate API response into Terraform state
	models.SSHTunnelAPIToResourceModel(sshTunnel, &data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SSHTunnelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform prior state data into the model
	var data models.SSHTunnelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting SSH Tunnel via API")
	if err := r.client.SSHTunnels().Delete(ctx, data.UUID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to Delete SSH Tunnel", err.Error())
	}
}

func (r *SSHTunnelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
