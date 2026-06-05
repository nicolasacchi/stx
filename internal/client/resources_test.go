package client

import (
	"context"
	"testing"
)

func TestResourceKinds_HasAll15(t *testing.T) {
	if got := len(ResourceKinds); got != 15 {
		t.Errorf("expected 15 resource kinds, got %d", got)
	}
}

func TestIsValidResourceKind(t *testing.T) {
	cases := map[string]bool{
		"container-zones":            true,
		"container-statuses":         true,
		"container-domain-cdn-types": true,
		"NOT-A-KIND":                 false,
		"":                           false,
		"container-zones-typo":       false,
	}
	for kind, want := range cases {
		if got := IsValidResourceKind(kind); got != want {
			t.Errorf("IsValidResourceKind(%q): got %v, want %v", kind, got, want)
		}
	}
}

func TestGetResource_Valid(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":[]}`)
	defer rs.close()
	c := rs.client("k", "")
	if _, err := c.GetResource(context.Background(), "container-zones"); err != nil {
		t.Fatal(err)
	}
	rs.assertRequest(t, "GET", "/api/v2/resources/container-zones")
}

func TestGetResource_InvalidKind(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	_, err := c.GetResource(context.Background(), "BOGUS")
	if err == nil {
		t.Fatal("expected error for invalid kind")
	}
	if rs.method != "" {
		t.Errorf("HTTP request should not have been sent for invalid kind; got %s %s", rs.method, rs.path)
	}
}
