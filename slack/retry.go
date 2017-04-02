package slack

import (
	"fmt"
	"time"
)

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
