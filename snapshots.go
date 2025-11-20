package aura

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/LackOfMorals/aura-client/internal/utils"
)

// Snapshots
type getSnapshotsResponse struct {
	Data []getSnapshotData `json:"data"`
}

type getSnapshotData struct {
	InstanceId string `json:"instance_id"`
	SnapshotId string `json:"snapshot_id"`
	Profile    string `json:"profile"`
	Status     string `json:"status"`
	Timestamp  string `json:"timestamp"`
}

type getSnapshotResponse struct {
	Data getSnapshotData `json:"data"`
}

type createSnapshotResponse struct {
	Data createSnapshotData `json:"data"`
}

type createSnapshotData struct {
	SnapshotId string `json:"snapshot_id"`
}

// snapshotService handles snapshot operations
type snapshotService struct {
	service *AuraAPIClient
	logger  *slog.Logger
}

// Snapshot methods

// a list of available snapshots for an instance on a ( optional ) given date. If a date is not specified, snapshots from the current day will be returned.
// Date is in ISO format YYYY-MM-DD
func (s *snapshotService) List(instanceID string, snapshotDate string) (*getSnapshotsResponse, error) {
	s.logger.DebugContext(s.service.config.ctx, "listing snapshots")

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

	resp, err := makeServiceRequest[getSnapshotsResponse](s.service.config.ctx, *s.service.transport, s.service.authMgr, endpoint, http.MethodGet, "", s.logger)
	if err != nil {
		s.logger.ErrorContext(s.service.config.ctx, "failed to list snapshots", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(s.service.config.ctx, "snapshots listed  successfully", slog.Int("count", len(resp.Data)))
	return resp, nil

}

// create a snapshot for an instance identified by its Id
func (s *snapshotService) Create(instanceID string) (*createSnapshotResponse, error) {
	s.logger.DebugContext(s.service.config.ctx, "creating snapshot")

	endpoint := s.service.config.version + "/instances/" + instanceID + "/snapshots"

	resp, err := makeServiceRequest[createSnapshotResponse](s.service.config.ctx, *s.service.transport, s.service.authMgr, endpoint, http.MethodPost, "", s.logger)
	if err != nil {
		s.logger.ErrorContext(s.service.config.ctx, "failed to create snapshot", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.DebugContext(s.service.config.ctx, "creating snapshot", slog.String("snapshost Id", resp.Data.SnapshotId))
	return resp, nil

}
