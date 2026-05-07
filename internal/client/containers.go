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

// CreateContainer — POST /api/v2/containers.
//
// body must satisfy ContainerCreateForm: required fields are
// name, code (≥64 chars), codeSettings, anonymizeOptions, zone.
// Pass either a typed struct or a raw map[string]any / []byte
// loaded from a JSON file.
func (c *Client) CreateContainer(ctx context.Context, body any) ([]byte, error) {
	return c.Post(ctx, "/api/v2/containers", body)
}

// UpdateContainer — PUT /api/v2/containers/{identifier}.
//
// body must satisfy ContainerUpdateForm: required fields are
// name, code, codeSettings, anonymizeOptions (NO zone). PUT semantics
// are full-document replace; partial PATCH is not supported by Stape.
func (c *Client) UpdateContainer(ctx context.Context, id string, body any) ([]byte, error) {
	return c.Put(ctx, "/api/v2/containers/"+url.PathEscape(id), body)
}

// DeleteContainer — DELETE /api/v2/containers/{identifier}.
//
// Body satisfies SubscriptionCancelForm. cancelReason is form-level
// optional; pass nil to send empty body. If Stape rejects with 400
// the caller can supply a populated body via SubscriptionCancelForm.
func (c *Client) DeleteContainer(ctx context.Context, id string, body any) ([]byte, error) {
	return c.Delete(ctx, "/api/v2/containers/"+url.PathEscape(id), body)
}

// TransferContainer — PUT /api/v2/containers/{identifier}/transfer.
//
// Body is ProductTransferFormType: {"email": "<new-owner>"}.
func (c *Client) TransferContainer(ctx context.Context, id, email string) ([]byte, error) {
	body := map[string]string{"email": email}
	return c.Put(ctx, "/api/v2/containers/"+url.PathEscape(id)+"/transfer", body)
}

// SubscriptionCancelForm matches the Stape DELETE /containers/{id} body shape.
// All fields optional at the form level; if cancelReason is set, both
// setup and cancel arrays must be non-nil per the spec.
type SubscriptionCancelForm struct {
	CancelReason *SubscriptionCancelReason `json:"cancelReason,omitempty"`
}

type SubscriptionCancelReason struct {
	Setup  []OptionForm `json:"setup"`
	Cancel []OptionForm `json:"cancel"`
}

type OptionForm struct {
	Type string `json:"type"`
}
