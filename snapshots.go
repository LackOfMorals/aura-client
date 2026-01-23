package aura

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/LackOfMorals/aura-client/internal/api"
	"github.com/LackOfMorals/aura-client/internal/utils"
)

// Snapshots

// GetSnapshotsResponse contains a list of snapshots for an instance
type GetSnapshotsResponse struct {
	Data []GetSnapshotData `json:"data"`
}

type GetSnapshotData struct {
	InstanceId string `json:"instance_id"`
	SnapshotId string `json:"snapshot_id"`
	Profile    string `json:"profile"`
	Status     string `json:"status"`
	Timestamp  string `json:"timestamp"`
}

// CreateSnapshotResponse contains the result of creating a snapshot
type CreateSnapshotResponse struct {
	Data CreateSnapshotData `json:"data"`
}

type CreateSnapshotData struct {
	SnapshotId string `json:"snapshot_id"`
}

// snapshotService handles snapshot operations
type snapshotService struct {
	api    api.APIRequestService
	ctx    context.Context
	logger *slog.Logger
}

// List returns snapshots for an instance, optionally filtered by date (YYYY-MM-DD)
func (s *snapshotService) List(instanceID string, snapshotDate string) (*GetSnapshotsResponse, error) {
	s.logger.DebugContext(s.ctx, "listing snapshots", slog.String("instanceID", instanceID))

	endpoint := "instances/" + instanceID + "/snapshots"

	switch datelen := len(snapshotDate); datelen {
	case 0:
		// empty string, no date filter
		break
	case 10:
		// Check if date is in correct format
		if err := utils.CheckDate(snapshotDate); err != nil {
			return nil, err
		}
		endpoint = endpoint + "?date=" + snapshotDate
	default:
		return nil, fmt.Errorf("date must be in the format of YYYY-MM-DD")
	}

	resp, err := s.api.Get(s.ctx, endpoint)
	if err != nil {
		s.logger.ErrorContext(s.ctx, "failed to list snapshots", slog.String("error", err.Error()))
		return nil, err
	}

	var result GetSnapshotsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		s.logger.ErrorContext(s.ctx, "failed to unmarshal snapshots response", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(s.ctx, "snapshots listed successfully", slog.Int("count", len(result.Data)))
	return &result, nil
}

// Create triggers an on-demand snapshot for an instance
func (s *snapshotService) Create(instanceID string) (*CreateSnapshotResponse, error) {
	s.logger.DebugContext(s.ctx, "creating snapshot", slog.String("instanceID", instanceID))

	endpoint := "instances/" + instanceID + "/snapshots"

	resp, err := s.api.Post(s.ctx, endpoint, "")
	if err != nil {
		s.logger.ErrorContext(s.ctx, "failed to create snapshot", slog.String("error", err.Error()))
		return nil, err
	}

	var result CreateSnapshotResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		s.logger.ErrorContext(s.ctx, "failed to unmarshal snapshot response", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(s.ctx, "snapshot created", slog.String("snapshotId", result.Data.SnapshotId))
	return &result, nil
}
