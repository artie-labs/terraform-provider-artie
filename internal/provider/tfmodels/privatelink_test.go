package tfmodels

import (
	"context"
	"testing"

	"github.com/google/uuid"
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
