package client

import (
	"context"
	"net/url"
	"strconv"
)

// LogsQuery — applies to aggregated and detailed log endpoints.
// start/end are Unix seconds (zero means "omit"). Platform/EventType are optional filters.
type LogsQuery struct {
	Start     int64
	End       int64
	Platform  string
	EventType string
}

func (q LogsQuery) values() url.Values {
	v := url.Values{}
	if q.Start > 0 {
		v.Set("start", strconv.FormatInt(q.Start, 10))
	}
	if q.End > 0 {
		v.Set("end", strconv.FormatInt(q.End, 10))
	}
	if q.Platform != "" {
		v.Set("platform", q.Platform)
	}
	if q.EventType != "" {
		v.Set("eventType", q.EventType)
	}
	return v
}

// ListLogsAggregated — GET /api/v2/containers/{id}/monitoring/logs/outgoing/aggregated.
func (c *Client) ListLogsAggregated(ctx context.Context, container string, q LogsQuery) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring/logs/outgoing/aggregated", q.values())
}

// ListLogsDetailed — GET /api/v2/containers/{id}/monitoring/logs/outgoing/detailed.
// Response is wrapped in {"body": {"items": [...]}, "error": {...}}.
func (c *Client) ListLogsDetailed(ctx context.Context, container string, q LogsQuery) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring/logs/outgoing/detailed", q.values())
}

// GetLogTrace — GET /api/v2/containers/{id}/monitoring/logs/outgoing/trace.
// date is Unix seconds (zero = omit; API defaults to recent window).
func (c *Client) GetLogTrace(ctx context.Context, container, traceID string, date int64) ([]byte, error) {
	v := url.Values{}
	if traceID != "" {
		v.Set("traceId", traceID)
	}
	if date > 0 {
		v.Set("date", strconv.FormatInt(date, 10))
	}
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring/logs/outgoing/trace", v)
}

// --- Monitoring rules CRUD ---

// ListRules — GET /api/v2/containers/{id}/monitoring.
func (c *Client) ListRules(ctx context.Context, container string) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring", nil)
}

// GetRule — GET /api/v2/containers/{id}/monitoring/{rule}.
func (c *Client) GetRule(ctx context.Context, container, rule string) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring/"+url.PathEscape(rule), nil)
}

// CreateRule — POST /api/v2/containers/{id}/monitoring.
// Body satisfies ContainerMonitoringFormType (complex — use --from-file).
func (c *Client) CreateRule(ctx context.Context, container string, body any) ([]byte, error) {
	return c.Post(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring", body)
}

// UpdateRule — PUT /api/v2/containers/{id}/monitoring/{rule}.
// Same body shape as CreateRule.
func (c *Client) UpdateRule(ctx context.Context, container, rule string, body any) ([]byte, error) {
	return c.Put(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring/"+url.PathEscape(rule), body)
}

// DeleteRule — DELETE /api/v2/containers/{id}/monitoring/{rule}.
func (c *Client) DeleteRule(ctx context.Context, container, rule string) ([]byte, error) {
	return c.Delete(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring/"+url.PathEscape(rule), nil)
}

// SwitchRule — PATCH /api/v2/containers/{id}/monitoring/{rule}.
// Body: {"enabled": bool}.
func (c *Client) SwitchRule(ctx context.Context, container, rule string, enabled bool) ([]byte, error) {
	return c.Patch(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring/"+url.PathEscape(rule), map[string]bool{"enabled": enabled})
}

// ResolveRule — POST /api/v2/containers/{id}/monitoring/{rule}/resolve.
// No body. Marks a fired rule as handled.
func (c *Client) ResolveRule(ctx context.Context, container, rule string) ([]byte, error) {
	return c.Post(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring/"+url.PathEscape(rule)+"/resolve", nil)
}

// --- Monitoring emails CRUD ---

// ListEmails — GET /api/v2/containers/{id}/monitoring/emails.
func (c *Client) ListEmails(ctx context.Context, container string) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring/emails", nil)
}

// AddEmail — POST /api/v2/containers/{id}/monitoring/emails.
// Body: {"email": "<addr>"}.
func (c *Client) AddEmail(ctx context.Context, container, email string) ([]byte, error) {
	return c.Post(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring/emails", map[string]string{"email": email})
}

// DeleteEmail — DELETE /api/v2/containers/{id}/monitoring/emails/{email}.
// Email is a path param; URL-encoded since it contains @.
func (c *Client) DeleteEmail(ctx context.Context, container, email string) ([]byte, error) {
	return c.Delete(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring/emails/"+url.PathEscape(email), nil)
}

// SwitchEmail — PATCH /api/v2/containers/{id}/monitoring/switch-email.
// Body: {"email": <addr>, "isEnabled": bool}.
func (c *Client) SwitchEmail(ctx context.Context, container, email string, enabled bool) ([]byte, error) {
	body := map[string]any{"email": email, "isEnabled": enabled}
	return c.Patch(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/monitoring/switch-email", body)
}
