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

// Snapshots
// snapshotService handles snapshot operations
type snapshotService struct {
	api     api.RequestService
	timeout time.Duration
	logger  *slog.Logger
}

// List returns snapshots for an instance, optionally filtered by date (YYYY-MM-DD)
func (s *snapshotService) List(ctx context.Context, instanceID string, snapshotDate *SnapshotDate) (*GetSnapshotsResponse, error) {
	// Guard against the caller passing a cancelled context
	// Check ctx.Err() at entry and return early:
	if err := ctx.Err(); err != nil {
		s.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	s.logger.DebugContext(ctx, "listing snapshots", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		s.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	endpoint := fmt.Sprintf("instances/%s/snapshots", instanceID)

	if snapshotDate != nil {
		endpoint += fmt.Sprintf("?date=%04d-%02d-%02d", snapshotDate.Year, int(snapshotDate.Month), snapshotDate.Day)
		s.logger.DebugContext(ctx, "Endpoint:", slog.String("URL", endpoint))
	}

	resp, err := s.api.Get(ctx, endpoint)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list snapshots", slog.String("error", err.Error()))
		return nil, err
	}

	var result GetSnapshotsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		s.logger.ErrorContext(ctx, "failed to unmarshal snapshots response", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(ctx, "snapshots listed successfully", slog.Int("count", len(result.Data)))
	return &result, nil
}

// Get returns the details for a snapshot of an instance
func (s *snapshotService) Get(ctx context.Context, instanceID string, snapshotID string) (*GetSnapshotDataResponse, error) {
	// Guard against the caller passing a cancelled context
	// Check ctx.Err() at entry and return early:
	if err := ctx.Err(); err != nil {
		s.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		s.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	if err := utils.ValidateSnapshotID(snapshotID); err != nil {
		s.logger.ErrorContext(ctx, "invalid snapshot Id ", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(ctx, "get snapshot details", slog.String("snapshotID", snapshotID), slog.String("instanceID", instanceID))

	resp, err := s.api.Get(ctx, fmt.Sprintf("instances/%s/snapshots/%s", instanceID, snapshotID))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get snapshot", slog.String("error", err.Error()))
		return nil, err
	}

	var result GetSnapshotDataResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		s.logger.ErrorContext(ctx, "failed to unmarshal snapshots response", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(ctx, "snapshot details obtained")
	return &result, nil
}

// Create triggers an on-demand snapshot for an instance
func (s *snapshotService) Create(ctx context.Context, instanceID string) (*CreateSnapshotResponse, error) {
	// Guard against the caller passing a cancelled context
	// Check ctx.Err() at entry and return early:
	if err := ctx.Err(); err != nil {
		s.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		s.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(ctx, "creating snapshot", slog.String("instanceID", instanceID))

	resp, err := s.api.Post(ctx, fmt.Sprintf("instances/%s/snapshots", instanceID), "")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create snapshot", slog.String("error", err.Error()))
		return nil, err
	}

	var result CreateSnapshotResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		s.logger.ErrorContext(ctx, "failed to unmarshal snapshot response", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(ctx, "snapshot created", slog.String("snapshotId", result.Data.SnapshotID))
	return &result, nil
}

// Restore restores an instance from a snapshot
func (s *snapshotService) Restore(ctx context.Context, instanceID string, snapshotID string) (*RestoreSnapshotResponse, error) {
	// Guard against the caller passing a cancelled context
	// Check ctx.Err() at entry and return early:
	if err := ctx.Err(); err != nil {
		s.logger.ErrorContext(ctx, "context already cancelled before function", slog.String("error", err.Error()))
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		s.logger.ErrorContext(ctx, "invalid instance Id ", slog.String("error", err.Error()))
		return nil, err
	}

	if err := utils.ValidateSnapshotID(snapshotID); err != nil {
		s.logger.ErrorContext(ctx, "invalid snapshot Id ", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(ctx, "restore instance with a snapshot", slog.String("snapshotID", snapshotID), slog.String("instanceID", instanceID))

	resp, err := s.api.Post(ctx, fmt.Sprintf("instances/%s/snapshots/%s/restore", instanceID, snapshotID), "")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to restore using snapshot", slog.String("error", err.Error()))
		return nil, err
	}

	var result RestoreSnapshotResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		s.logger.ErrorContext(ctx, "failed to unmarshal snapshots restore response", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(ctx, "snapshot restore started")
	return &result, nil
}
