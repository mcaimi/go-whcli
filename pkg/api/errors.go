package api

import (
	"errors"
	"fmt"
	"net/http"
)

// ApiErrorMessage turns an api package error into a user-friendly message,
// rephrasing HTTP status codes (e.g. 404 -> "not found"). Shared by the CLI
// and GUI front ends so error text stays consistent between them.
func ApiErrorMessage(what, id string, err error) string {
	var hErr *HTTPError
	if errors.As(err, &hErr) && hErr.Status == http.StatusNotFound && id != "" {
		return fmt.Sprintf("%s '%s' not found (HTTP 404)", what, id)
	}
	if id != "" {
		return fmt.Sprintf("failed to %s '%s': %s", what, id, err)
	}
	return fmt.Sprintf("failed to %s: %s", what, err)
}
