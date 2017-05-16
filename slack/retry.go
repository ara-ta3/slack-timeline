package slack

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ara-ta3/retry"
	"github.com/pkg/errors"
)

type SlackRetryAble struct {
	N      int
	logger *log.Logger
}

func (retryAble *SlackRetryAble) request(httpFn func() (*http.Response, error)) (*http.Response, error) {
	res, err := retry.Retry(
		retryAble.N,
		func(n int, result interface{}) time.Duration {
			defaultSec := retry.ExponentialBackOff(n, result)
			r, ok := result.(*http.Response)
			if !ok {
				retryAble.logger.Printf("cannot cast response %+v. waiting %+v\n", result, defaultSec)
				return defaultSec
			}
			h := r.Header
			ts := h.Get("Retry-After")
			if ts == "" {
				retryAble.logger.Printf("cannot find time string from header %+v waiting %+v\n", r.Header, defaultSec)
				return defaultSec
			}
			t, e := strconv.Atoi(ts)
			if e != nil {
				retryAble.logger.Printf("cannot parse time string %+v waiting %+v\n", t, defaultSec)
				return defaultSec
			}

			sec := time.Duration(t+1) * time.Second
			retryAble.logger.Printf("Response was too many request. waiting %+v\n", sec)
			return sec
		},
		func() (interface{}, error) {
			r, err := httpFn()
			if err != nil {
				return r, err
			}
			if r.StatusCode == http.StatusTooManyRequests {
				return r, fmt.Errorf("Response code was too many requests.")
			}
			return r, nil

		},
	)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("%d times tried but failed.", retryAble.N))
	}
	response, ok := res.(*http.Response)
	if !ok {
		return nil, fmt.Errorf("Failed to cast to http response. result: %+v", res)
	}
	return response, nil

}

func (r *SlackRetryAble) GetRequest(url string) (*http.Response, error) {
	return r.request(func() (*http.Response, error) {
		return http.Get(url)
	})
}

func (r *SlackRetryAble) PostReqest(url string, params url.Values) (*http.Response, error) {
	return r.request(func() (*http.Response, error) {
		return http.PostForm(url, params)
	})
}
