package client

import (
	"context"
	"testing"
)

func TestGetAnalyticsInfo(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.GetAnalyticsInfo(context.Background(), "abc")
	rs.assertRequest(t, "GET", "/api/v2/containers/abc/analytics/info")
	if rs.rawQuery != "" {
		t.Errorf("expected no query, got %q", rs.rawQuery)
	}
}

func TestGetAnalyticsBrowsers(t *testing.T) {
	rs := newRecordingServer(t, 200, `[]`)
	defer rs.close()
	c := rs.client("k", "")
	c.GetAnalyticsBrowsers(context.Background(), "abc", AnalyticsQuery{Start: 1700000000, End: 1700003600})
	rs.assertRequest(t, "GET", "/api/v2/containers/abc/analytics/browsers")
	v := rs.queryValues()
	if v.Get("start") != "1700000000" || v.Get("end") != "1700003600" {
		t.Errorf("got query %v", v)
	}
}

func TestGetAnalyticsClients_OmitsZero(t *testing.T) {
	rs := newRecordingServer(t, 200, `[]`)
	defer rs.close()
	c := rs.client("k", "")
	c.GetAnalyticsClients(context.Background(), "abc", AnalyticsQuery{}) // zero values
	rs.assertRequest(t, "GET", "/api/v2/containers/abc/analytics/clients")
	if rs.rawQuery != "" {
		t.Errorf("zero AnalyticsQuery should produce no query; got %q", rs.rawQuery)
	}
}

func TestEnableAnalytics(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.EnableAnalytics(context.Background(), "abc", true)
	rs.assertRequest(t, "PATCH", "/api/v2/containers/abc/analytics-enable")
	got := rs.bodyJSON()
	if got["enabled"] != true {
		t.Errorf("body: %v", got)
	}
}
