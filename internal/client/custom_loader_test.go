package client

import (
	"context"
	"testing"
)

// Critical: the path uses singular `container` (no `s`) — the only Stape
// endpoint that does so. Regression-test it.
func TestGenerateCustomLoader_SingularPath(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":{"jsCode":"<script>...</script>"}}`)
	defer rs.close()
	c := rs.client("k", "")
	c.GenerateCustomLoader(context.Background(), "abc", map[string]string{"webGtmId": "GTM-X"})
	rs.assertRequest(t, "POST", "/api/v2/container/abc/custom-loader")
	if rs.path == "/api/v2/containers/abc/custom-loader" {
		t.Fatal("regression: path uses plural; spec uses singular /container/")
	}
	got := rs.bodyJSON()
	if got["webGtmId"] != "GTM-X" {
		t.Errorf("body: %v", got)
	}
}
