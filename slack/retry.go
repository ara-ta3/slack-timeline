package slack

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type SlackRetryAble struct {
	N        int
	HttpFunc func() (*http.Response, error)
}

var exponentialBackOff = func(n int) time.Duration {
	return time.Duration(n) * time.Duration(n) * time.Second
}

func (r *SlackRetryAble) Request() (*http.Response, error) {
	res, err := Retry(
		r.N,
		func(n int, result interface{}) time.Duration {
			r, ok := result.(*http.Response)
			if !ok {
				return exponentialBackOff(n)
			}
			h := r.Header
			ts := h.Get("Retry-After")
			if ts == "" {
				return exponentialBackOff(n)
			}
			t, e := strconv.Atoi(ts)
			if e != nil {
				return exponentialBackOff(n)
			}

			return time.Duration(t) * time.Second
		},
		func() (interface{}, error) {
			return r.HttpFunc()
		},
		func(res interface{}) bool {
			r, ok := res.(*http.Response)
			if !ok {
				return true
			}
			return r.StatusCode == http.StatusTooManyRequests
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("%d times tried but failed.", r.N))
	}
	response, ok := res.(*http.Response)
	if !ok {
		return nil, fmt.Errorf("Failed to cast to http response. result: %+v", res)
	}
	return response, nil
}

func Retry(
	n int,
	interval func(n int, result interface{}) time.Duration,
	fn func() (interface{}, error),
	shouldRetry func(res interface{}) bool,
) (interface{}, error) {
	return loop(1, n, interval, fn, shouldRetry)
}

func loop(
	i, n int,
	interval func(n int, result interface{}) time.Duration,
	fn func() (interface{}, error),
	shouldRetry func(res interface{}) bool,
) (interface{}, error) {
	res, err := fn()
	if err == nil && shouldRetry(res) {
		err = fmt.Errorf("result should be retry.")
	}

	if i >= n {
		return res, err
	}

	if err != nil {
		time.Sleep(interval(i, res))
		return loop(i+1, n, interval, fn, shouldRetry)
	}

	return res, nil
}
