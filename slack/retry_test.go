package slack

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry10Times(t *testing.T) {
	c := 0
	_, err := Retry(
		10,
		func(n int, r interface{}) time.Duration {
			return 0 * time.Second
		},
		func() (interface{}, error) {
			c++
			return nil, fmt.Errorf("dummy")
		},
		func(res interface{}) bool {
			return false
		},
	)
	assert.Error(t, err)
	assert.Equal(t, c, 10)
}

func TestRetry10TimesWhenShouldRetry(t *testing.T) {
	c := 0
	_, err := Retry(
		10,
		func(n int, r interface{}) time.Duration {
			return 0 * time.Second
		},
		func() (interface{}, error) {
			c++
			return nil, nil
		},
		func(res interface{}) bool {
			return true
		},
	)
	assert.Error(t, err)
	assert.Equal(t, c, 10)
}

func TestNotRetryWhenShouldNotRetry(t *testing.T) {
	c := 0
	_, err := Retry(
		10,
		func(n int, r interface{}) time.Duration {
			return 0 * time.Second
		},
		func() (interface{}, error) {
			c++
			return nil, nil
		},
		func(res interface{}) bool {
			return false
		},
	)
	if assert.NoError(t, err) {
		assert.Equal(t, c, 1)
	}
}
