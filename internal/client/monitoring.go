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
