package slack

import (
	"net/http"
	"time"
)

func Retry(
	n int,
	interval func(n int) time.Duration,
	httpFn func() (*http.Response, error),
) (*http.Response, error) {
	return loop(1, n, interval, httpFn)
}

func loop(
	i, n int,
	interval func(n int) time.Duration,
	httpFn func() (*http.Response, error),
) (*http.Response, error) {
	res, err := httpFn()
	if i >= n {
		return res, err
	}

	if err != nil {
		time.Sleep(interval(i))
		return loop(i+1, n, interval, httpFn)
	}

	if res.StatusCode == http.StatusTooManyRequests {
		time.Sleep(interval(i))
		return loop(i+1, n, interval, httpFn)
	}
	return res, nil
}
