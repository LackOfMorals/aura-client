// types.go - Exported response types for use by consumers
package aura

// ============================================================================
// Tenant Types
// ============================================================================

// ListTenantsResponse contains a list of tenants
type ListTenantsResponse = listTenantsResponse

// TenantsResponseData contains tenant summary information
type TenantsResponseData = tenantsResponseData

// GetTenantResponse contains detailed tenant information
type GetTenantResponse = getTenantResponse

// TenantResponseData contains tenant details including instance configurations
// Note: The underlying type has a typo (tenantReponseData) which is preserved for compatibility
type TenantResponseData = tenantReponseData

// TenantInstanceConfiguration describes available instance configurations for a tenant
type TenantInstanceConfiguration = tenantInstanceConfiguration

// ============================================================================
// Instance Types
// ============================================================================

// ListInstancesResponse contains a list of instances
type ListInstancesResponse = listInstancesResponse

// ListInstanceData contains instance summary information
type ListInstanceData = listInstanceData

// GetInstanceResponse contains detailed instance information
type GetInstanceResponse = getInstanceResponse

// GetInstanceData contains full instance details
type GetInstanceData = getInstanceData

// CreateInstanceResponse contains the response from creating an instance
type CreateInstanceResponse = createInstanceResponse

// OverwriteInstanceResponse contains the response from an overwrite operation
type OverwriteInstanceResponse = overwriteInstanceResponse

// ============================================================================
// Snapshot Types
// ============================================================================

// GetSnapshotsResponse contains a list of snapshots
type GetSnapshotsResponse = getSnapshotsResponse

// GetSnapshotData contains snapshot details
type GetSnapshotData = getSnapshotData

// GetSnapshotResponse contains a single snapshot
type GetSnapshotResponse = getSnapshotResponse

// CreateSnapshotResponse contains the response from creating a snapshot
type CreateSnapshotResponse = createSnapshotResponse

// CreateSnapshotData contains the created snapshot ID
type CreateSnapshotData = createSnapshotData

// ============================================================================
// CMEK Types
// ============================================================================

// GetCmeksResponse contains a list of customer-managed encryption keys
type GetCmeksResponse = getCmeksResponse

// GetCmeksData contains CMEK details
type GetCmeksData = getCmeksData

// ============================================================================
// GDS Session Types
// ============================================================================

// GetGDSSessionResponse contains a list of GDS sessions
type GetGDSSessionResponse = getGDSSessionResponse

// GetGDSSessionData contains GDS session details
type GetGDSSessionData = getGDSSessionData
