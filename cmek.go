package aura

import (
	"context"
	"net/http"
)

// Customer Managed Encryption Keys

type GetCmeksResponse struct {
	Data []GetCmeksData `json:"data"`
}

type GetCmeksData struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	TenantId string `json:"tenant_id"`
}

// CmekService handles customer managed encryption key operations
type CmekService struct {
	Service *AuraAPIActionsService
}

// Customer managed key methods

// List any customer managed keys. Can filter for a tenant Id
func (c *CmekService) List(ctx context.Context, tenantId string) (*GetCmeksResponse, error) {
	// Get or update token if needed
	err := c.Service.authMgr.getToken(ctx, *c.Service.transport)
	if err != nil { // Token process failed
		return nil, err
	}

	content := "application/json"
	auth := c.Service.authMgr.Type + " " + c.Service.authMgr.Token
	endpoint := c.Service.Config.Version + "/customer-managed-keys"

	// There is a tenant ID so we filter by it
	if len(tenantId) > 0 {
		endpoint = endpoint + "?tenantId={" + tenantId
	}

	return makeAuthenticatedRequest[GetCmeksResponse](ctx, *c.Service.transport, auth, endpoint, http.MethodGet, content, "")
}
