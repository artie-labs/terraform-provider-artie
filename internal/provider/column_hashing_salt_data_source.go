package provider

import (
	"context"
	"fmt"

	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/provider/tfmodels"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &ColumnHashingSaltDataSource{}
var _ datasource.DataSourceWithConfigure = &ColumnHashingSaltDataSource{}

func NewColumnHashingSaltDataSource() datasource.DataSource {
	return &ColumnHashingSaltDataSource{}
}

type ColumnHashingSaltDataSource struct {
	client artieclient.Client
}

func (d *ColumnHashingSaltDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_column_hashing_salt"
}

func (d *ColumnHashingSaltDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an existing Artie Column Hashing Salt by UUID. Use this to reference a salt that was created outside of Terraform (e.g. via the Artie UI) from an `artie_pipeline`'s `column_hashing_salt_uuid`.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The UUID of the column hashing salt to look up.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The human-readable name of the salt.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "A description of the salt's purpose.",
			},
			"salt": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The salt value. This value is sensitive and will not be displayed in plan output.",
			},
		},
	}
}

func (d *ColumnHashingSaltDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(ArtieProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected ArtieProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}

	client, err := providerData.NewClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to build Artie client", err.Error())
		return
	}

	d.client = client
}

func (d *ColumnHashingSaltDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var configData tfmodels.ColumnHashingSalt
	resp.Diagnostics.Append(req.Config.Get(ctx, &configData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	salt, err := d.client.ColumnHashingSalts().Get(ctx, configData.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Column Hashing Salt", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, tfmodels.ColumnHashingSaltFromAPIModel(salt))...)
}
