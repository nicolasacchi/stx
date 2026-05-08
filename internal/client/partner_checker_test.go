package client

import (
	"context"
	"testing"
)

func TestCreatePartnerChecker(t *testing.T) {
	rs := newRecordingServer(t, 201, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.CreatePartnerChecker(context.Background(), "https://site.example", "https://callback.example")
	rs.assertRequest(t, "POST", "/api/v2/partner-tracking-checker")
	got := rs.bodyJSON()
	if got["siteUrl"] != "https://site.example" {
		t.Errorf("siteUrl: %v", got)
	}
	if got["callbackUrl"] != "https://callback.example" {
		t.Errorf("callbackUrl: %v", got)
	}
}

func TestGetPartnerCheckerLimit(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":{"remaining":42}}`)
	defer rs.close()
	c := rs.client("k", "")
	c.GetPartnerCheckerLimit(context.Background())
	rs.assertRequest(t, "GET", "/api/v2/partner-tracking-checker/limit")
}
