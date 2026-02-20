package aura

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/LackOfMorals/aura-client/internal/utils"
)

// Snapshots

// List returns snapshots for an instance, optionally filtered by date (YYYY-MM-DD)
func (s *snapshotService) List(ctx context.Context, instanceID string, snapshotDate string) (*GetSnapshotsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	s.logger.DebugContext(ctx, "listing snapshots", slog.String("instanceID", instanceID))

	endpoint := "instances/" + instanceID + "/snapshots"

	switch datelen := len(snapshotDate); datelen {
	case 0:
		// empty string, no date filter
		break
	case 10:
		if err := utils.CheckDate(snapshotDate); err != nil {
			return nil, err
		}
		endpoint = endpoint + "?date=" + snapshotDate
	default:
		return nil, fmt.Errorf("date must be in the format of YYYY-MM-DD")
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
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	s.logger.DebugContext(ctx, "get snapshot details", slog.String("snapshotID", snapshotID), slog.String("instanceID", instanceID))

	resp, err := s.api.Get(ctx, "instances/"+instanceID+"/snapshots/"+snapshotID)
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
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	s.logger.DebugContext(ctx, "creating snapshot", slog.String("instanceID", instanceID))

	resp, err := s.api.Post(ctx, "instances/"+instanceID+"/snapshots", "")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create snapshot", slog.String("error", err.Error()))
		return nil, err
	}

	var result CreateSnapshotResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		s.logger.ErrorContext(ctx, "failed to unmarshal snapshot response", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(ctx, "snapshot created", slog.String("snapshotId", result.Data.SnapshotId))
	return &result, nil
}

// Restore restores an instance from a snapshot
func (s *snapshotService) Restore(ctx context.Context, instanceID string, snapshotID string) (*RestoreSnapshotResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	s.logger.DebugContext(ctx, "restore instance with a snapshot", slog.String("snapshotID", snapshotID), slog.String("instanceID", instanceID))

	resp, err := s.api.Post(ctx, "instances/"+instanceID+"/snapshots/"+snapshotID+"/restore", "")
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
