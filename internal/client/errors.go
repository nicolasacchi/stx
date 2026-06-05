package client

import (
	"fmt"

	"github.com/nicolasacchi/clicore/cierrors"
)

type APIError struct {
	StatusCode int
	Method     string
	URL        string
	Body       string
}

func (e *APIError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("%s %s: %d %s", e.Method, e.URL, e.StatusCode, e.Body)
	}
	return fmt.Sprintf("%s %s: %d", e.Method, e.URL, e.StatusCode)
}

// ExitCode delegates to the fleet-canonical table (auth=2, validation=3,
// not_found=4, rate_limited=5, else 1).
func (e *APIError) ExitCode() int {
	return cierrors.ExitCodeFor(e.StatusCode, "")
}
