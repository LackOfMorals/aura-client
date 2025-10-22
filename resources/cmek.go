package resources

import (
	"context"
	"net/http"

	"github.com/LackOfMorals/aura-client"
)

// CmekService handles customer managed encryption key operations
type CmekService struct {
	Service *aura.AuraAPIActionsService
}

// Customer managed key methods

// List any customer managed keys. Can filter for a tenant Id
func (c *CmekService) List(ctx context.Context, token *aura.AuthAPIToken, tenantId string) (*aura.GetCmeksResponse, error) {
	endpoint := c.service.AuraAPIVersion + "/customer-managed-keys"

	// There is a tenant ID so we filter by it
	if len(tenantId) > 0 {
		endpoint = endpoint + "?tenantId={" + tenantId
	}

	return makeAuthenticatedRequest[aura.GetCmeksResponse](ctx, c.service, token, endpoint, http.MethodGet, "application/json", "")
}
