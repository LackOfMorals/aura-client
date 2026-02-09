// interfaces.go
package aura

// TenantService defines operations for managing tenants
type TenantService interface {
	// List returns all tenants accessible to the authenticated user
	List() (*ListTenantsResponse, error)
	// Get retrieves details for a specific tenant by ID
	Get(tenantID string) (*GetTenantResponse, error)
	// GetMetrics gets URL for project level Prometheus metrics
	GetMetrics(tenantID string) (*GetTenantMetricsURLResponse, error)
}

// InstanceService defines operations for managing database instances
type InstanceService interface {
	// List returns all instances accessible to the authenticated user
	List() (*ListInstancesResponse, error)
	// Get retrieves details for a specific instance by ID
	Get(instanceID string) (*GetInstanceResponse, error)
	// Create provisions a new database instance
	Create(instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error)
	// CreateFromStore provisions a new database instance using a stored configuration
	CreateFromStore(label string) (*CreateInstanceResponse, error)
	// Delete removes an instance by ID
	Delete(instanceID string) (*GetInstanceResponse, error)
	// Pause suspends an instance by ID
	Pause(instanceID string) (*GetInstanceResponse, error)
	// Resume restarts a paused instance by ID
	Resume(instanceID string) (*GetInstanceResponse, error)
	// Update modifies an instance's configuration
	Update(instanceID string, instanceRequest *UpdateInstanceData) (*GetInstanceResponse, error)
	// Overwrite replaces instance data from another instance or snapshot
	Overwrite(instanceID string, sourceInstanceID string, sourceSnapshotID string) (*OverwriteInstanceResponse, error)
}

// SnapshotService defines operations for managing instance snapshots
type SnapshotService interface {
	// List returns snapshots for an instance, optionally filtered by date (YYYY-MM-DD)
	List(instanceID string, snapshotDate string) (*GetSnapshotsResponse, error)
	// Create triggers an on-demand snapshot for an instance
	Create(instanceID string) (*CreateSnapshotResponse, error)
	// Gets details for a snapshot of an instance
	Get(instanceID string, snapshotID string) (*GetSnapshotDataResponse, error)
	// Restore instance from a snapshot
	Restore(instanceID string, snapshotID string) (*RestoreSnapshotResponse, error)
}

// CmekService defines operations for customer-managed encryption keys
type CmekService interface {
	// List returns all customer-managed encryption keys, optionally filtered by tenant
	List(tenantID string) (*GetCmeksResponse, error)
}

// GDSSessionService defines operations for Graph Data Science sessions
type GDSSessionService interface {
	// List returns all GDS sessions accessible to the authenticated user
	List() (*GetGDSSessionListResponse, error)
	// Estimate the size of a GDS session
	Estimate(GDSSessionSizeEstimateRequest *GetGDSSessionSizeEstimation) (*GDSSessionSizeEstimationResponse, error)
	// Create a new GDS session
	Create(GDSSessionConfigRequest *CreateGDSSessionConfigData) (*GetGDSSessionResponse, error)
	// Get the details for a single GDS Session
	Get(GDSSessionID string) (*GetGDSSessionResponse, error)
	// Delete a single GDS Session
	Delete(GDSSessionID string) (*DeleteGDSSessionResponse, error)
}

// PrometheusService defines operations for querying Prometheus metrics
type PrometheusService interface {
	// FetchRawMetrics fetches and parses raw Prometheus metrics from an Aura metrics endpoint
	FetchRawMetrics(prometheusURL string) (*PrometheusMetricsResponse, error)
	// GetMetricValue retrieves a specific metric value by name and optional label filters
	GetMetricValue(metrics *PrometheusMetricsResponse, name string, labelFilters map[string]string) (float64, error)
	// GetInstanceHealth retrieves comprehensive health metrics for an instance
	GetInstanceHealth(instanceID string, prometheusURL string) (*PrometheusHealthMetrics, error)
}

// StoreService defines operations for managing instance configuration storage
type StoreService interface {
	// Create stores a new instance configuration with the given label
	Create(label string, config *CreateInstanceConfigData) error
	// Read retrieves an instance configuration by label
	Read(label string) (*CreateInstanceConfigData, error)
	// Update modifies an existing instance configuration
	Update(label string, config *CreateInstanceConfigData) error
	// Delete removes an instance configuration by label
	Delete(label string) error
	// List returns all stored configuration labels
	List() ([]string, error)
}

// Compile-time interface compliance checks
var (
	_ TenantService     = (*tenantService)(nil)
	_ InstanceService   = (*instanceService)(nil)
	_ SnapshotService   = (*snapshotService)(nil)
	_ CmekService       = (*cmekService)(nil)
	_ GDSSessionService = (*gDSSessionService)(nil)
	_ PrometheusService = (*prometheusService)(nil)
	_ StoreService      = (*storeService)(nil)
)
