package client

import (
	"context"
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

// CreateDomain — POST /api/v2/containers/{c}/domains.
// Body satisfies ContainerDomainForm: required {name, cdnType (stape|custom|none)};
// optional {connectionType (stape|entri), useCnameRecord, useARecord}.
func (c *Client) CreateDomain(ctx context.Context, container string, body any) ([]byte, error) {
	return c.Post(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/domains", body)
}

// UpdateDomain — PUT /api/v2/containers/{c}/domains/{d}.
// Same body schema as CreateDomain.
func (c *Client) UpdateDomain(ctx context.Context, container, domain string, body any) ([]byte, error) {
	return c.Put(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/domains/"+url.PathEscape(domain), body)
}

// ValidateDomain — POST /api/v2/containers/{c}/domains/validate.
// Same body schema as CreateDomain. Returns validation result without
// persisting anything.
func (c *Client) ValidateDomain(ctx context.Context, container string, body any) ([]byte, error) {
	return c.Post(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/domains/validate", body)
}

// RevalidateDomain — POST /api/v2/containers/{c}/domains/{d}/revalidate.
// No body. Re-checks DNS / cert / readiness for an existing domain.
func (c *Client) RevalidateDomain(ctx context.Context, container, domain string) ([]byte, error) {
	return c.Post(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/domains/"+url.PathEscape(domain)+"/revalidate", nil)
}
