package tfmodels

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-artie/internal/openapi"
)

func TestPrivateLinkFromAPIModelCloudProvider(t *testing.T) {
	apiModel := openapi.PayloadsPrivateLinkConnection{
		Uuid:                uuid.New(),
		Name:                "my-privatelink",
		VpcServiceName:      "com.amazonaws.vpce.us-east-1.vpce-svc-1234567890abcdef0",
		AvailabilityZoneIds: []string{"use1-az1"},
		Region:              "us-east-1",
		Status:              "available",
		Provider:            openapi.EnumsCloudProviderAws,
	}

	tfModel, diags := PrivateLinkFromAPIModel(context.Background(), apiModel)
	require.False(t, diags.HasError())
	assert.Equal(t, "aws", tfModel.CloudProvider.ValueString())
	assert.Equal(t, "us-east-1", tfModel.Region.ValueString())
}

func TestPrivateLinkFromAPIModelTargets(t *testing.T) {
	{
		// AWS: the AWS target reads back as set, the empty alias as null.
		apiModel := openapi.PayloadsPrivateLinkConnection{
			Uuid:                uuid.New(),
			VpcServiceName:      "com.amazonaws.vpce.us-east-1.vpce-svc-1234567890abcdef0",
			AvailabilityZoneIds: []string{"use1-az1", "use1-az2"},
			Provider:            openapi.EnumsCloudProviderAws,
		}

		tfModel, diags := PrivateLinkFromAPIModel(t.Context(), apiModel)
		require.False(t, diags.HasError())
		assert.Equal(t, "com.amazonaws.vpce.us-east-1.vpce-svc-1234567890abcdef0", tfModel.VpcServiceName.ValueString())
		azIDs, diags := parseList[string](t.Context(), tfModel.AzIDs)
		require.False(t, diags.HasError())
		assert.Equal(t, []string{"use1-az1", "use1-az2"}, azIDs)
		assert.True(t, tfModel.PrivateLinkServiceAlias.IsNull())
	}
	{
		// Legacy AWS link with no AZ IDs: keeps reading back as [] so `az_ids = []` doesn't drift.
		apiModel := openapi.PayloadsPrivateLinkConnection{
			Uuid:                uuid.New(),
			VpcServiceName:      "com.amazonaws.vpce.us-east-1.vpce-svc-1234567890abcdef0",
			AvailabilityZoneIds: []string{},
			Provider:            openapi.EnumsCloudProviderAws,
		}

		tfModel, diags := PrivateLinkFromAPIModel(t.Context(), apiModel)
		require.False(t, diags.HasError())
		assert.False(t, tfModel.AzIDs.IsNull())
		assert.Empty(t, tfModel.AzIDs.Elements())
	}
	{
		// Azure: the empty AWS target reads back as null.
		apiModel := openapi.PayloadsPrivateLinkConnection{
			Uuid:                    uuid.New(),
			PrivateLinkServiceAlias: "my-pls.abc-123.eastus.azure.privatelinkservice",
			AvailabilityZoneIds:     []string{},
			Provider:                openapi.EnumsCloudProviderAzure,
		}

		tfModel, diags := PrivateLinkFromAPIModel(t.Context(), apiModel)
		require.False(t, diags.HasError())
		assert.Equal(t, "my-pls.abc-123.eastus.azure.privatelinkservice", tfModel.PrivateLinkServiceAlias.ValueString())
		assert.True(t, tfModel.VpcServiceName.IsNull())
		assert.True(t, tfModel.AzIDs.IsNull())
	}
}

func TestPrivateLinkToAPIRequestTargets(t *testing.T) {
	{
		// AWS: the alias is left out.
		azIDs, diags := types.ListValueFrom(t.Context(), types.StringType, []string{"use1-az1"})
		require.False(t, diags.HasError())
		tfModel := PrivateLink{
			Name:                    types.StringValue("aws"),
			VpcServiceName:          types.StringValue("com.amazonaws.vpce.us-east-1.vpce-svc-1234567890abcdef0"),
			AzIDs:                   azIDs,
			PrivateLinkServiceAlias: types.StringNull(),
		}

		create, diags := tfModel.ToAPICreateRequest(t.Context())
		require.False(t, diags.HasError())
		assert.Equal(t, "com.amazonaws.vpce.us-east-1.vpce-svc-1234567890abcdef0", *create.VpcServiceName)
		assert.Equal(t, []string{"use1-az1"}, *create.AvailabilityZoneIds)
		assert.Nil(t, create.PrivateLinkServiceAlias)

		update, diags := tfModel.ToAPIUpdateRequest(t.Context())
		require.False(t, diags.HasError())
		assert.Equal(t, []string{"use1-az1"}, *update.AvailabilityZoneIds)
		assert.Nil(t, update.PrivateLinkServiceAlias)
	}
	{
		// Azure: the AWS target is left out.
		tfModel := PrivateLink{
			Name:                    types.StringValue("azure"),
			VpcServiceName:          types.StringNull(),
			AzIDs:                   types.ListNull(types.StringType),
			PrivateLinkServiceAlias: types.StringValue("my-pls.abc-123.eastus.azure.privatelinkservice"),
		}

		create, diags := tfModel.ToAPICreateRequest(t.Context())
		require.False(t, diags.HasError())
		assert.Nil(t, create.VpcServiceName)
		assert.Nil(t, create.AvailabilityZoneIds)
		assert.Equal(t, "my-pls.abc-123.eastus.azure.privatelinkservice", *create.PrivateLinkServiceAlias)

		update, diags := tfModel.ToAPIUpdateRequest(t.Context())
		require.False(t, diags.HasError())
		assert.Nil(t, update.VpcServiceName)
		assert.Nil(t, update.AvailabilityZoneIds)
		assert.Equal(t, "my-pls.abc-123.eastus.azure.privatelinkservice", *update.PrivateLinkServiceAlias)
	}
}
