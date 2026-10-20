package api

import (
	"encoding/json"
	"fmt"
	"io"
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
		resp.Body.Close()
		return nil, httpError(resp)
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
		resp.Body.Close()
		return nil, httpError(resp)
	}
	return resp, nil
}

// httpError builds a descriptive error for a non-200 response.
// 404 is returned as *HTTPError so callers can rephrase it as "not found".
func httpError(resp *http.Response) error {
	if resp.StatusCode == http.StatusNotFound {
		return &HTTPError{Status: http.StatusNotFound}
	}
	return fmt.Errorf("HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
}

// HTTPError represents a non-2xx HTTP response from the WallHaven API.
type HTTPError struct {
	Status int
}

func (e *HTTPError) Error() string {
	if e.Status == http.StatusNotFound {
		return "HTTP 404 (not found)"
	}
	return fmt.Sprintf("HTTP %d %s", e.Status, http.StatusText(e.Status))
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

// doRequest executes DoGet, reads the body and unmarshals it into result.
func doRequest(url string, result interface{}) error {
	resp, err := DoGet(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read API response: %w", err)
	}

	if err = json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to decode API response: %w", err)
	}

	return nil
}
