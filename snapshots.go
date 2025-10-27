package aura

import (
	"context"
	"fmt"
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
	Service *AuraAPIActionsService
}

// Snaphot methods

// a list of available snapshots for an instance on a ( optional ) given date. If a date is not specified, snapshots from the current day will be returned.
// Date is in ISO format YYYY-MM-DD
func (s *SnapshotService) List(ctx context.Context, instanceID string, snapshotDate string) (*GetSnapshotsResponse, error) {
	// Get or update token if needed
	err := s.Service.authMgr.getToken(ctx, *s.Service.transport)
	if err != nil { // Token process failed
		return nil, err
	}

	content := "application/json"
	auth := s.Service.authMgr.Type + " " + s.Service.authMgr.Token
	endpoint := s.Service.Config.Version + "/instances/" + instanceID + "/snapshots"

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

	return makeAuthenticatedRequest[GetSnapshotsResponse](ctx, *s.Service.transport, auth, endpoint, http.MethodGet, content, "")
}

// create a snapshot for an instance identified by its Id
func (s *SnapshotService) Create(ctx context.Context, instanceID string) (*CreateSnapshotResponse, error) {
	// Get or update token if needed
	err := s.Service.authMgr.getToken(ctx, *s.Service.transport)
	if err != nil { // Token process failed
		return nil, err
	}

	content := "application/json"
	auth := s.Service.authMgr.Type + " " + s.Service.authMgr.Token
	endpoint := s.Service.Config.Version + "/instances/" + instanceID + "/snapshots"

	return makeAuthenticatedRequest[CreateSnapshotResponse](ctx, *s.Service.transport, auth, endpoint, http.MethodPost, content, "")
}
