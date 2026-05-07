package client

import (
	"context"
	"net/url"
)

// ListProxyFiles — GET /api/v2/containers/{id}/proxy-files.
// Returns an array (no envelope wrapping).
func (c *Client) ListProxyFiles(ctx context.Context, container string) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/proxy-files", nil)
}

// UpdateProxyFiles — PUT /api/v2/containers/{id}/proxy-files.
// Body: ContainerProxyFileEditForm {containerProxyFiles: [{originalFileUrl, customPath, cacheMaxAge}]}.
// cacheMaxAge is one of a fixed enum (-1, 120, 300, 1200, 1800, 3600, …, 31536000).
// Use --from-file <path> to pass the full body.
func (c *Client) UpdateProxyFiles(ctx context.Context, container string, body any) ([]byte, error) {
	return c.Put(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/proxy-files", body)
}
