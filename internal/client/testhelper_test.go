package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// recordingServer captures the request received and returns a canned response.
type recordingServer struct {
	t        *testing.T
	server   *httptest.Server
	method   string
	path     string // decoded
	rawPath  string // raw (preserves URL escaping)
	rawQuery string
	headers  http.Header
	body     []byte
	// Optional: control the response per-request
	respStatus int
	respBody   string
	// Optional: queue of responses for retry tests
	respQueue []responseSpec
}

type responseSpec struct {
	status int
	body   string
	header http.Header
}

func newRecordingServer(t *testing.T, status int, body string) *recordingServer {
	t.Helper()
	rs := &recordingServer{t: t, respStatus: status, respBody: body}
	rs.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rs.method = r.Method
		rs.path = r.URL.Path
		// RequestURI is the raw request line — preserves URL escaping;
		// fall back via URL.RawPath, then decoded Path.
		if r.RequestURI != "" {
			rs.rawPath = r.RequestURI
			if i := strings.Index(rs.rawPath, "?"); i >= 0 {
				rs.rawPath = rs.rawPath[:i]
			}
		} else if r.URL.RawPath != "" {
			rs.rawPath = r.URL.RawPath
		} else {
			rs.rawPath = r.URL.Path
		}
		rs.rawQuery = r.URL.RawQuery
		rs.headers = r.Header.Clone()
		rs.body, _ = io.ReadAll(r.Body)

		if len(rs.respQueue) > 0 {
			next := rs.respQueue[0]
			rs.respQueue = rs.respQueue[1:]
			for k, v := range next.header {
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}
			w.WriteHeader(next.status)
			io.WriteString(w, next.body)
			return
		}
		w.WriteHeader(rs.respStatus)
		io.WriteString(w, rs.respBody)
	}))
	return rs
}

func (rs *recordingServer) close() { rs.server.Close() }

// client creates a Client pointing at the test server.
func (rs *recordingServer) client(apiKey, workspace string) *Client {
	c := New(apiKey, "eu", workspace, false, 0)
	c.baseURL = rs.server.URL
	return c
}

// queryValues returns parsed query params from the last request.
func (rs *recordingServer) queryValues() url.Values {
	v, _ := url.ParseQuery(rs.rawQuery)
	return v
}

// bodyJSON unmarshals the recorded request body into a generic map.
func (rs *recordingServer) bodyJSON() map[string]any {
	if len(rs.body) == 0 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(rs.body, &out); err != nil {
		rs.t.Fatalf("body not JSON: %v: %q", err, rs.body)
	}
	return out
}

// assertRequest checks method + path equal expected values.
func (rs *recordingServer) assertRequest(t *testing.T, method, path string) {
	t.Helper()
	if rs.method != method {
		t.Errorf("method: got %q, want %q", rs.method, method)
	}
	if rs.path != path {
		t.Errorf("path: got %q, want %q", rs.path, path)
	}
}

// assertHeader confirms a header equals expected value.
func (rs *recordingServer) assertHeader(t *testing.T, name, want string) {
	t.Helper()
	if got := rs.headers.Get(name); got != want {
		t.Errorf("header %s: got %q, want %q", name, got, want)
	}
}

// containsHeader returns true if the recorded request had this header set (any value).
func (rs *recordingServer) hasHeader(name string) bool {
	return rs.headers.Get(name) != ""
}

// dump returns a debug string for failure messages.
func (rs *recordingServer) dump() string {
	return strings.Join([]string{
		"method=" + rs.method,
		"path=" + rs.path,
		"query=" + rs.rawQuery,
		"body=" + string(rs.body),
	}, " ")
}
