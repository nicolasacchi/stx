package client

import (
	"context"
	"testing"
)

func TestGetPowerUp(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.GetPowerUp(context.Background(), "abc", "ad-blocker")
	rs.assertRequest(t, "GET", "/api/v2/containers/abc/power-ups/ad-blocker")
}

func TestSwitchPowerUp_TierA(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.SwitchPowerUp(context.Background(), "abc", "ad-blocker", map[string]any{"isActive": true})
	rs.assertRequest(t, "PATCH", "/api/v2/containers/abc/power-ups/ad-blocker")
	got := rs.bodyJSON()
	if got["isActive"] != true {
		t.Errorf("body: %v", got)
	}
	if _, hasOpts := got["options"]; hasOpts {
		t.Errorf("Tier A should not include options key; got %v", got)
	}
}

func TestSwitchPowerUp_TierB_WithOptions(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	body := map[string]any{
		"isActive": true,
		"options":  map[string]bool{"sp_ip": true},
	}
	c.SwitchPowerUp(context.Background(), "abc", "anonymizer", body)
	rs.assertRequest(t, "PATCH", "/api/v2/containers/abc/power-ups/anonymizer")
	got := rs.bodyJSON()
	opts, _ := got["options"].(map[string]any)
	if opts == nil || opts["sp_ip"] != true {
		t.Errorf("body: %v", got)
	}
}

func TestSwitchPowerUp_PreviewHeaderConfigPath(t *testing.T) {
	rs := newRecordingServer(t, 200, `{}`)
	defer rs.close()
	c := rs.client("k", "")
	c.SwitchPowerUp(context.Background(), "abc", "preview-header-config", map[string]any{"isActive": false})
	rs.assertRequest(t, "PATCH", "/api/v2/containers/abc/power-ups/preview-header-config")
}
