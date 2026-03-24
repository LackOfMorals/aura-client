package aura

// instance status values as defined in the Aura API specification
type InstanceStatus string

const (
	StatusRunning       InstanceStatus = "running"
	StatusStopped       InstanceStatus = "paused"
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
	StatusRestroying    InstanceStatus = "restoring"
	StatusUpdating      InstanceStatus = "updating"
	StatusOverwriting   InstanceStatus = "overwriting"
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
