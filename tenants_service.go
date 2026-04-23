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

// Tenants
// tenantService handles tenant operations
type tenantService struct {
	api     api.RequestService
	timeout time.Duration
	logger  *slog.Logger
}

// List returns all tenants accessible to the authenticated user
func (t *tenantService) List(ctx context.Context) (*ListTenantsResponse, error) {
	if err := ctx.Err(); err != nil {
		t.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}

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
	if err := ctx.Err(); err != nil {
		t.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	if err := utils.ValidateTenantID(tenantID); err != nil {
		t.logger.ErrorContext(ctx, "invalid tenant Id ", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(ctx, "getting tenant details", slog.String("tenantID", tenantID))

	resp, err := t.api.Get(ctx, fmt.Sprintf("tenants/%s", tenantID))
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
	if err := ctx.Err(); err != nil {
		t.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	if err := utils.ValidateTenantID(tenantID); err != nil {
		t.logger.ErrorContext(ctx, "invalid tenant Id ", slog.String("error", err.Error()))
		return nil, err
	}

	t.logger.DebugContext(ctx, "getting tenant prometheus metrics url", slog.String("tenantID", tenantID))

	resp, err := t.api.Get(ctx, fmt.Sprintf("tenants/%s/metrics-integration", tenantID))
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
