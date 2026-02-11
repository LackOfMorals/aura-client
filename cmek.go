package aura

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
	"github.com/LackOfMorals/aura-client/internal/utils"
)

// Customer Managed Encryption Keys

// GetCmeksResponse contains a list of customer managed encryption keys
type GetCmeksResponse struct {
	Data []GetCmeksData `json:"data"`
}

type GetCmeksData struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	TenantId string `json:"tenant_id"`
}

// cmekService handles customer managed encryption key operations
type cmekService struct {
	api     api.RequestService
	ctx     context.Context
	timeout time.Duration
	logger  *slog.Logger
}

// List returns all customer-managed encryption keys, optionally filtered by tenant
func (c *cmekService) List(tenantId string) (*GetCmeksResponse, error) {
	// Create child context with timeout for this operation
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()

	c.logger.DebugContext(ctx, "listing customer managed keys")

	endpoint := "customer-managed-keys"

	// Check / validate for tenant Id. If it is ok, add filter to endpoint
	switch tenantIdLen := len(tenantId); tenantIdLen {
	case 0:
		// empty string, do not need to do anything
		break
	case 36:
		// Check if tenant ID has correct format
		if err := utils.ValidateTenantID(tenantId); err != nil {
			return nil, err
		}
		endpoint = endpoint + "?tenantid=" + tenantId
	default:
		return nil, fmt.Errorf("tenant ID must be in the format of hex 8-4-4-12 pattern")
	}

	resp, err := c.api.Get(ctx, endpoint)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to obtain customer managed keys", slog.String("error", err.Error()))
		return nil, err
	}

	var result GetCmeksResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		c.logger.ErrorContext(ctx, "failed to unmarshal cmek response", slog.String("error", err.Error()))
		return nil, err
	}

	c.logger.DebugContext(ctx, "obtained customer managed keys", slog.Int("count", len(result.Data)))
	return &result, nil
}
