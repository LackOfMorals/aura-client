package auraAPIClient

import (
	"context"
	"fmt"
	"net/http"
)

// Snaphot methods

// a list of available snapshots for an instance on a ( optional ) given date. If a date is not specified, snapshots from the current day will be returned.
// Date is in ISO format YYYY-MM-DD
func (s *SnapshotService) List(ctx context.Context, token *AuthAPIToken, instanceID string, snapshotDate string) (*GetSnapshotsResponse, error) {

	endpoint := s.service.auraAPIVersion + "/instances/" + instanceID + "/snapshots"

	switch datelen := len(snapshotDate); datelen {

	// empty string
	case 0:
		break
	case 10:
		// Check if date is in correct format
		err := checkDate(snapshotDate)
		if err != nil {
			return nil, err
		}
		endpoint = endpoint + "?date=" + snapshotDate
	default:
		return nil, fmt.Errorf("date must be in the format of YYYY-MM-DD")
	}

	return makeAuthenticatedRequest[GetSnapshotsResponse](ctx, s.service, token, endpoint, http.MethodGet, "application/json", "")
}

// create a snapshot for an instance identified by its Id
func (s *SnapshotService) Create(ctx context.Context, token *AuthAPIToken, instanceID string) (*CreateSnapshotResponse, error) {
	endpoint := s.service.auraAPIVersion + "/instances/" + instanceID + "/snapshots"
	return makeAuthenticatedRequest[CreateSnapshotResponse](ctx, s.service, token, endpoint, http.MethodPost, "application/json", "")
}
