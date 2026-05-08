package client

import (
	"context"
	"testing"
)

// --- logs ---

func TestListLogsAggregated(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":[]}`)
	defer rs.close()
	c := rs.client("k", "")
	c.ListLogsAggregated(context.Background(), "abc", LogsQuery{Start: 100, End: 200, Platform: "GA4", EventType: "PageView"})
	rs.assertRequest(t, "GET", "/api/v2/containers/abc/monitoring/logs/outgoing/aggregated")
	v := rs.queryValues()
	if v.Get("start") != "100" || v.Get("end") != "200" {
		t.Errorf("got query %v", v)
	}
	if v.Get("platform") != "GA4" || v.Get("eventType") != "PageView" {
		t.Errorf("got query %v", v)
	}
}

func TestListLogsDetailed_OmitsEmptyFilters(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":{"items":[]}}`)
	defer rs.close()
	c := rs.client("k", "")
	c.ListLogsDetailed(context.Background(), "abc", LogsQuery{Start: 100})
	v := rs.queryValues()
	if v.Get("start") != "100" {
		t.Errorf("missing start: %v", v)
	}
	if v.Get("end") != "" || v.Get("platform") != "" || v.Get("eventType") != "" {
		t.Errorf("zero values should be omitted: %v", v)
	}
}

func TestGetLogTrace(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":[]}`)
	defer rs.close()
	c := rs.client("k", "")
	c.GetLogTrace(context.Background(), "abc", "trace-uuid", 1700000000)
	rs.assertRequest(t, "GET", "/api/v2/containers/abc/monitoring/logs/outgoing/trace")
	v := rs.queryValues()
	if v.Get("traceId") != "trace-uuid" {
		t.Errorf("got traceId %q", v.Get("traceId"))
	}
	if v.Get("date") != "1700000000" {
		t.Errorf("got date %q", v.Get("date"))
	}
}

func TestGetLogTrace_OmitsZeroDate(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.GetLogTrace(context.Background(), "abc", "trace-uuid", 0)
	v := rs.queryValues()
	if v.Get("date") != "" {
		t.Errorf("zero date should be omitted; got %q", v.Get("date"))
	}
}

// --- rules ---

func TestListRules(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":{"items":[]}}`)
	defer rs.close()
	c := rs.client("k", "")
	c.ListRules(context.Background(), "abc")
	rs.assertRequest(t, "GET", "/api/v2/containers/abc/monitoring")
}

func TestGetRule(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.GetRule(context.Background(), "abc", "rule-id")
	rs.assertRequest(t, "GET", "/api/v2/containers/abc/monitoring/rule-id")
}

func TestCreateRule(t *testing.T) {
	rs := newRecordingServer(t, 201, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.CreateRule(context.Background(), "abc", map[string]string{"name": "alert"})
	rs.assertRequest(t, "POST", "/api/v2/containers/abc/monitoring")
}

func TestUpdateRule(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.UpdateRule(context.Background(), "abc", "rid", map[string]string{"name": "alert"})
	rs.assertRequest(t, "PUT", "/api/v2/containers/abc/monitoring/rid")
}

func TestDeleteRule(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.DeleteRule(context.Background(), "abc", "rid")
	rs.assertRequest(t, "DELETE", "/api/v2/containers/abc/monitoring/rid")
}

func TestSwitchRule(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.SwitchRule(context.Background(), "abc", "rid", true)
	rs.assertRequest(t, "PATCH", "/api/v2/containers/abc/monitoring/rid")
	got := rs.bodyJSON()
	if got["enabled"] != true {
		t.Errorf("body: %v", got)
	}
}

func TestResolveRule(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.ResolveRule(context.Background(), "abc", "rid")
	rs.assertRequest(t, "POST", "/api/v2/containers/abc/monitoring/rid/resolve")
	if len(rs.body) != 0 {
		t.Errorf("expected empty body, got %q", rs.body)
	}
}

// --- emails ---

func TestListEmails(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":{"items":[]}}`)
	defer rs.close()
	c := rs.client("k", "")
	c.ListEmails(context.Background(), "abc")
	rs.assertRequest(t, "GET", "/api/v2/containers/abc/monitoring/emails")
}

func TestAddEmail(t *testing.T) {
	rs := newRecordingServer(t, 201, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.AddEmail(context.Background(), "abc", "x@y.z")
	rs.assertRequest(t, "POST", "/api/v2/containers/abc/monitoring/emails")
	got := rs.bodyJSON()
	if got["email"] != "x@y.z" {
		t.Errorf("body: %v", got)
	}
}

func TestDeleteEmail_PathEscape(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.DeleteEmail(context.Background(), "abc", "x+tag@y.z")
	rs.assertRequest(t, "DELETE", "/api/v2/containers/abc/monitoring/emails/x+tag@y.z")
}

func TestSwitchEmail(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.SwitchEmail(context.Background(), "abc", "x@y.z", false)
	rs.assertRequest(t, "PATCH", "/api/v2/containers/abc/monitoring/switch-email")
	got := rs.bodyJSON()
	if got["email"] != "x@y.z" || got["isEnabled"] != false {
		t.Errorf("body: %v", got)
	}
}
