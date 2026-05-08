package client

import (
	"context"
	"testing"
)

func TestListAPIKeys(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":[]}`)
	defer rs.close()
	c := rs.client("k", "")
	c.ListAPIKeys(context.Background())
	rs.assertRequest(t, "GET", "/api/v2/users/api-key")
}

func TestCreateAPIKey(t *testing.T) {
	rs := newRecordingServer(t, 201, `{"body":{"token":"new"}}`)
	defer rs.close()
	c := rs.client("k", "")
	c.CreateAPIKey(context.Background(), "test-key")
	rs.assertRequest(t, "POST", "/api/v2/users/api-key")
	got := rs.bodyJSON()
	if got["name"] != "test-key" {
		t.Errorf("body: %v", got)
	}
}

func TestDeleteAPIKey(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.DeleteAPIKey(context.Background(), "key-id")
	rs.assertRequest(t, "DELETE", "/api/v2/users/api-key/key-id")
}
