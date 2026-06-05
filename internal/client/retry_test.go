package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// TestDo_PostNotRetriedOnNetworkError pins the SEC-1 fix at the stx level: a
// POST whose connection drops mid-flight must NOT be retried (the write may
// already have committed server-side), so the server sees exactly one attempt.
// stx delegates the decision to clicore httpclient.ShouldRetryNetwork.
func TestDo_PostNotRetriedOnNetworkError(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("no hijacker")
		}
		conn, _, _ := hj.Hijack()
		_ = conn.Close()
	}))
	defer srv.Close()

	c := New("k", "eu", "", false, 0)
	c.baseURL = srv.URL
	if _, err := c.Do(context.Background(), http.MethodPost, "/x", nil, map[string]any{"a": 1}); err == nil {
		t.Fatal("want network error, got nil")
	}
	if n := atomic.LoadInt32(&attempts); n != 1 {
		t.Fatalf("POST must not be retried on network error: got %d attempts", n)
	}
}
