package aura

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/LackOfMorals/aura-client/internal/utils"
)

// Tenants

// List returns all tenants accessible to the authenticated user
func (t *tenantService) List(ctx context.Context) (*ListTenantsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	t.logger.DebugContext(ctx, "listing tenants")

	resp, err := t.api.Get(ctx, "tenants")
	if err != nil {
		t.logger.ErrorContext(ctx, "failed to list tenants", slog.String("error", err.Error()))
		return nil, err
	}

	var result ListTenantsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		t.logger.ErrorContext(ctx, "failed to unmarshal tenants response", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(ctx, "tenants listed successfully", slog.Int("count", len(result.Data)))
	return &result, nil
}

// Get retrieves details for a specific tenant by ID
func (t *tenantService) Get(ctx context.Context, tenantID string) (*GetTenantResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	if err := utils.ValidateTenantID(tenantID); err != nil {
		t.logger.ErrorContext(ctx, "invalid tenant Id ", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(ctx, "getting tenant details", slog.String("tenantID", tenantID))

	resp, err := t.api.Get(ctx, "tenants/"+tenantID)
	if err != nil {
		t.logger.ErrorContext(ctx, "failed to get tenant details", slog.String("tenantID", tenantID), slog.String("error", err.Error()))
		return nil, err
	}

	var result GetTenantResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		t.logger.ErrorContext(ctx, "failed to unmarshal tenant response", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(ctx, "tenant obtained successfully", slog.String("name", result.Data.Name))
	return &result, nil
}

// GetMetrics retrieves the Prometheus metrics URL for a specific tenant
func (t *tenantService) GetMetrics(ctx context.Context, tenantID string) (*GetTenantMetricsURLResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	if err := utils.ValidateTenantID(tenantID); err != nil {
		t.logger.ErrorContext(ctx, "invalid tenant Id ", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(ctx, "getting tenant prometheus metrics url", slog.String("tenantID", tenantID))

	resp, err := t.api.Get(ctx, "tenants/"+tenantID+"/metrics-integration")
	if err != nil {
		return nil, err
	}

	var result GetTenantMetricsURLResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		t.logger.ErrorContext(ctx, "failed to unmarshal tenant metrics url response", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(ctx, "tenant metrics url obtained successfully", slog.String("endpoint", result.Data.Endpoint))
	return &result, nil
}
