package client

import "fmt"

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

func (e *APIError) ExitCode() int {
	switch e.StatusCode {
	case 401, 403:
		return 2
	case 404:
		return 4
	default:
		return 1
	}
}
