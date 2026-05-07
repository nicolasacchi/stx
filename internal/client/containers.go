package client

import (
	"context"
	"net/url"
)

// ListContainers — GET /api/v2/containers.
func (c *Client) ListContainers(ctx context.Context, limit int) ([]byte, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", intStr(limit))
	}
	return c.Get(ctx, "/api/v2/containers", q)
}

// GetContainer — GET /api/v2/containers/{identifier}.
func (c *Client) GetContainer(ctx context.Context, id string) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(id), nil)
}
