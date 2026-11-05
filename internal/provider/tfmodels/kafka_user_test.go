package tfmodels

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/openapi"
)

func TestKafkaUserFromAPIModel(t *testing.T) {
	apiUser := openapi.PayloadsKafkaUser{
		Uuid:                uuid.New(),
		Username:            "my-app",
		AccessProfile:       KafkaUserAccessProfileProducerConsumer,
		ConsumerGroupPrefix: "acme.",
		Status:              "active",
	}
	user := KafkaUserFromAPIModel("cluster-uuid", apiUser, types.StringValue("secret"))
	assert.Equal(t, apiUser.Uuid.String(), user.UUID.ValueString())
	assert.Equal(t, "cluster-uuid", user.ClusterUUID.ValueString())
	assert.Equal(t, "my-app", user.Username.ValueString())
	assert.Equal(t, "acme.", user.ConsumerGroupPrefix.ValueString())
	assert.Equal(t, "active", user.Status.ValueString())
	assert.Equal(t, "secret", user.Password.ValueString())
	assert.False(t, user.IsAdmin())
}

func TestKafkaUserToAPICreateRequest(t *testing.T) {
	user := KafkaUser{Username: types.StringValue("my-app"), AccessProfile: types.StringValue(KafkaUserAccessProfileAdmin)}
	assert.True(t, user.IsAdmin())
	// The /users endpoint only takes producer_consumer; admin users go through ToAPICreateAdminRequest.
	assert.Equal(t, openapi.RouterCreateKafkaUserRequest{AccessProfile: "producer_consumer", Username: "my-app"}, user.ToAPICreateRequest())
	assert.Equal(t, openapi.RouterCreateKafkaAdminUserRequest{Username: "my-app"}, user.ToAPICreateAdminRequest())
}
