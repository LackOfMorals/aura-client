package aura

import "time"

// GetSnapshotsResponse contains a list of snapshots for an instance
type GetSnapshotsResponse struct {
	Data []GetSnapshotData `json:"data"`
}

type GetSnapshotDataResponse struct {
	Data GetSnapshotData `json:"data"`
}

type GetSnapshotData struct {
	InstanceID string `json:"instance_id"`
	SnapshotID string `json:"snapshot_id"`
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
	SnapshotID string `json:"snapshot_id"`
}

// RestoreSnapshotResponse stores the response from initiating restoration of an instance using a snapshot
// The response is the same as for getting instance configuration details
type RestoreSnapshotResponse struct {
	Data InstanceData `json:"data"`
}

// This is ( optionally ) used when listing Snaphots as a filter
type SnapshotDate struct {
	Year  int
	Month time.Month
	Day   int
}

// Return Todays date as *SnapshotDate.  Primarily for use as a filter when listing an instances snapshots
func Today() *SnapshotDate {
	y, m, d := time.Now().Date()
	return &SnapshotDate{y, m, d}
}
