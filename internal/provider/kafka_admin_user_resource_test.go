package provider

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/openapi"
)

func TestParseKafkaAdminUserImportID(t *testing.T) {
	clusterUUID, userUUID, ok := parseKafkaAdminUserImportID("cluster-1/user-1")
	assert.True(t, ok)
	assert.Equal(t, "cluster-1", clusterUUID)
	assert.Equal(t, "user-1", userUUID)

	for _, id := range []string{"", "user-1", "/user-1", "cluster-1/", "cluster-1/user-1/extra"} {
		_, _, ok := parseKafkaAdminUserImportID(id)
		assert.False(t, ok, id)
	}
}

func TestFindKafkaUser(t *testing.T) {
	first, second := uuid.New(), uuid.New()
	users := []openapi.PayloadsKafkaUser{{Uuid: first, Username: "first"}, {Uuid: second, Username: "second"}}

	assert.Equal(t, "second", findKafkaUser(users, second.String()).Username)
	assert.Nil(t, findKafkaUser(users, uuid.NewString()))
	assert.Nil(t, findKafkaUser(nil, first.String()))
}
