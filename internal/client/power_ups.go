package client

import (
	"context"
	"net/url"
)

// GetPowerUp — GET /api/v2/containers/{id}/power-ups/{type}.
//
// Type names match the spec paths (e.g. "ad-blocker", "anonymizer",
// "custom-loader", "preview-header-config", …). The CLI exposes
// "header-config" as a brief alias; client method takes the
// spec-exact name.
func (c *Client) GetPowerUp(ctx context.Context, container, powerUpType string) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/power-ups/"+url.PathEscape(powerUpType), nil)
}
