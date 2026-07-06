package api

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	defaultApiTimeout      = 30 * time.Second
	defaultDownloadTimeout = 10 * time.Minute
)

var (
	downloadHttpClient *http.Client = &http.Client{
		Timeout: defaultDownloadTimeout,
	}
	httpClient *http.Client = &http.Client{
		Timeout: defaultApiTimeout,
	}
	APIKey string = ""
)

// init loads API key from environment variable
func init() {
	if key := apiKeyFromEnv(); key != "" {
		APIKey = key
	}
}

func apiKeyFromEnv() string {
	return os.Getenv("WALLHAVEN_API_KEY")
}

// DoGet executes a GET request using the shared API HTTP client,
// injecting the API key header if present.
func DoGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if APIKey != "" {
		req.Header.Set("Accept-ApikeyV2", APIKey)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return resp, nil
}

// DoGetDownload executes a GET request using the download HTTP client (no API key needed).
func DoGetDownload(url string) (*http.Response, error) {
	resp, err := downloadHttpClient.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return resp, nil
}

func SetApiTimeout(timeout time.Duration) {
	httpClient.Timeout = timeout
}

func SetDownloadTimeout(timeout time.Duration) {
	downloadHttpClient.Timeout = timeout
}

func SetApiKey(key string) {
	APIKey = key
}

type errStatusCode int

func (e errStatusCode) Error() string {
	return fmt.Sprintf("HTTP %d", int(e))
}
