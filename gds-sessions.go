package aura

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/LackOfMorals/aura-client/internal/api"
	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// GDS Sessions

// GetGDSSessionListResponse contains a list of GDS sessions
type GetGDSSessionListResponse struct {
	Data []GetGDSSessionData `json:"data"`
}

// GetGDSSessionResponse contains information about a single GDS Session
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

type CreateGDSSessionConfigData struct {
	Name          string `json:"name"`
	Ttl           string `json:"ttl"`
	TenantId      string `json:"tenant_id"`
	InstanceId    string `json:"instance_id"`
	DatabaseId    string `json:"database_uuid"`
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
	Memory        string `json:"memory"`
}

type GetGDSSessionSizeEstimation struct {
	NodeCount                 int      `json:"node_count"`
	NodePropertyCount         int      `json:"node_property_count"`
	NodeLabelCount            int      `json:"node_label_count"`
	RelationshipCount         int      `json:"relationship_count"`
	RelationshipPropertyCount int      `json:"relationship_property_count"`
	AlgorithmCategories       []string `json:"algorithm_categories"`
}

type GDSSessionSizeEstimationResponse struct {
	Data GDSSessionSizeEstimationData `json:"data"`
}

type GDSSessionSizeEstimationData struct {
	EstimatedMemory string `json:"estimated_memory"`
	RecommendedSize string `json:"recommended_size"`
}

type DeleteGDSSessionResponse struct {
	Data DeleteGDSSession `json:"data"`
}

type DeleteGDSSession struct {
	ID string `json:"id"`
}

// gDSSessionService handles GDS Session operations
type gDSSessionService struct {
	api    api.APIRequestService
	ctx    context.Context
	logger *slog.Logger
}

// List returns all GDS sessions accessible to the authenticated user
func (g *gDSSessionService) List() (*GetGDSSessionListResponse, error) {
	g.logger.DebugContext(g.ctx, "listing GDS sessions")

	resp, err := g.api.Get(g.ctx, "graph-analytics/sessions")
	if err != nil {
		g.logger.ErrorContext(g.ctx, "failed to list GDS sessions", slog.String("error", err.Error()))
		return nil, err
	}

	var result GetGDSSessionListResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		g.logger.ErrorContext(g.ctx, "failed to unmarshal GDS sessions response", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(g.ctx, "GDS sessions listed successfully", slog.Int("count", len(result.Data)))
	return &result, nil
}

// Get information on a GDS session accessible to the authenticated user
func (g *gDSSessionService) Get(GDSSessionID string) (*GetGDSSessionResponse, error) {
	g.logger.DebugContext(g.ctx, "listing GDS sessions")

	resp, err := g.api.Get(g.ctx, "graph-analytics/sessions/"+GDSSessionID)
	if err != nil {
		g.logger.ErrorContext(g.ctx, "failed to get GDS session", slog.String("error", err.Error()))
		return nil, err
	}

	var result GetGDSSessionResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		g.logger.ErrorContext(g.ctx, "failed to unmarshal GDS session response", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(g.ctx, "GDS session obtained successfully", slog.Int("count", len(result.Data)))
	return &result, nil
}

// Create a new GDS session
func (g *gDSSessionService) Create(GDSSessionConfigRequest *CreateGDSSessionConfigData) (*GetGDSSessionResponse, error) {
	g.logger.DebugContext(g.ctx, "creating GDS sessions")

	body, err := utils.Marshall(GDSSessionConfigRequest)
	if err != nil {
		g.logger.ErrorContext(g.ctx, "failed to marshal create gds session request", slog.String("error", err.Error()))
		return nil, err
	}
	resp, err := g.api.Post(g.ctx, "graph-analytics/sessions", string(body))
	if err != nil {
		g.logger.ErrorContext(g.ctx, "failed to create GDS session", slog.String("error", err.Error()))
		return nil, err
	}

	var result GetGDSSessionResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		g.logger.ErrorContext(g.ctx, "failed to unmarshal create GDS sessions response", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(g.ctx, "GDS session created successfully")
	return &result, nil
}

// Estimate the size of a new GDS session
func (g *gDSSessionService) Estimate(GDSSessionSizeEstimateRequest *GetGDSSessionSizeEstimation) (*GDSSessionSizeEstimationResponse, error) {
	g.logger.DebugContext(g.ctx, "estimating GDS sessions")

	body, err := utils.Marshall(GDSSessionSizeEstimateRequest)
	if err != nil {
		g.logger.ErrorContext(g.ctx, "failed to marshal estimate gds session request", slog.String("error", err.Error()))
		return nil, err
	}
	resp, err := g.api.Post(g.ctx, "graph-analytics/sessions/sizing", string(body))
	if err != nil {
		g.logger.ErrorContext(g.ctx, "failed to estimate GDS session", slog.String("error", err.Error()))
		return nil, err
	}

	var result GDSSessionSizeEstimationResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		g.logger.ErrorContext(g.ctx, "failed to unmarshal estimate GDS sessions response", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(g.ctx, "GDS session estimated successfully")
	return &result, nil
}

// Deletes a GDS Session
func (g *gDSSessionService) Delete(GDSSessionID string) (*DeleteGDSSessionResponse, error) {
	g.logger.DebugContext(g.ctx, "deleting a GDS session")

	resp, err := g.api.Delete(g.ctx, "graph-analytics/sessions/"+GDSSessionID)
	if err != nil {
		g.logger.ErrorContext(g.ctx, "failed to delete a GDS session", slog.String("error", err.Error()))
		return nil, err
	}

	var result DeleteGDSSessionResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		g.logger.ErrorContext(g.ctx, "failed to unmarshal GDS session delete response", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(g.ctx, "GDS session deleted successfully")
	return &result, nil
}
