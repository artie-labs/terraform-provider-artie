package provider

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitForCredentials(t *testing.T) {
	calls := 0
	password := "secret"
	creds, err := waitForCredentials(t.Context(), time.Second, time.Millisecond, func(context.Context) (*string, int, error) {
		calls++
		if calls < 3 {
			return nil, http.StatusConflict, errors.New("password isn't ready yet")
		}
		return &password, http.StatusOK, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "secret", *creds)
	assert.Equal(t, 3, calls)
}

func TestWaitForCredentials_OtherErrorStops(t *testing.T) {
	calls := 0
	_, err := waitForCredentials(t.Context(), time.Second, time.Millisecond, func(context.Context) (*string, int, error) {
		calls++
		return nil, http.StatusNotFound, errors.New("not found")
	})
	assert.ErrorContains(t, err, "not found")
	assert.Equal(t, 1, calls)
}

func TestWaitForCredentials_Timeout(t *testing.T) {
	_, err := waitForCredentials(t.Context(), 20*time.Millisecond, time.Millisecond, func(context.Context) (*string, int, error) {
		return nil, http.StatusConflict, errors.New("password isn't ready yet")
	})
	assert.ErrorContains(t, err, "timed out")
}

func TestWaitForCredentials_TimeoutCancelsBlockedFetch(t *testing.T) {
	_, err := waitForCredentials(t.Context(), 20*time.Millisecond, time.Millisecond, func(ctx context.Context) (*string, int, error) {
		<-ctx.Done()
		return nil, 0, ctx.Err()
	})
	assert.ErrorContains(t, err, "timed out")
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

type fakeResponse struct{ code int }

func (f fakeResponse) StatusCode() int { return f.code }

func TestStatusCode(t *testing.T) {
	assert.Equal(t, 0, statusCode[fakeResponse](nil))
	assert.Equal(t, http.StatusConflict, statusCode(&fakeResponse{code: http.StatusConflict}))
}
