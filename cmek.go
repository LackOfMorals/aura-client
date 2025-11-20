package aura

import (
	"log/slog"
	"net/http"
)

// Customer Managed Encryption Keys

type getCmeksResponse struct {
	Data []getCmeksData `json:"data"`
}

type getCmeksData struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	TenantId string `json:"tenant_id"`
}

// CmekService handles customer managed encryption key operations
type cmekService struct {
	service *AuraAPIClient
	logger  *slog.Logger
}

// Customer managed key methods

// List any customer managed keys. Can filter for a tenant Id
func (c *cmekService) List(tenantId string) (*getCmeksResponse, error) {
	c.logger.DebugContext(c.service.config.ctx, "listing custom managed keys")

	endpoint := c.service.config.version + "/customer-managed-keys"

	resp, err := makeServiceRequest[getCmeksResponse](c.service.config.ctx, *c.service.transport, c.service.authMgr, endpoint, http.MethodGet, "", c.logger)
	if err != nil {
		c.logger.ErrorContext(c.service.config.ctx, "failed to obtain customer managed keys", slog.String("error", err.Error()))
		return nil, err
	}

	c.logger.DebugContext(c.service.config.ctx, "obtained customer managed keys", slog.Int("count", len(resp.Data)))
	return resp, nil
}
