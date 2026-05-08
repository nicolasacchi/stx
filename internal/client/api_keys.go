package client

import (
	"context"
	"net/url"
)

// ListAPIKeys — GET /api/v2/users/api-key.
// Returns metadata for keys (id + name); the secret value is shown only at
// creation time and not echoed on subsequent reads.
func (c *Client) ListAPIKeys(ctx context.Context) ([]byte, error) {
	return c.Get(ctx, "/api/v2/users/api-key", nil)
}

// CreateAPIKey — POST /api/v2/users/api-key.
// Body satisfies CreateApiKeyDTO: required `name` only.
// IMPORTANT: the response includes the new key's secret value — saved ONCE.
func (c *Client) CreateAPIKey(ctx context.Context, name string) ([]byte, error) {
	return c.Post(ctx, "/api/v2/users/api-key", map[string]string{"name": name})
}

// DeleteAPIKey — DELETE /api/v2/users/api-key/{id}.
// WARNING: do not delete the key currently in use; you'll lock yourself out.
func (c *Client) DeleteAPIKey(ctx context.Context, id string) ([]byte, error) {
	return c.Delete(ctx, "/api/v2/users/api-key/"+url.PathEscape(id), nil)
}
