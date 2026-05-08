package client

import (
	"context"
	"testing"
)

func TestListSchedules(t *testing.T) {
	rs := newRecordingServer(t, 200, `{"body":[]}`)
	defer rs.close()
	c := rs.client("k", "")
	c.ListSchedules(context.Background(), "abc")
	rs.assertRequest(t, "GET", "/api/v2/containers/abc/schedules")
}

func TestUpdateSchedules(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	body := map[string]any{"containerSchedules": []map[string]any{{
		"path": "/x", "frequencyType": "onceADay",
	}}}
	c.UpdateSchedules(context.Background(), "abc", body)
	rs.assertRequest(t, "PUT", "/api/v2/containers/abc/schedules")
}
