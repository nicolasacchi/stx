package client

import (
	"context"
	"testing"
)

func TestListDomains(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":{"items":[]}}`)
	defer rs.close()
	c := rs.client("k", "")
	c.ListDomains(context.Background(), "abc")
	rs.assertRequest(t, "GET", "/api/v2/containers/abc/domains")
}

func TestGetDomain(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.GetDomain(context.Background(), "abc", "uuid-123")
	rs.assertRequest(t, "GET", "/api/v2/containers/abc/domains/uuid-123")
}

func TestCreateDomain(t *testing.T) {
	rs := newRecordingServer(t, 201, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.CreateDomain(context.Background(), "abc", map[string]string{"name": "x.com", "cdnType": "stape"})
	rs.assertRequest(t, "POST", "/api/v2/containers/abc/domains")
	got := rs.bodyJSON()
	if got["name"] != "x.com" || got["cdnType"] != "stape" {
		t.Errorf("body: %v", got)
	}
}

func TestUpdateDomain(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.UpdateDomain(context.Background(), "abc", "uuid", map[string]string{"name": "y.com", "cdnType": "stape"})
	rs.assertRequest(t, "PUT", "/api/v2/containers/abc/domains/uuid")
}

func TestDeleteDomain(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.DeleteDomain(context.Background(), "abc", "uuid")
	rs.assertRequest(t, "DELETE", "/api/v2/containers/abc/domains/uuid")
}

func TestValidateDomain(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.ValidateDomain(context.Background(), "abc", map[string]string{"name": "x.com", "cdnType": "stape"})
	rs.assertRequest(t, "POST", "/api/v2/containers/abc/domains/validate")
}

func TestRevalidateDomain(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.RevalidateDomain(context.Background(), "abc", "uuid")
	rs.assertRequest(t, "POST", "/api/v2/containers/abc/domains/uuid/revalidate")
	if len(rs.body) != 0 {
		t.Errorf("revalidate should send no body; got %q", rs.body)
	}
}
