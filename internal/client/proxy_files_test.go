package client

import (
	"context"
	"testing"
)

func TestListProxyFiles(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":[]}`)
	defer rs.close()
	c := rs.client("k", "")
	c.ListProxyFiles(context.Background(), "abc")
	rs.assertRequest(t, "GET", "/api/v2/containers/abc/proxy-files")
}

func TestUpdateProxyFiles(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	body := map[string]any{"containerProxyFiles": []map[string]any{{
		"originalFileUrl": "https://x", "customPath": "/y", "cacheMaxAge": 3600,
	}}}
	c.UpdateProxyFiles(context.Background(), "abc", body)
	rs.assertRequest(t, "PUT", "/api/v2/containers/abc/proxy-files")
}
