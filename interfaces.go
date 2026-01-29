// interfaces.go
package aura

import "time"

// TenantService defines operations for managing tenants
type TenantService interface {
	// List returns all tenants accessible to the authenticated user
	List() (*ListTenantsResponse, error)
	// Get retrieves details for a specific tenant by ID
	Get(tenantID string) (*GetTenantResponse, error)
}

// InstanceService defines operations for managing database instances
type InstanceService interface {
	// List returns all instances accessible to the authenticated user
	List() (*ListInstancesResponse, error)
	// Get retrieves details for a specific instance by ID
	Get(instanceID string) (*GetInstanceResponse, error)
	// Create provisions a new database instance
	Create(instanceRequest *CreateInstanceConfigData) (*CreateInstanceResponse, error)
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
}

// CmekService defines operations for customer-managed encryption keys
type CmekService interface {
	// List returns all customer-managed encryption keys, optionally filtered by tenant
	List(tenantID string) (*GetCmeksResponse, error)
}

// GDSSessionService defines operations for Graph Data Science sessions
type GDSSessionService interface {
	// List returns all GDS sessions accessible to the authenticated user
	List() (*GetGDSSessionResponse, error)
}

// PrometheusService defines operations for querying Prometheus metrics
type PrometheusService interface {
	// Query executes an instant query against a Prometheus endpoint
	Query(prometheusURL string, query string) (*PrometheusQueryResponse, error)
	// QueryRange executes a range query against a Prometheus endpoint
	QueryRange(prometheusURL string, query string, start, end time.Time, step string) (*PrometheusRangeQueryResponse, error)
	// GetInstanceHealth retrieves comprehensive health metrics for an instance
	GetInstanceHealth(instanceID string, prometheusURL string) (*PrometheusHealthMetrics, error)
}

// Compile-time interface compliance checks
var (
	_ TenantService     = (*tenantService)(nil)
	_ InstanceService   = (*instanceService)(nil)
	_ SnapshotService   = (*snapshotService)(nil)
	_ CmekService       = (*cmekService)(nil)
	_ GDSSessionService = (*gDSSessionService)(nil)
	_ PrometheusService = (*prometheusService)(nil)
)
