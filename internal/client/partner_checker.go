package client

import "context"

// CreatePartnerChecker — POST /api/v2/partner-tracking-checker.
// Body: {siteUrl, callbackUrl} — both expect full URLs (incl. https://).
// Consumes monthly quota.
func (c *Client) CreatePartnerChecker(ctx context.Context, siteURL, callbackURL string) ([]byte, error) {
	body := map[string]string{
		"siteUrl":     siteURL,
		"callbackUrl": callbackURL,
	}
	return c.Post(ctx, "/api/v2/partner-tracking-checker", body)
}

// GetPartnerCheckerLimit — GET /api/v2/partner-tracking-checker/limit.
// Returns remaining monthly quota.
func (c *Client) GetPartnerCheckerLimit(ctx context.Context) ([]byte, error) {
	return c.Get(ctx, "/api/v2/partner-tracking-checker/limit", nil)
}
