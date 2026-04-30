package aura

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// ============================================================================
// Types
// ============================================================================

// GetGDSSessionListResponse contains a list of GDS sessions.
type GetGDSSessionListResponse struct {
	Data []GetGDSSessionData `json:"data"`
}

// GetGDSSessionResponse contains information about a single GDS session.
type GetGDSSessionResponse struct {
	Data GetGDSSessionData `json:"data"`
}

// GetGDSSessionData holds the fields returned for a single GDS session.
type GetGDSSessionData struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Memory        string `json:"memory"`
	InstanceID    string `json:"instance_id"`
	DatabaseID    string `json:"database_uuid"`
	Status        string `json:"status"`
	Create        string `json:"created_at"`
	Host          string `json:"host"`
	Expiry        string `json:"expiry_date"`
	TTL           string `json:"ttl"`
	UserID        string `json:"user_id"`
	TenantID      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
}

// CreateGDSSessionConfigData holds the configuration required to create a new GDS session.
type CreateGDSSessionConfigData struct {
	Name          string `json:"name"`
	TTL           string `json:"ttl"`
	TenantID      string `json:"tenant_id"`
	InstanceID    string `json:"instance_id"`
	DatabaseID    string `json:"database_uuid"`
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
	Memory        string `json:"memory"`
}

// GetGDSSessionSizeEstimation holds graph statistics used to estimate the memory
// requirements for a new GDS session.
type GetGDSSessionSizeEstimation struct {
	NodeCount                 int      `json:"node_count"`
	NodePropertyCount         int      `json:"node_property_count"`
	NodeLabelCount            int      `json:"node_label_count"`
	RelationshipCount         int      `json:"relationship_count"`
	RelationshipPropertyCount int      `json:"relationship_property_count"`
	AlgorithmCategories       []string `json:"algorithm_categories"`
}

// GDSSessionSizeEstimationResponse wraps the size estimation result.
type GDSSessionSizeEstimationResponse struct {
	Data GDSSessionSizeEstimationData `json:"data"`
}

// GDSSessionSizeEstimationData holds the estimated memory and recommended size tier.
type GDSSessionSizeEstimationData struct {
	EstimatedMemory string `json:"estimated_memory"`
	RecommendedSize string `json:"recommended_size"`
}

// DeleteGDSSessionResponse wraps the response returned when a GDS session is deleted.
type DeleteGDSSessionResponse struct {
	Data DeleteGDSSession `json:"data"`
}

// DeleteGDSSession holds the ID of the deleted session.
type DeleteGDSSession struct {
	ID string `json:"id"`
}

// ============================================================================
// Service
// ============================================================================

// gDSSessionService handles Graph Data Science session operations.
type gDSSessionService struct {
	api     api.RequestService
	timeout time.Duration
	logger  *slog.Logger
}

// List returns all GDS sessions accessible to the authenticated user.
func (g *gDSSessionService) List(ctx context.Context) (*GetGDSSessionListResponse, error) {
	if err := ctx.Err(); err != nil {
		g.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	g.logger.DebugContext(ctx, "listing GDS sessions")

	resp, err := g.api.Get(ctx, "graph-analytics/sessions")
	if err != nil {
		g.logger.ErrorContext(ctx, "failed to list GDS sessions", slog.String("error", err.Error()))
		return nil, err
	}

	var result GetGDSSessionListResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		g.logger.ErrorContext(ctx, "failed to unmarshal GDS sessions response", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(ctx, "GDS sessions listed successfully", slog.Int("count", len(result.Data)))
	return &result, nil
}

// Get returns information on a single GDS session.
func (g *gDSSessionService) Get(ctx context.Context, gdsSessionID string) (*GetGDSSessionResponse, error) {
	if err := ctx.Err(); err != nil {
		g.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	if gdsSessionID == "" {
		return nil, fmt.Errorf("GDS session ID must not be empty")
	}

	g.logger.DebugContext(ctx, "getting GDS session", slog.String("sessionID", gdsSessionID))

	resp, err := g.api.Get(ctx, fmt.Sprintf("graph-analytics/sessions/%s", gdsSessionID))
	if err != nil {
		g.logger.ErrorContext(ctx, "failed to get GDS session", slog.String("error", err.Error()))
		return nil, err
	}

	var result GetGDSSessionResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		g.logger.ErrorContext(ctx, "failed to unmarshal GDS session response", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(ctx, "GDS session obtained successfully")
	return &result, nil
}

// Create creates a new GDS session.
func (g *gDSSessionService) Create(ctx context.Context, gdsSessionConfigRequest *CreateGDSSessionConfigData) (*GetGDSSessionResponse, error) {
	if err := ctx.Err(); err != nil {
		g.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	if gdsSessionConfigRequest == nil {
		return nil, fmt.Errorf("gdsSessionConfigRequest must not be nil")
	}

	g.logger.DebugContext(ctx, "creating GDS session")

	body, err := utils.Marshal(gdsSessionConfigRequest)
	if err != nil {
		g.logger.ErrorContext(ctx, "failed to marshal create gds session request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := g.api.Post(ctx, "graph-analytics/sessions", string(body))
	if err != nil {
		g.logger.ErrorContext(ctx, "failed to create GDS session", slog.String("error", err.Error()))
		return nil, err
	}

	var result GetGDSSessionResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		g.logger.ErrorContext(ctx, "failed to unmarshal create GDS sessions response", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(ctx, "GDS session created successfully")
	return &result, nil
}

// Estimate estimates the size of a new GDS session.
func (g *gDSSessionService) Estimate(ctx context.Context, gdsSessionSizeEstimateRequest *GetGDSSessionSizeEstimation) (*GDSSessionSizeEstimationResponse, error) {
	if err := ctx.Err(); err != nil {
		g.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	if gdsSessionSizeEstimateRequest == nil {
		return nil, fmt.Errorf("gdsSessionSizeEstimateRequest must not be nil")
	}

	g.logger.DebugContext(ctx, "estimating GDS session")

	body, err := utils.Marshal(gdsSessionSizeEstimateRequest)
	if err != nil {
		g.logger.ErrorContext(ctx, "failed to marshal estimate gds session request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := g.api.Post(ctx, "graph-analytics/sessions/sizing", string(body))
	if err != nil {
		g.logger.ErrorContext(ctx, "failed to estimate GDS session", slog.String("error", err.Error()))
		return nil, err
	}

	var result GDSSessionSizeEstimationResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		g.logger.ErrorContext(ctx, "failed to unmarshal estimate GDS sessions response", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(ctx, "GDS session estimated successfully")
	return &result, nil
}

// Delete deletes a GDS session.
func (g *gDSSessionService) Delete(ctx context.Context, gdsSessionID string) (*DeleteGDSSessionResponse, error) {
	if err := ctx.Err(); err != nil {
		g.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	if gdsSessionID == "" {
		return nil, fmt.Errorf("GDS session ID must not be empty")
	}

	g.logger.DebugContext(ctx, "deleting a GDS session", slog.String("sessionID", gdsSessionID))

	resp, err := g.api.Delete(ctx, fmt.Sprintf("graph-analytics/sessions/%s", gdsSessionID))
	if err != nil {
		g.logger.ErrorContext(ctx, "failed to delete a GDS session", slog.String("error", err.Error()))
		return nil, err
	}

	var result DeleteGDSSessionResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		g.logger.ErrorContext(ctx, "failed to unmarshal GDS session delete response", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(ctx, "GDS session deleted successfully")
	return &result, nil
}
