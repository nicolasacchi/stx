// Package client wraps the Stape API.
//
// Auth: X-AUTH-TOKEN header (Stape's only securityScheme).
// Optional X-WORKSPACE header for multi-workspace accounts (56/83 endpoints reference it).
// Region routing: eu -> https://api.app.eu.stape.io, global -> https://api.app.stape.io.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	DefaultTimeout = 30 * time.Second
	MaxRetries     = 3
)

type Client struct {
	http      *http.Client
	apiKey    string
	workspace string
	baseURL   string
	verbose   bool
}

// New constructs a Client for the given region.
// region must be "eu" (default) or "global". workspace may be empty.
func New(apiKey, region, workspace string, verbose bool) *Client {
	base := "https://api.app.eu.stape.io"
	if region == "global" {
		base = "https://api.app.stape.io"
	}
	return &Client{
		http:      &http.Client{Timeout: DefaultTimeout},
		apiKey:    apiKey,
		workspace: workspace,
		baseURL:   base,
		verbose:   verbose,
	}
}

func (c *Client) BaseURL() string { return c.baseURL }

// Get performs a GET request. params may be nil.
func (c *Client) Get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	return c.Do(ctx, http.MethodGet, path, params, nil)
}

// Post performs a POST with a JSON body (body=nil sends empty body).
func (c *Client) Post(ctx context.Context, path string, body any) ([]byte, error) {
	return c.Do(ctx, http.MethodPost, path, nil, body)
}

func (c *Client) Put(ctx context.Context, path string, body any) ([]byte, error) {
	return c.Do(ctx, http.MethodPut, path, nil, body)
}

func (c *Client) Patch(ctx context.Context, path string, body any) ([]byte, error) {
	return c.Do(ctx, http.MethodPatch, path, nil, body)
}

func (c *Client) Delete(ctx context.Context, path string, body any) ([]byte, error) {
	return c.Do(ctx, http.MethodDelete, path, nil, body)
}

// Do is the workhorse. method/path required, query and body optional.
// Honors Retry-After on 429; otherwise exponential backoff 1s/2s/4s + jitter on 5xx/429.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any) ([]byte, error) {
	full := c.baseURL + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		if attempt > 0 {
			delay := backoffDelay(lastErr, attempt)
			if c.verbose {
				fmt.Fprintf(httpStderr(), "stx: retry %d/%d after %s\n", attempt, MaxRetries, delay)
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, full, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("X-AUTH-TOKEN", c.apiKey)
		req.Header.Set("Accept", "application/json")
		if c.workspace != "" {
			req.Header.Set("X-WORKSPACE", c.workspace)
		}
		if len(bodyBytes) > 0 {
			req.Header.Set("Content-Type", "application/json")
		}

		if c.verbose {
			fmt.Fprintf(httpStderr(), "stx: %s %s\n", method, full)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			if !shouldRetryNetwork(err) {
				return nil, err
			}
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if c.verbose {
			fmt.Fprintf(httpStderr(), "stx: -> %d (%d bytes)\n", resp.StatusCode, len(respBody))
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, nil
		}

		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			URL:        full,
			Body:       string(respBody),
		}
		if shouldRetryStatus(resp.StatusCode) && attempt < MaxRetries {
			lastErr = retryableErr{apiErr: apiErr, retryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}
			continue
		}
		return nil, apiErr
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

type retryableErr struct {
	apiErr     *APIError
	retryAfter time.Duration
}

func (r retryableErr) Error() string { return r.apiErr.Error() }

func backoffDelay(lastErr error, attempt int) time.Duration {
	if r, ok := lastErr.(retryableErr); ok && r.retryAfter > 0 {
		return r.retryAfter
	}
	base := time.Duration(1<<(attempt-1)) * time.Second // 1s, 2s, 4s
	jitter := time.Duration(rand.Int63n(int64(500 * time.Millisecond)))
	return base + jitter
}

func shouldRetryStatus(code int) bool {
	return code == 429 || (code >= 500 && code < 600)
}

func shouldRetryNetwork(err error) bool {
	// Network/connection errors are retryable
	return err != nil
}

func parseRetryAfter(s string) time.Duration {
	if s == "" {
		return 0
	}
	if secs, err := strconv.Atoi(s); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return 0
}

// httpStderr is a var so tests can swap it.
var httpStderr = func() io.Writer { return verboseDest }

var verboseDest io.Writer = io.Discard

// SetVerboseDest lets the command layer route verbose output (default: io.Discard).
func SetVerboseDest(w io.Writer) { verboseDest = w }
