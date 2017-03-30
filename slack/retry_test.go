package slack

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry10Times(t *testing.T) {
	c := 0
	_, err := Retry(
		10,
		func(n int) time.Duration {
			return 0 * time.Second
		},
		func() (interface{}, error) {
			c++
			return nil, fmt.Errorf("dummy")
		},
	)
	assert.Error(t, err)
	assert.Equal(t, c, 10)
}

func TestNotRetryWhenSuccessedHTTPResponse(t *testing.T) {
	r, err := Retry(
		10,
		func(n int) time.Duration {
			return 0 * time.Second
		},
		func() (interface{}, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
			}, nil
		},
	)
	if assert.NoError(t, err) {
		assert.Equal(t, r.StatusCode, http.StatusOK)
	}
}
