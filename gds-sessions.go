package aura

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// GDS Sessions

// GetGDSSessionResponse contains a list of GDS sessions
type GetGDSSessionResponse struct {
	Data []GetGDSSessionData `json:"data"`
}

type GetGDSSessionData struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Memory        string `json:"memory"`
	InstanceId    string `json:"instance_id"`
	DatabaseId    string `json:"database_uuid"`
	Status        string `json:"status"`
	Create        string `json:"created_at"`
	Host          string `json:"host"`
	Expiry        string `json:"expiry_date"`
	Ttl           string `json:"ttl"`
	UserId        string `json:"user_id"`
	TenantId      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
}

// gDSSessionService handles GDS Session operations
type gDSSessionService struct {
	api    api.APIRequestService
	ctx    context.Context
	logger *slog.Logger
}

// List returns all GDS sessions accessible to the authenticated user
func (g *gDSSessionService) List() (*GetGDSSessionResponse, error) {
	g.logger.DebugContext(g.ctx, "listing GDS sessions")

	resp, err := g.api.Get(g.ctx, "graph-analytics/sessions")
	if err != nil {
		g.logger.ErrorContext(g.ctx, "failed to list GDS sessions", slog.String("error", err.Error()))
		return nil, err
	}

	var result GetGDSSessionResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		g.logger.ErrorContext(g.ctx, "failed to unmarshal GDS sessions response", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(g.ctx, "GDS sessions listed successfully", slog.Int("count", len(result.Data)))
	return &result, nil
}
