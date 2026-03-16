package aura

import (
	"context"
	"encoding/json"
	"log/slog"
	"path"

	utils "github.com/LackOfMorals/aura-client/internal/utils"
)

// GDS Sessions

// List returns all GDS sessions accessible to the authenticated user
func (g *gDSSessionService) List(ctx context.Context) (*GetGDSSessionListResponse, error) {
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

// Get returns information on a single GDS session
func (g *gDSSessionService) Get(ctx context.Context, gdsSessionID string) (*GetGDSSessionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	g.logger.DebugContext(ctx, "getting GDS session", slog.String("sessionID", gdsSessionID))

	resp, err := g.api.Get(ctx, path.Join("graph-analytics", "sessions", gdsSessionID))
	if err != nil {
		g.logger.ErrorContext(ctx, "failed to get GDS session", slog.String("error", err.Error()))
		return nil, err
	}

	var result GetGDSSessionResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		g.logger.ErrorContext(ctx, "failed to unmarshal GDS session response", slog.String("error", err.Error()))
		return nil, err
	}

	g.logger.DebugContext(ctx, "GDS session obtained successfully", slog.Int("count", len(result.Data)))
	return &result, nil
}

// Create creates a new GDS session
func (g *gDSSessionService) Create(ctx context.Context, gdsSessionConfigRequest *CreateGDSSessionConfigData) (*GetGDSSessionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	g.logger.DebugContext(ctx, "creating GDS session")

	body, err := utils.Marshal(gdsSessionConfigRequest)
	if err != nil {
		g.logger.ErrorContext(ctx, "failed to marshal create gds session request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := g.api.Post(ctx, path.Join("graph-analytics", "sessions"), string(body))
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

// Estimate estimates the size of a new GDS session
func (g *gDSSessionService) Estimate(ctx context.Context, gdsSessionSizeEstimateRequest *GetGDSSessionSizeEstimation) (*GDSSessionSizeEstimationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	g.logger.DebugContext(ctx, "estimating GDS session")

	body, err := utils.Marshal(gdsSessionSizeEstimateRequest)
	if err != nil {
		g.logger.ErrorContext(ctx, "failed to marshal estimate gds session request", slog.String("error", err.Error()))
		return nil, err
	}

	resp, err := g.api.Post(ctx, path.Join("graph-analytics", "sessions", "sizing"), string(body))
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

// Delete deletes a GDS session
func (g *gDSSessionService) Delete(ctx context.Context, gdsSessionID string) (*DeleteGDSSessionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	g.logger.DebugContext(ctx, "deleting a GDS session", slog.String("sessionID", gdsSessionID))

	resp, err := g.api.Delete(ctx, path.Join("graph-analytics", "sessions", gdsSessionID))
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
