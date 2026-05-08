package client

import (
	"context"
	"strings"
	"testing"
)

func TestListContainers(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":{"items":[]}}`)
	defer rs.close()
	c := rs.client("k", "")

	if _, err := c.ListContainers(context.Background(), 0); err != nil {
		t.Fatal(err)
	}
	rs.assertRequest(t, "GET", "/api/v2/containers")
	if rs.rawQuery != "" {
		t.Errorf("expected empty query for limit=0; got %q", rs.rawQuery)
	}

	if _, err := c.ListContainers(context.Background(), 50); err != nil {
		t.Fatal(err)
	}
	if got := rs.queryValues().Get("limit"); got != "50" {
		t.Errorf("got limit=%q, want 50", got)
	}
}

func TestGetContainer(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":{"identifier":"abc"}}`)
	defer rs.close()
	c := rs.client("k", "")
	if _, err := c.GetContainer(context.Background(), "abc"); err != nil {
		t.Fatal(err)
	}
	rs.assertRequest(t, "GET", "/api/v2/containers/abc")
}

func TestGetContainer_PathEscape(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.GetContainer(context.Background(), "id with space")
	// raw URL preserves escaping; decoded path normalises spaces
	if !strings.Contains(rs.rawPath, "id%20with%20space") {
		t.Errorf("path not escaped: %q (raw=%q)", rs.path, rs.rawPath)
	}
}

func TestCreateContainer(t *testing.T) {
	rs := newRecordingServer(t, 201, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	body := map[string]any{"name": "foo", "code": "x", "zone": "eut"}
	c.CreateContainer(context.Background(), body)
	rs.assertRequest(t, "POST", "/api/v2/containers")
	got := rs.bodyJSON()
	if got["name"] != "foo" || got["zone"] != "eut" {
		t.Errorf("body: %v", got)
	}
}

func TestUpdateContainer(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.UpdateContainer(context.Background(), "abc", map[string]string{"name": "new"})
	rs.assertRequest(t, "PUT", "/api/v2/containers/abc")
}

func TestDeleteContainer_NoBody(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.DeleteContainer(context.Background(), "abc", nil)
	rs.assertRequest(t, "DELETE", "/api/v2/containers/abc")
	if len(rs.body) != 0 {
		t.Errorf("expected empty body, got %q", rs.body)
	}
}

func TestDeleteContainer_WithReason(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.DeleteContainer(context.Background(), "abc", SubscriptionCancelForm{
		CancelReason: &SubscriptionCancelReason{
			Setup:  []OptionForm{{Type: "ga"}},
			Cancel: []OptionForm{{Type: "use"}},
		},
	})
	rs.assertRequest(t, "DELETE", "/api/v2/containers/abc")
	got := rs.bodyJSON()
	cancelReason, _ := got["cancelReason"].(map[string]any)
	if cancelReason == nil {
		t.Fatalf("body missing cancelReason: %v", got)
	}
	setup, _ := cancelReason["setup"].([]any)
	if len(setup) != 1 {
		t.Errorf("expected 1 setup item, got %d (%v)", len(setup), setup)
	}
}

func TestTransferContainer(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.TransferContainer(context.Background(), "abc", "new@example.com")
	rs.assertRequest(t, "PUT", "/api/v2/containers/abc/transfer")
	got := rs.bodyJSON()
	if got["email"] != "new@example.com" {
		t.Errorf("body: %v", got)
	}
}
