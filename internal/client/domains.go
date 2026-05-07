package client

import (
	"context"
	"errors"
	"net/url"
)

// ListDomains — GET /api/v2/containers/{c}/domains.
func (c *Client) ListDomains(ctx context.Context, container string) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/domains", nil)
}

// GetDomain — GET /api/v2/containers/{c}/domains/{d}.
func (c *Client) GetDomain(ctx context.Context, container, domain string) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/domains/"+url.PathEscape(domain), nil)
}

// DeleteDomain — DELETE /api/v2/containers/{c}/domains/{d}.
func (c *Client) DeleteDomain(ctx context.Context, container, domain string) ([]byte, error) {
	return c.Delete(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/domains/"+url.PathEscape(domain), nil)
}

// --- M3 stubs (write paths) ---
//
// These exist so the package surface is stable. Wired up in M3 along with
// command-layer flag plumbing (--cdn-type, --use-cname-record, etc).

var errM3 = errors.New("M3 — write path not yet implemented; see implementation-brief.md")

func (c *Client) CreateDomain(ctx context.Context, container string, body any) ([]byte, error) {
	return nil, errM3
}

func (c *Client) UpdateDomain(ctx context.Context, container, domain string, body any) ([]byte, error) {
	return nil, errM3
}

func (c *Client) ValidateDomain(ctx context.Context, container string, body any) ([]byte, error) {
	return nil, errM3
}

func (c *Client) RevalidateDomain(ctx context.Context, container, domain string) ([]byte, error) {
	return nil, errM3
}
