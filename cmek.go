package aura

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/LackOfMorals/aura-client/internal/utils"
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

	// Check / validate for tenant Id . If it is ok, add filter to endpoint
	switch tenantIdLen := len(tenantId); tenantIdLen {

	// empty string, do not need to do anything
	case 0:
		break
	case 36:
		// Check if tenant ID has correct format
		err := utils.ValidateTenantID(tenantId)
		if err != nil {
			return nil, err
		}
		endpoint = endpoint + "?tenantid=" + tenantId
	default:
		return nil, fmt.Errorf("Tenant Id must be in the format of hex 8-4-4-12 pattern ")
	}

	resp, err := makeServiceRequest[getCmeksResponse](c.service.config.ctx, *c.service.transport, c.service.authMgr, endpoint, http.MethodGet, "", c.logger)
	if err != nil {
		c.logger.ErrorContext(c.service.config.ctx, "failed to obtain customer managed keys", slog.String("error", err.Error()))
		return nil, err
	}

	c.logger.DebugContext(c.service.config.ctx, "obtained customer managed keys", slog.Int("count", len(resp.Data)))
	return resp, nil
}
