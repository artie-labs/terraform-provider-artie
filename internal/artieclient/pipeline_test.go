package artieclient

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T { return &v }

func TestAdvancedSettingsMaxConcurrentSnapshotsJSON(t *testing.T) {
	{
		body, err := json.Marshal(AdvancedSettings{})
		require.NoError(t, err)
		assert.NotContains(t, string(body), "maxConcurrentSnapshots")
	}

	{
		body, err := json.Marshal(AdvancedSettings{
			MaxConcurrentSnapshots: ptr[int64](6),
		})
		require.NoError(t, err)
		assert.Contains(t, string(body), `"maxConcurrentSnapshots":6`)
	}
}
