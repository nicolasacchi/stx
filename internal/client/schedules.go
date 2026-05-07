package client

import (
	"context"
	"net/url"
)

// ListSchedules — GET /api/v2/containers/{id}/schedules.
// Returns an array (no envelope wrapping).
func (c *Client) ListSchedules(ctx context.Context, container string) ([]byte, error) {
	return c.Get(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/schedules", nil)
}

// UpdateSchedules — PUT /api/v2/containers/{id}/schedules.
// Body: ContainerScheduleEditForm {containerSchedules: [{path, frequencyType (onceADay|everyHour), domain.identifier?, hour?, minute?}]}.
// Use --from-file <path> to pass the full body.
func (c *Client) UpdateSchedules(ctx context.Context, container string, body any) ([]byte, error) {
	return c.Put(ctx, "/api/v2/containers/"+url.PathEscape(container)+"/schedules", body)
}
