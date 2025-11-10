package aura

import (
	"context"
	"log/slog"
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
	service *AuraAPIClient
	logger  *slog.Logger
}

// Customer managed key methods

// List any customer managed keys. Can filter for a tenant Id
func (c *CmekService) List(ctx context.Context, tenantId string) (*GetCmeksResponse, error) {
	c.logger.DebugContext(ctx, "listing custom managed keys")

	// Get or update token if needed
	err := c.service.authMgr.getToken(ctx, *c.service.transport)
	if err != nil { // Token process failed
		c.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := c.service.authMgr.tokenType + " " + c.service.authMgr.token
	endpoint := c.service.config.version + "/customer-managed-keys"

	c.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetCmeksResponse](ctx, *c.service.transport, auth, endpoint, http.MethodGet, content, "")
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to obtain customer managed keys", slog.String("error", err.Error()))
		return nil, err
	}

	c.logger.DebugContext(ctx, "obtained customer managed keys", slog.Int("count", len(resp.Data)))
	return resp, nil
}
