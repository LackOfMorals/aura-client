package aura

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/LackOfMorals/aura-client/internal/utils"
)

// Snapshots
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

type GetSnapshotResponse struct {
	Data GetSnapshotData `json:"data"`
}

type CreateSnapshotResponse struct {
	Data CreateSnapshotData `json:"data"`
}

type CreateSnapshotData struct {
	SnapshotId string `json:"snapshot_id"`
}

// SnapshotService handles snapshot operations
type SnapshotService struct {
	service *AuraAPIClient
	logger  *slog.Logger
}

// Snaphot methods

// a list of available snapshots for an instance on a ( optional ) given date. If a date is not specified, snapshots from the current day will be returned.
// Date is in ISO format YYYY-MM-DD
func (s *SnapshotService) List(ctx context.Context, instanceID string, snapshotDate string) (*GetSnapshotsResponse, error) {
	s.logger.DebugContext(ctx, "listing snapshots")

	// Get or update token if needed
	err := s.service.authMgr.getToken(ctx, *s.service.transport)
	if err != nil { // Token process failed
		s.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := s.service.authMgr.tokenType + " " + s.service.authMgr.token
	endpoint := s.service.config.version + "/instances/" + instanceID + "/snapshots"

	switch datelen := len(snapshotDate); datelen {

	// empty string
	case 0:
		break
	case 10:
		// Check if date is in correct format
		err := utils.CheckDate(snapshotDate)
		if err != nil {
			return nil, err
		}
		endpoint = endpoint + "?date=" + snapshotDate
	default:
		return nil, fmt.Errorf("date must be in the format of YYYY-MM-DD")
	}

	s.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[GetSnapshotsResponse](ctx, *s.service.transport, auth, endpoint, http.MethodGet, content, "")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list snapshots", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(ctx, "snapshots listed  successfully", slog.Int("count", len(resp.Data)))
	return resp, nil

}

// create a snapshot for an instance identified by its Id
func (s *SnapshotService) Create(ctx context.Context, instanceID string) (*CreateSnapshotResponse, error) {
	s.logger.DebugContext(ctx, "creating snapshot")

	// Get or update token if needed
	err := s.service.authMgr.getToken(ctx, *s.service.transport)
	if err != nil { // Token process failed
		s.logger.ErrorContext(ctx, "failed to obtain authentication token", slog.String("error", err.Error()))
		return nil, err
	}

	content := "application/json"
	auth := s.service.authMgr.tokenType + " " + s.service.authMgr.token
	endpoint := s.service.config.version + "/instances/" + instanceID + "/snapshots"

	s.logger.DebugContext(ctx, "making authenticated request",
		slog.String("method", http.MethodGet),
		slog.String("endpoint", endpoint),
	)

	resp, err := makeAuthenticatedRequest[CreateSnapshotResponse](ctx, *s.service.transport, auth, endpoint, http.MethodPost, content, "")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create snapshot", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(ctx, "snapshot creating", slog.String("snapshost Id", resp.Data.SnapshotId))
	return resp, nil

}
