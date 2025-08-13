package utils

import (
	"fmt"
	"net/http"
	"time"
)

var delays = []int{
	1,
	3,
	5,
}

func WithRetry(f func() error) error {
	var err error

	for i := 0; i <= len(delays); i++ {
		if err = f(); err == nil {
			return nil
		}

		if i == len(delays) {
			break
		}
		time.Sleep(time.Duration(delays[i]) * time.Second)
	}

	return err
}

type HTTPClientWRetry struct {
	client *http.Client
}

var DefaultClient = HTTPClientWRetry{
	client: http.DefaultClient,
}

func (c *HTTPClientWRetry) Do(req *http.Request) (*http.Response, error) {
	for i := 0; i <= len(delays); i++ {
		if resp, err := c.client.Do(req); err == nil {
			return resp, nil
		}

		if i == len(delays) {
			break
		}

		time.Sleep(time.Duration(delays[i]) * time.Second)
	}

	return nil, fmt.Errorf("failed to make request after %d atempts", len(delays))
}
