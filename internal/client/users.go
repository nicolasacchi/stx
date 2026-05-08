package client

import (
	"context"
	"net/url"
)

// ListUsers — GET /api/v2/users.
// Agency sub-user management endpoint; returns 403 on personal accounts.
func (c *Client) ListUsers(ctx context.Context) ([]byte, error) {
	return c.Get(ctx, "/api/v2/users", nil)
}

// GetUser — GET /api/v2/users/{identifier}.
func (c *Client) GetUser(ctx context.Context, id string) ([]byte, error) {
	return c.Get(ctx, "/api/v2/users/"+url.PathEscape(id), nil)
}

// CreateUser — POST /api/v2/users.
// Body satisfies UserForm: required username, password; optional nameFirst,
// nameLast, sendEmail.
func (c *Client) CreateUser(ctx context.Context, body any) ([]byte, error) {
	return c.Post(ctx, "/api/v2/users", body)
}

// AttachUser — POST /api/v2/users/attach-user.
// Body satisfies UserAttachUserForm: required email, hasNoProducts;
// optional productName.
func (c *Client) AttachUser(ctx context.Context, body any) ([]byte, error) {
	return c.Post(ctx, "/api/v2/users/attach-user", body)
}

// DetachUser — POST /api/v2/users/{id}/detach-user.
func (c *Client) DetachUser(ctx context.Context, id string) ([]byte, error) {
	return c.Post(ctx, "/api/v2/users/"+url.PathEscape(id)+"/detach-user", nil)
}

// ExportUsersCSV — GET /api/v2/users/export-csv.
// Returns text/csv. Caller must NOT pipe through the JSON output layer.
func (c *Client) ExportUsersCSV(ctx context.Context) ([]byte, error) {
	return c.Get(ctx, "/api/v2/users/export-csv", nil)
}
