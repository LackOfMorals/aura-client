package aura

import (
	"log/slog"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

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
	Data InstanceData `json:"data"`
}

// snapshotService handles snapshot operations
type snapshotService struct {
	api     api.RequestService
	timeout time.Duration
	logger  *slog.Logger
}
