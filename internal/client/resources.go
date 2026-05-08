package client

import (
	"context"
	"fmt"
	"net/url"
)

// ResourceKinds is the hand-maintained list of valid `resources <kind>` arguments.
//
// MUST match the spec's 15 paths under /api/v2/resources/. `make refresh-spec`
// does NOT regenerate this slice; if Stape adds a new resource endpoint, append
// the kebab-case kind here by hand.
var ResourceKinds = []string{
	"container-anonymize-options",
	"container-cookie-keeper-options",
	"container-domain-cdn-types",
	"container-domain-record-types",
	"container-monitoring-compare-to-type",
	"container-monitoring-comparison-type",
	"container-monitoring-log-types",
	"container-monitoring-periods-type",
	"container-monitoring-rules-fields-type",
	"container-monitoring-rules-operators-type",
	"container-monitoring-rules-values-type",
	"container-proxy-file-cache-max-ages",
	"container-schedule-types",
	"container-statuses",
	"container-zones",
}

// IsValidResourceKind returns true iff kind appears in ResourceKinds.
func IsValidResourceKind(kind string) bool {
	for _, k := range ResourceKinds {
		if k == kind {
			return true
		}
	}
	return false
}

// GetResource — GET /api/v2/resources/<kind>.
// kind must be one of the 15 entries in ResourceKinds.
func (c *Client) GetResource(ctx context.Context, kind string) ([]byte, error) {
	if !IsValidResourceKind(kind) {
		return nil, fmt.Errorf("invalid resource kind %q; see `stx resources kinds`", kind)
	}
	return c.Get(ctx, "/api/v2/resources/"+url.PathEscape(kind), nil)
}
