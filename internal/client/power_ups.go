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

// SwitchPowerUp — PATCH /api/v2/containers/{id}/power-ups/{type}.
//
// Generic. The body shape varies per power-up type:
//   - Tier A (8 types): {"isActive": bool} only
//   - Tier B (12 types): {"isActive": bool, "options": <type-specific>}
//
// service-account is the special case: options is a string blob, not an object.
//
// The command layer constructs the appropriate body per type; this method
// just wraps the HTTP PATCH.
func (c *Client) SwitchPowerUp(ctx context.Context, container, powerUpType string, body any) ([]byte, error) {
	return c.Patch(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/power-ups/"+url.PathEscape(powerUpType), body)
}
