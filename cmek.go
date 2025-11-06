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
	Service *AuraAPIActionsService
	logger  *slog.Logger
}

// Customer managed key methods

// List any customer managed keys. Can filter for a tenant Id
func (c *CmekService) List(ctx context.Context, tenantId string) (*GetCmeksResponse, error) {
	c.logger.DebugContext(ctx, "listing custom managed keys")

	// Get or update token if needed
	err := c.Service.authMgr.getToken(ctx, *c.Service.transport)
	if err != nil { // Token process failed
		c.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := c.Service.authMgr.Type + " " + c.Service.authMgr.Token
	endpoint := c.Service.Config.Version + "/customer-managed-keys"

	c.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetCmeksResponse](ctx, *c.Service.transport, auth, endpoint, http.MethodGet, content, "")
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to obtain customer managed keys", slog.String("error", err.Error()))
		return nil, err
	}

	c.logger.DebugContext(ctx, "obtained customer managed keys", slog.Int("count", len(resp.Data)))
	return resp, nil
}
