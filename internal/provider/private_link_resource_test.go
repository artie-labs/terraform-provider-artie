package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/provider/tfmodels"
)

func TestValidatePrivateLinkConfig(t *testing.T) {
	awsAZIDs := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("use1-az1")})
	{
		// AWS target.
		diags := validatePrivateLinkConfig(tfmodels.PrivateLink{
			VpcServiceName:          types.StringValue("com.amazonaws.vpce.us-east-1.vpce-svc-1234567890abcdef0"),
			AzIDs:                   awsAZIDs,
			PrivateLinkServiceAlias: types.StringNull(),
		})
		assert.False(t, diags.HasError())
	}
	{
		// Azure target.
		diags := validatePrivateLinkConfig(tfmodels.PrivateLink{
			VpcServiceName:          types.StringNull(),
			AzIDs:                   types.ListNull(types.StringType),
			PrivateLinkServiceAlias: types.StringValue("my-pls.abc-123.eastus.azure.privatelinkservice"),
		})
		assert.False(t, diags.HasError())
	}
	{
		// No target.
		diags := validatePrivateLinkConfig(tfmodels.PrivateLink{
			VpcServiceName:          types.StringNull(),
			AzIDs:                   types.ListNull(types.StringType),
			PrivateLinkServiceAlias: types.StringNull(),
		})
		assert.Equal(t, 2, diags.ErrorsCount())
	}
	{
		// Half the AWS target.
		diags := validatePrivateLinkConfig(tfmodels.PrivateLink{
			VpcServiceName:          types.StringValue("com.amazonaws.vpce.us-east-1.vpce-svc-1234567890abcdef0"),
			AzIDs:                   types.ListNull(types.StringType),
			PrivateLinkServiceAlias: types.StringNull(),
		})
		assert.Equal(t, 1, diags.ErrorsCount())
		assert.Contains(t, diags.Errors()[0].Detail(), "`az_ids` is required")
	}
	{
		// Both targets.
		diags := validatePrivateLinkConfig(tfmodels.PrivateLink{
			VpcServiceName:          types.StringValue("com.amazonaws.vpce.us-east-1.vpce-svc-1234567890abcdef0"),
			AzIDs:                   awsAZIDs,
			PrivateLinkServiceAlias: types.StringValue("my-pls.abc-123.eastus.azure.privatelinkservice"),
		})
		assert.Equal(t, 2, diags.ErrorsCount())
	}
	{
		// Alias with only az_ids.
		diags := validatePrivateLinkConfig(tfmodels.PrivateLink{
			VpcServiceName:          types.StringNull(),
			AzIDs:                   awsAZIDs,
			PrivateLinkServiceAlias: types.StringValue("my-pls.abc-123.eastus.azure.privatelinkservice"),
		})
		assert.Equal(t, 1, diags.ErrorsCount())
		assert.Contains(t, diags.Errors()[0].Detail(), "`az_ids` is for AWS")
	}
	{
		// Legacy AWS link written with no AZ IDs.
		diags := validatePrivateLinkConfig(tfmodels.PrivateLink{
			VpcServiceName:          types.StringValue("com.amazonaws.vpce.us-east-1.vpce-svc-1234567890abcdef0"),
			AzIDs:                   types.ListValueMust(types.StringType, []attr.Value{}),
			PrivateLinkServiceAlias: types.StringNull(),
		})
		assert.False(t, diags.HasError())
	}
	{
		// An empty alias counts as unset, so an AWS config doesn't get a conflict error (the attribute validator rejects "").
		diags := validatePrivateLinkConfig(tfmodels.PrivateLink{
			VpcServiceName:          types.StringValue("com.amazonaws.vpce.us-east-1.vpce-svc-1234567890abcdef0"),
			AzIDs:                   awsAZIDs,
			PrivateLinkServiceAlias: types.StringValue(""),
		})
		assert.False(t, diags.HasError())
	}
	{
		// Unknown values (e.g. from another resource) are checked later.
		diags := validatePrivateLinkConfig(tfmodels.PrivateLink{
			VpcServiceName:          types.StringNull(),
			AzIDs:                   types.ListNull(types.StringType),
			PrivateLinkServiceAlias: types.StringUnknown(),
		})
		assert.False(t, diags.HasError())
	}
}

func TestTrimmedNonEmptyString(t *testing.T) {
	validate := func(value string) bool {
		resp := &validator.StringResponse{}
		trimmedNonEmptyString.ValidateString(t.Context(), validator.StringRequest{Path: path.Root("x"), ConfigValue: types.StringValue(value)}, resp)
		return resp.Diagnostics.HasError()
	}
	{
		// Valid values.
		assert.False(t, validate("my-pls.abc-123.eastus.azure.privatelinkservice"))
		assert.False(t, validate("a"))
	}
	{
		// Empty.
		assert.True(t, validate(""))
	}
	{
		// Leading or trailing whitespace, which the API would trim.
		assert.True(t, validate(" my-pls"))
		assert.True(t, validate("my-pls\n"))
		assert.True(t, validate("my-pls "))
	}
}
