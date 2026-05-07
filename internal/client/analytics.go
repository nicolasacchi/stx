package client

import (
	"context"
	"net/url"
	"strconv"
)

// AnalyticsQuery — start/end as Unix seconds (zero means "omit").
type AnalyticsQuery struct {
	Start int64
	End   int64
}

func (q AnalyticsQuery) values() url.Values {
	v := url.Values{}
	if q.Start > 0 {
		v.Set("start", strconv.FormatInt(q.Start, 10))
	}
	if q.End > 0 {
		v.Set("end", strconv.FormatInt(q.End, 10))
	}
	return v
}

// GetAnalyticsInfo — GET /api/v2/containers/{id}/analytics/info.
func (c *Client) GetAnalyticsInfo(ctx context.Context, container string) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/analytics/info", nil)
}

// GetAnalyticsBrowsers — GET /api/v2/containers/{id}/analytics/browsers.
// Returns array of {name, count, adBlock}.
func (c *Client) GetAnalyticsBrowsers(ctx context.Context, container string, q AnalyticsQuery) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/analytics/browsers", q.values())
}

// GetAnalyticsClients — GET /api/v2/containers/{id}/analytics/clients.
// Returns array of {date, clients[...]}.
func (c *Client) GetAnalyticsClients(ctx context.Context, container string, q AnalyticsQuery) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/analytics/clients", q.values())
}
