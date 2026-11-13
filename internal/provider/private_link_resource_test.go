package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
		// Unknown values (e.g. from another resource) are checked later.
		diags := validatePrivateLinkConfig(tfmodels.PrivateLink{
			VpcServiceName:          types.StringNull(),
			AzIDs:                   types.ListNull(types.StringType),
			PrivateLinkServiceAlias: types.StringUnknown(),
		})
		assert.False(t, diags.HasError())
	}
}
