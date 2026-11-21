package provider

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipLargeBody(t *testing.T) {
	newRequest := func(body string) *http.Request {
		req, err := http.NewRequest(http.MethodPost, "https://api.artie.com/pipelines/x", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		return req
	}

	{
		// At or under the threshold, the body is sent as is.
		body := strings.Repeat("x", gzipThresholdBytes)
		req := newRequest(body)
		require.NoError(t, gzipLargeBody(req))
		assert.Empty(t, req.Header.Get("Content-Encoding"))
		sent, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Equal(t, body, string(sent))
	}
	{
		// Past it, the body is gzipped and still replayable for retries.
		body := `{"a":"` + strings.Repeat("x", gzipThresholdBytes) + `"}`
		req := newRequest(body)
		require.NoError(t, gzipLargeBody(req))
		assert.Equal(t, "gzip", req.Header.Get("Content-Encoding"))
		assert.Less(t, req.ContentLength, int64(gzipThresholdBytes))

		for range 2 {
			readCloser, err := req.GetBody()
			require.NoError(t, err)
			gzipReader, err := gzip.NewReader(readCloser)
			require.NoError(t, err)
			decompressed, err := io.ReadAll(gzipReader)
			require.NoError(t, err)
			assert.Equal(t, body, string(decompressed))
		}
	}
	{
		// GET requests have no body.
		req, err := http.NewRequest(http.MethodGet, "https://api.artie.com/pipelines/x", nil)
		require.NoError(t, err)
		require.NoError(t, gzipLargeBody(req))
		assert.Empty(t, req.Header.Get("Content-Encoding"))
	}
}
