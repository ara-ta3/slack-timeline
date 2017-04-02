package slack

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type SlackRetryAble struct {
	N      int
	logger *log.Logger
}

var exponentialBackOff = func(n int) time.Duration {
	return time.Duration(n) * time.Duration(n) * time.Second
}

func (retry *SlackRetryAble) request(httpFn func() (*http.Response, error)) (*http.Response, error) {
	res, err := Retry(
		retry.N,
		func(n int, result interface{}) time.Duration {
			defaultSec := exponentialBackOff(n)
			r, ok := result.(*http.Response)
			if !ok {
				retry.logger.Printf("cannot cast response %+v. waiting %+v\n", result, defaultSec)
				return defaultSec
			}
			h := r.Header
			ts := h.Get("Retry-After")
			if ts == "" {
				retry.logger.Printf("cannot find time string from header %+v waiting %+v\n", r.Header, defaultSec)
				return defaultSec
			}
			t, e := strconv.Atoi(ts)
			if e != nil {
				retry.logger.Printf("cannot parse time string %+v waiting %+v\n", t, defaultSec)
				return defaultSec
			}

			sec := time.Duration(t+1) * time.Second
			retry.logger.Printf("Response was too many request. waiting %+v\n", sec)
			return sec
		},
		func() (interface{}, error) {
			return httpFn()
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
		return nil, errors.Wrap(err, fmt.Sprintf("%d times tried but failed.", retry.N))
	}
	response, ok := res.(*http.Response)
	if !ok {
		return nil, fmt.Errorf("Failed to cast to http response. result: %+v", res)
	}
	return response, nil

}

func (r *SlackRetryAble) GetRequest(url string) (*http.Response, error) {
	r.logger.Printf("Get Request to %+v\n", url)
	return r.request(func() (*http.Response, error) {
		return http.Get(url)
	})
}

func (r *SlackRetryAble) PostReqest(url string, params url.Values) (*http.Response, error) {
	r.logger.Printf("Post Request to %+v\n", url)
	return r.request(func() (*http.Response, error) {
		return http.PostForm(url, params)
	})
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
