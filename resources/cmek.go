package resources

import (
	"context"
	"net/http"
)

// CmekService handles customer managed encryption key operations
type CmekService struct {
	Service *AuraAPIActionsService
}

// Customer managed key methods

// List any customer managed keys. Can filter for a tenant Id
func (c *CmekService) List(ctx context.Context, token *AuthAPIToken, tenantId string) (*GetCmeksResponse, error) {
	endpoint := c.Service.Version + "/customer-managed-keys"

	// There is a tenant ID so we filter by it
	if len(tenantId) > 0 {
		endpoint = endpoint + "?tenantId={" + tenantId
	}

	return makeAuthenticatedRequest[GetCmeksResponse](ctx, c.Service, token, endpoint, http.MethodGet, "application/json", "")
}
