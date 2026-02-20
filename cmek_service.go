package aura

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/LackOfMorals/aura-client/internal/utils"
)

// List returns all customer-managed encryption keys, optionally filtered by tenant
func (c *cmekService) List(ctx context.Context, tenantId string) (*GetCmeksResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	c.logger.DebugContext(ctx, "listing customer managed keys")

	endpoint := "customer-managed-keys"

	switch tenantIdLen := len(tenantId); tenantIdLen {
	case 0:
		// empty string, no tenant filter
		break
	case 36:
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
