package tfmodels

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-artie/internal/artieclient"
)

type Table struct {
	UUID                 types.String `tfsdk:"uuid"`
	Name                 types.String `tfsdk:"name"`
	Schema               types.String `tfsdk:"schema"`
	EnableHistoryMode    types.Bool   `tfsdk:"enable_history_mode"`
	IndividualDeployment types.Bool   `tfsdk:"individual_deployment"`
	IsPartitioned        types.Bool   `tfsdk:"is_partitioned"`
}

func (t Table) ToAPIModel() artieclient.Table {
	tableUUID := t.UUID.ValueString()
	if tableUUID == "" {
		tableUUID = uuid.Nil.String()
	}

	return artieclient.Table{
		UUID:                 uuid.MustParse(tableUUID),
		Name:                 t.Name.ValueString(),
		Schema:               t.Schema.ValueString(),
		EnableHistoryMode:    t.EnableHistoryMode.ValueBool(),
		IndividualDeployment: t.IndividualDeployment.ValueBool(),
		IsPartitioned:        t.IsPartitioned.ValueBool(),
	}
}
