package client

import "context"

// ListUsers — GET /api/v2/users.
// Used by `stx config doctor` as the cheapest authenticated probe.
func (c *Client) ListUsers(ctx context.Context) ([]byte, error) {
	return c.Get(ctx, "/api/v2/users", nil)
}
