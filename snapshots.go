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

// GetSnapshotsResponse contains a list of snapshots for an instance
type GetSnapshotsResponse struct {
	Data []GetSnapshotData `json:"data"`
}

type GetSnapshotDataResponse struct {
	Data GetSnapshotData `json:"data"`
}

type GetSnapshotData struct {
	InstanceId string `json:"instance_id"`
	SnapshotId string `json:"snapshot_id"`
	Profile    string `json:"profile"`
	Status     string `json:"status"`
	Timestamp  string `json:"timestamp"`
	Exportable bool   `json:"exportable"`
}

// CreateSnapshotResponse contains the result of creating a snapshot
type CreateSnapshotResponse struct {
	Data CreateSnapshotData `json:"data"`
}

type CreateSnapshotData struct {
	SnapshotId string `json:"snapshot_id"`
}

// RestoreSnapshotResponse stores the response from initiating restoration of an instance using a snapshot
// The response is the same as for getting instance configuration details
type RestoreSnapshotResponse struct {
	Data GetInstanceData `json:"data"`
}

// snapshotService handles snapshot operations
type snapshotService struct {
	api     api.RequestService
	ctx     context.Context
	timeout time.Duration
	logger  *slog.Logger
}

// List returns snapshots for an instance, optionally filtered by date (YYYY-MM-DD)
func (s *snapshotService) List(instanceID string, snapshotDate string) (*GetSnapshotsResponse, error) {
	// Create child context with timeout for this operation
	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	s.logger.DebugContext(ctx, "listing snapshots", slog.String("instanceID", instanceID))

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

// Get returns the details for a snapshot of an instance, identified by a snapshot Id and instance Id
func (s *snapshotService) Get(instanceID string, snapshotID string) (*GetSnapshotDataResponse, error) {
	// Create child context with timeout for this operation
	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	s.logger.DebugContext(ctx, "get snapshot details", slog.String("snapshotID", snapshotID), slog.String("instanceID", instanceID))

	endpoint := "instances/" + instanceID + "/snapshots/" + snapshotID

	resp, err := s.api.Get(ctx, endpoint)
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
func (s *snapshotService) Create(instanceID string) (*CreateSnapshotResponse, error) {
	// Create child context with timeout for this operation
	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	s.logger.DebugContext(ctx, "creating snapshot", slog.String("instanceID", instanceID))

	endpoint := "instances/" + instanceID + "/snapshots"

	resp, err := s.api.Post(ctx, endpoint, "")
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

// Restore an instance with a snapshot.
func (s *snapshotService) Restore(instanceID string, snapshotID string) (*RestoreSnapshotResponse, error) {
	// Create child context with timeout for this operation
	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	s.logger.DebugContext(ctx, "restore instance with a snapshot", slog.String("snapshotID", snapshotID), slog.String("instanceID", instanceID))

	endpoint := "instances/" + instanceID + "/snapshots/" + snapshotID + "/restore"

	resp, err := s.api.Post(ctx, endpoint, "")
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
