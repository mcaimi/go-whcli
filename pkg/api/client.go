package api

import (
	"net/http"
	"time"
)

const (
	defaultApiTimeout       = 30 * time.Second
	defaultDownloadTimeout  = 10 * time.Minute
)

var downloadHttpClient = &http.Client{
	Timeout: defaultDownloadTimeout,
}

var httpClient = &http.Client{
	Timeout: defaultApiTimeout,
}

func DoGet(url string) (*http.Response, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		return nil, errStatusCode(resp.StatusCode)
	}
	return resp, nil
}

func DoGetDownload(url string) (*http.Response, error) {
	resp, err := downloadHttpClient.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		return nil, errStatusCode(resp.StatusCode)
	}
	return resp, nil
}

func SetApiTimeout(timeout time.Duration) {
	httpClient.Timeout = timeout
}

func SetDownloadTimeout(timeout time.Duration) {
	downloadHttpClient.Timeout = timeout
}

type errStatusCode int

func (e errStatusCode) Error() string {
	return "HTTP error: status code %d"
}
