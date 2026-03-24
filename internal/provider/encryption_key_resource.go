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

var _ resource.Resource = &EncryptionKeyResource{}
var _ resource.ResourceWithConfigure = &EncryptionKeyResource{}
var _ resource.ResourceWithImportState = &EncryptionKeyResource{}

func NewEncryptionKeyResource() resource.Resource {
	return &EncryptionKeyResource{}
}

type EncryptionKeyResource struct {
	client artieclient.Client
}

func (r *EncryptionKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_encryption_key"
}

func (r *EncryptionKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Encryption Key resource. This resource manages encryption keys used for column-level encryption in pipelines. Keys can be passphrase-based (default) or KMS-backed.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "A human-readable name for the encryption key.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "A description of the key's purpose.",
			},
			"kms_key_uuid": schema.StringAttribute{
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "UUID of an `artie_kms_key`. If omitted, a passphrase-type key is generated. Changing this forces replacement.",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The type of the encryption key: `passphrase` or `kms`.",
			},
			"key": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "The key material (passphrase or encrypted DEK). This value is sensitive and will not be displayed in plan output.",
			},
		},
	}
}

func (r *EncryptionKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EncryptionKeyResource) GetUUIDFromState(ctx context.Context, state tfsdk.State, diagnostics *diag.Diagnostics) (string, bool) {
	var stateData tfmodels.EncryptionKey
	diagnostics.Append(state.Get(ctx, &stateData)...)
	return stateData.UUID.ValueString(), diagnostics.HasError()
}

func (r *EncryptionKeyResource) GetPlanData(ctx context.Context, plan tfsdk.Plan, diagnostics *diag.Diagnostics) (tfmodels.EncryptionKey, bool) {
	var planData tfmodels.EncryptionKey
	diagnostics.Append(plan.Get(ctx, &planData)...)
	return planData, diagnostics.HasError()
}

func (r *EncryptionKeyResource) SetStateData(ctx context.Context, state *tfsdk.State, diagnostics *diag.Diagnostics, apiModel artieclient.EncryptionKey) {
	diagnostics.Append(state.Set(ctx, tfmodels.EncryptionKeyFromAPIModel(apiModel))...)
}

func (r *EncryptionKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	apiBaseModel, diags := planData.ToAPIBaseModel()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	encryptionKey, err := r.client.EncryptionKeys().Create(ctx, apiBaseModel)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Encryption Key", err.Error())
		return
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, encryptionKey)
}

func (r *EncryptionKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateData tfmodels.EncryptionKey
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	encryptionKey, err := r.client.EncryptionKeys().Get(ctx, stateData.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Encryption Key", err.Error())
		return
	}

	// The API only returns key material on Create, so preserve the value from state.
	if encryptionKey.Key == "" {
		encryptionKey.Key = stateData.Key.ValueString()
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, encryptionKey)
}

func (r *EncryptionKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	planData, hasError := r.GetPlanData(ctx, req.Plan, &resp.Diagnostics)
	if hasError {
		return
	}

	var stateData tfmodels.EncryptionKey
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	encryptionKeyUUID := planData.UUID.ValueString()
	body := map[string]any{
		"name":        planData.Name.ValueString(),
		"description": planData.Description.ValueString(),
	}

	encryptionKey, err := r.client.EncryptionKeys().Update(ctx, encryptionKeyUUID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Encryption Key", err.Error())
		return
	}

	// The API only returns key material on Create, so preserve the value from state.
	if encryptionKey.Key == "" {
		encryptionKey.Key = stateData.Key.ValueString()
	}

	r.SetStateData(ctx, &resp.State, &resp.Diagnostics, encryptionKey)
}

func (r *EncryptionKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	encryptionKeyUUID, hasError := r.GetUUIDFromState(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	if err := r.client.EncryptionKeys().Delete(ctx, encryptionKeyUUID); err != nil {
		resp.Diagnostics.AddError("Unable to delete Encryption Key", err.Error())
	}
}

func (r *EncryptionKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
