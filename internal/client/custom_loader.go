package client

import (
	"context"
	"net/url"
)

// GenerateCustomLoader — POST /api/v2/container/{id}/custom-loader.
//
// IMPORTANT: this is the ONE endpoint where Stape uses singular `container`
// in the path (everything else is plural `containers`).
//
// Body satisfies ContainerCustomLoaderDTO. All fields optional in the spec:
//
//	webGtmId, domain, source (enum), dataLayerObjectName,
//	userIdentifierType (enum), userIdentifierValue, sameOriginPath
func (c *Client) GenerateCustomLoader(ctx context.Context, container string, body any) ([]byte, error) {
	return c.Post(ctx, "/api/v2/container/"+url.PathEscape(container)+"/custom-loader", body)
}
