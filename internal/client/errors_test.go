package client

import (
	"strings"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	cases := []struct {
		name string
		err  *APIError
		want string
	}{
		{
			"with body",
			&APIError{StatusCode: 401, Method: "GET", URL: "https://x/y", Body: `{"error":"Unauthorized"}`},
			`GET https://x/y: 401 {"error":"Unauthorized"}`,
		},
		{
			"empty body",
			&APIError{StatusCode: 500, Method: "POST", URL: "https://x/y"},
			`POST https://x/y: 500`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAPIError_ExitCode(t *testing.T) {
	cases := map[int]int{
		200: 1, // unexpected (not used in practice)
		400: 3,
		401: 2,
		403: 2,
		404: 4,
		429: 5,
		500: 1,
		502: 1,
		503: 1,
	}
	for status, want := range cases {
		err := &APIError{StatusCode: status}
		if got := err.ExitCode(); got != want {
			t.Errorf("status %d: got exit code %d, want %d", status, got, want)
		}
	}
}

func TestAPIError_AsErrorString(t *testing.T) {
	// Sanity: ensure the body field appears verbatim in Error()
	err := &APIError{StatusCode: 400, Method: "POST", URL: "https://api/x", Body: `{"validation":"fail"}`}
	if !strings.Contains(err.Error(), `"validation":"fail"`) {
		t.Fatalf("expected body in error: %s", err.Error())
	}
}
