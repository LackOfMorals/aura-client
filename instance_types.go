package aura

import "fmt"

// instance status values as defined in the Aura API specification
type InstanceStatus string

const (
	StatusRunning       InstanceStatus = "running"
	StatusPaused        InstanceStatus = "paused"
	StatusAvailable     InstanceStatus = "available"
	StatusCreating      InstanceStatus = "creating"
	StatusDestroying    InstanceStatus = "destroying"
	StatusPausing       InstanceStatus = "pausing"
	StatusSuspending    InstanceStatus = "suspending"
	StatusSuspended     InstanceStatus = "suspended"
	StatusResuming      InstanceStatus = "resuming"
	StatusLoading       InstanceStatus = "loading"
	StatusLoadingFailed InstanceStatus = "loading failed"
	StatusRestoring     InstanceStatus = "restoring"
	StatusUpdating      InstanceStatus = "updating"
	StatusOverwriting   InstanceStatus = "overwriting"

	// Deprecated: StatusRestroying was a misspelling. Use StatusRestoring.
	StatusRestroying = StatusRestoring
)

// ListInstancesResponse contains a list of instances in a tenant
type ListInstancesResponse struct {
	Data []ListInstanceData `json:"data"`
}

type ListInstanceData struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Created       string `json:"created_at"`
	TenantID      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
}

type CreateInstanceConfigData struct {
	Name          string `json:"name"`
	TenantID      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
	Region        string `json:"region"`
	Type          string `json:"type"`
	Version       string `json:"version,omitempty"`
	Memory        string `json:"memory"`
}

type CreateInstanceResponse struct {
	Data CreateInstanceData `json:"data"`
}

// CreateInstanceData holds the response fields for a newly provisioned instance.
// It contains the database password returned by the API — treat this value as a
// secret and avoid logging or serialising the struct directly. The String()
// method redacts the password for safe use in log output.
type CreateInstanceData struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	TenantID      string `json:"tenant_id"`
	CloudProvider string `json:"cloud_provider"`
	ConnectionURL string `json:"connection_url"`
	Region        string `json:"region"`
	Type          string `json:"type"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

// String implements fmt.Stringer and redacts the Password field so that
// accidentally logging or printing this struct never exposes credentials.
func (c CreateInstanceData) String() string {
	return fmt.Sprintf(
		"CreateInstanceData{ID:%s Name:%s TenantID:%s CloudProvider:%s Region:%s Type:%s Username:%s Password:[redacted]}",
		c.ID, c.Name, c.TenantID, c.CloudProvider, c.Region, c.Type, c.Username,
	)
}

type UpdateInstanceData struct {
	Name   string `json:"name,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type GetInstanceResponse struct {
	Data InstanceData `json:"data"`
}

type DeleteInstanceResponse struct {
	Data InstanceData `json:"data"`
}

type InstanceData struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Status          InstanceStatus `json:"status"`
	TenantID        string         `json:"tenant_id"`
	CloudProvider   string         `json:"cloud_provider"`
	ConnectionURL   string         `json:"connection_url"`
	Region          string         `json:"region"`
	Type            string         `json:"type"`
	Memory          string         `json:"memory"`
	Storage         *string        `json:"storage"`
	CDCEnrichment   string         `json:"cdc_enrichment_mode"`
	GDSPlugin       bool           `json:"graph_analytics_plugin"`
	MetricsURL      string         `json:"metrics_integration_url"`
	Secondaries     int            `json:"secondaries_count"`
	VectorOptimized bool           `json:"vector_optimized"`
}

type overwriteInstanceRequest struct {
	SourceInstanceID string `json:"source_instance_id,omitempty"`
	SourceSnapshotID string `json:"source_snapshot_id,omitempty"`
}

type OverwriteInstanceResponse struct {
	Data string `json:"data"`
}
