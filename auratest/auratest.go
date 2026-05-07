// Package auratest provides test doubles for the aura client's service
// interfaces. Import it in your own tests to avoid hitting the real Aura API.
//
// Each fake has one exported function field per interface method. When a field
// is nil the fake returns a zero-value response and a nil error, which is the
// most useful default for tests that only care about a subset of the API.
//
// Example:
//
//	svc := &auratest.FakeInstanceService{
//	    ListFunc: func(ctx context.Context) (*aura.ListInstancesResponse, error) {
//	        return &aura.ListInstancesResponse{Data: []aura.ListInstanceData{{ID: "abc"}}}, nil
//	    },
//	}
//	// Use svc wherever an aura.InstanceService is expected.
package auratest

import (
	"context"

	aura "github.com/LackOfMorals/aura-client"
)

// FakeTenantService is a configurable test double for aura.TenantService.
type FakeTenantService struct {
	ListFunc       func(ctx context.Context) (*aura.ListTenantsResponse, error)
	GetFunc        func(ctx context.Context, tenantID string) (*aura.GetTenantResponse, error)
	GetMetricsFunc func(ctx context.Context, tenantID string) (*aura.GetTenantMetricsURLResponse, error)
}

// List calls ListFunc if set, otherwise returns a zero-value response.
func (f *FakeTenantService) List(ctx context.Context) (*aura.ListTenantsResponse, error) {
	if f.ListFunc != nil {
		return f.ListFunc(ctx)
	}
	return &aura.ListTenantsResponse{}, nil
}

// Get calls GetFunc if set, otherwise returns a zero-value response.
func (f *FakeTenantService) Get(ctx context.Context, tenantID string) (*aura.GetTenantResponse, error) {
	if f.GetFunc != nil {
		return f.GetFunc(ctx, tenantID)
	}
	return &aura.GetTenantResponse{}, nil
}

// GetMetrics calls GetMetricsFunc if set, otherwise returns a zero-value response.
func (f *FakeTenantService) GetMetrics(ctx context.Context, tenantID string) (*aura.GetTenantMetricsURLResponse, error) {
	if f.GetMetricsFunc != nil {
		return f.GetMetricsFunc(ctx, tenantID)
	}
	return &aura.GetTenantMetricsURLResponse{}, nil
}

// FakeInstanceService is a configurable test double for aura.InstanceService.
type FakeInstanceService struct {
	ListFunc                  func(ctx context.Context) (*aura.ListInstancesResponse, error)
	GetFunc                   func(ctx context.Context, instanceID string) (*aura.GetInstanceResponse, error)
	CreateFunc                func(ctx context.Context, req *aura.CreateInstanceConfigData) (*aura.CreateInstanceResponse, error)
	DeleteFunc                func(ctx context.Context, instanceID string) (*aura.DeleteInstanceResponse, error)
	PauseFunc                 func(ctx context.Context, instanceID string) (*aura.GetInstanceResponse, error)
	ResumeFunc                func(ctx context.Context, instanceID string) (*aura.GetInstanceResponse, error)
	UpdateFunc                func(ctx context.Context, instanceID string, req *aura.UpdateInstanceData) (*aura.GetInstanceResponse, error)
	OverwriteFromInstanceFunc func(ctx context.Context, instanceID string, sourceInstanceID string) (*aura.OverwriteInstanceResponse, error)
	OverwriteFromSnapshotFunc func(ctx context.Context, instanceID string, sourceSnapshotID string) (*aura.OverwriteInstanceResponse, error)
}

// List calls ListFunc if set, otherwise returns a zero-value response.
func (f *FakeInstanceService) List(ctx context.Context) (*aura.ListInstancesResponse, error) {
	if f.ListFunc != nil {
		return f.ListFunc(ctx)
	}
	return &aura.ListInstancesResponse{}, nil
}

// Get calls GetFunc if set, otherwise returns a zero-value response.
func (f *FakeInstanceService) Get(ctx context.Context, instanceID string) (*aura.GetInstanceResponse, error) {
	if f.GetFunc != nil {
		return f.GetFunc(ctx, instanceID)
	}
	return &aura.GetInstanceResponse{}, nil
}

// Create calls CreateFunc if set, otherwise returns a zero-value response.
func (f *FakeInstanceService) Create(ctx context.Context, req *aura.CreateInstanceConfigData) (*aura.CreateInstanceResponse, error) {
	if f.CreateFunc != nil {
		return f.CreateFunc(ctx, req)
	}
	return &aura.CreateInstanceResponse{}, nil
}

// Delete calls DeleteFunc if set, otherwise returns a zero-value response.
func (f *FakeInstanceService) Delete(ctx context.Context, instanceID string) (*aura.DeleteInstanceResponse, error) {
	if f.DeleteFunc != nil {
		return f.DeleteFunc(ctx, instanceID)
	}
	return &aura.DeleteInstanceResponse{}, nil
}

// Pause calls PauseFunc if set, otherwise returns a zero-value response.
func (f *FakeInstanceService) Pause(ctx context.Context, instanceID string) (*aura.GetInstanceResponse, error) {
	if f.PauseFunc != nil {
		return f.PauseFunc(ctx, instanceID)
	}
	return &aura.GetInstanceResponse{}, nil
}

// Resume calls ResumeFunc if set, otherwise returns a zero-value response.
func (f *FakeInstanceService) Resume(ctx context.Context, instanceID string) (*aura.GetInstanceResponse, error) {
	if f.ResumeFunc != nil {
		return f.ResumeFunc(ctx, instanceID)
	}
	return &aura.GetInstanceResponse{}, nil
}

// Update calls UpdateFunc if set, otherwise returns a zero-value response.
func (f *FakeInstanceService) Update(ctx context.Context, instanceID string, req *aura.UpdateInstanceData) (*aura.GetInstanceResponse, error) {
	if f.UpdateFunc != nil {
		return f.UpdateFunc(ctx, instanceID, req)
	}
	return &aura.GetInstanceResponse{}, nil
}

// OverwriteFromInstance calls OverwriteFromInstanceFunc if set, otherwise returns a zero-value response.
func (f *FakeInstanceService) OverwriteFromInstance(ctx context.Context, instanceID string, sourceInstanceID string) (*aura.OverwriteInstanceResponse, error) {
	if f.OverwriteFromInstanceFunc != nil {
		return f.OverwriteFromInstanceFunc(ctx, instanceID, sourceInstanceID)
	}
	return &aura.OverwriteInstanceResponse{}, nil
}

// OverwriteFromSnapshot calls OverwriteFromSnapshotFunc if set, otherwise returns a zero-value response.
func (f *FakeInstanceService) OverwriteFromSnapshot(ctx context.Context, instanceID string, sourceSnapshotID string) (*aura.OverwriteInstanceResponse, error) {
	if f.OverwriteFromSnapshotFunc != nil {
		return f.OverwriteFromSnapshotFunc(ctx, instanceID, sourceSnapshotID)
	}
	return &aura.OverwriteInstanceResponse{}, nil
}

// FakeSnapshotService is a configurable test double for aura.SnapshotService.
type FakeSnapshotService struct {
	ListFunc    func(ctx context.Context, instanceID string, date *aura.SnapshotDate) (*aura.GetSnapshotsResponse, error)
	CreateFunc  func(ctx context.Context, instanceID string) (*aura.CreateSnapshotResponse, error)
	GetFunc     func(ctx context.Context, instanceID string, snapshotID string) (*aura.GetSnapshotDataResponse, error)
	RestoreFunc func(ctx context.Context, instanceID string, snapshotID string) (*aura.RestoreSnapshotResponse, error)
}

// List calls ListFunc if set, otherwise returns a zero-value response.
func (f *FakeSnapshotService) List(ctx context.Context, instanceID string, date *aura.SnapshotDate) (*aura.GetSnapshotsResponse, error) {
	if f.ListFunc != nil {
		return f.ListFunc(ctx, instanceID, date)
	}
	return &aura.GetSnapshotsResponse{}, nil
}

// Create calls CreateFunc if set, otherwise returns a zero-value response.
func (f *FakeSnapshotService) Create(ctx context.Context, instanceID string) (*aura.CreateSnapshotResponse, error) {
	if f.CreateFunc != nil {
		return f.CreateFunc(ctx, instanceID)
	}
	return &aura.CreateSnapshotResponse{}, nil
}

// Get calls GetFunc if set, otherwise returns a zero-value response.
func (f *FakeSnapshotService) Get(ctx context.Context, instanceID string, snapshotID string) (*aura.GetSnapshotDataResponse, error) {
	if f.GetFunc != nil {
		return f.GetFunc(ctx, instanceID, snapshotID)
	}
	return &aura.GetSnapshotDataResponse{}, nil
}

// Restore calls RestoreFunc if set, otherwise returns a zero-value response.
func (f *FakeSnapshotService) Restore(ctx context.Context, instanceID string, snapshotID string) (*aura.RestoreSnapshotResponse, error) {
	if f.RestoreFunc != nil {
		return f.RestoreFunc(ctx, instanceID, snapshotID)
	}
	return &aura.RestoreSnapshotResponse{}, nil
}

// FakeCmekService is a configurable test double for aura.CmekService.
type FakeCmekService struct {
	ListFunc func(ctx context.Context, tenantID string) (*aura.GetCmeksResponse, error)
}

// List calls ListFunc if set, otherwise returns a zero-value response.
func (f *FakeCmekService) List(ctx context.Context, tenantID string) (*aura.GetCmeksResponse, error) {
	if f.ListFunc != nil {
		return f.ListFunc(ctx, tenantID)
	}
	return &aura.GetCmeksResponse{}, nil
}

// FakeGDSSessionService is a configurable test double for aura.GDSSessionService.
type FakeGDSSessionService struct {
	ListFunc     func(ctx context.Context) (*aura.GetGDSSessionListResponse, error)
	EstimateFunc func(ctx context.Context, req *aura.GetGDSSessionSizeEstimation) (*aura.GDSSessionSizeEstimationResponse, error)
	CreateFunc   func(ctx context.Context, req *aura.CreateGDSSessionConfigData) (*aura.GetGDSSessionResponse, error)
	GetFunc      func(ctx context.Context, sessionID string) (*aura.GetGDSSessionResponse, error)
	DeleteFunc   func(ctx context.Context, sessionID string) (*aura.DeleteGDSSessionResponse, error)
}

// List calls ListFunc if set, otherwise returns a zero-value response.
func (f *FakeGDSSessionService) List(ctx context.Context) (*aura.GetGDSSessionListResponse, error) {
	if f.ListFunc != nil {
		return f.ListFunc(ctx)
	}
	return &aura.GetGDSSessionListResponse{}, nil
}

// Estimate calls EstimateFunc if set, otherwise returns a zero-value response.
func (f *FakeGDSSessionService) Estimate(ctx context.Context, req *aura.GetGDSSessionSizeEstimation) (*aura.GDSSessionSizeEstimationResponse, error) {
	if f.EstimateFunc != nil {
		return f.EstimateFunc(ctx, req)
	}
	return &aura.GDSSessionSizeEstimationResponse{}, nil
}

// Create calls CreateFunc if set, otherwise returns a zero-value response.
func (f *FakeGDSSessionService) Create(ctx context.Context, req *aura.CreateGDSSessionConfigData) (*aura.GetGDSSessionResponse, error) {
	if f.CreateFunc != nil {
		return f.CreateFunc(ctx, req)
	}
	return &aura.GetGDSSessionResponse{}, nil
}

// Get calls GetFunc if set, otherwise returns a zero-value response.
func (f *FakeGDSSessionService) Get(ctx context.Context, sessionID string) (*aura.GetGDSSessionResponse, error) {
	if f.GetFunc != nil {
		return f.GetFunc(ctx, sessionID)
	}
	return &aura.GetGDSSessionResponse{}, nil
}

// Delete calls DeleteFunc if set, otherwise returns a zero-value response.
func (f *FakeGDSSessionService) Delete(ctx context.Context, sessionID string) (*aura.DeleteGDSSessionResponse, error) {
	if f.DeleteFunc != nil {
		return f.DeleteFunc(ctx, sessionID)
	}
	return &aura.DeleteGDSSessionResponse{}, nil
}

// FakePrometheusService is a configurable test double for aura.PrometheusService.
type FakePrometheusService struct {
	FetchRawMetricsFunc   func(ctx context.Context, prometheusURL string) (*aura.PrometheusMetricsResponse, error)
	GetMetricValueFunc    func(ctx context.Context, metrics *aura.PrometheusMetricsResponse, name string, labelFilters map[string]string) (float64, error)
	GetInstanceHealthFunc func(ctx context.Context, instanceID string, prometheusURL string) (*aura.PrometheusHealthMetrics, error)
}

// FetchRawMetrics calls FetchRawMetricsFunc if set, otherwise returns an empty metrics response.
func (f *FakePrometheusService) FetchRawMetrics(ctx context.Context, prometheusURL string) (*aura.PrometheusMetricsResponse, error) {
	if f.FetchRawMetricsFunc != nil {
		return f.FetchRawMetricsFunc(ctx, prometheusURL)
	}
	return &aura.PrometheusMetricsResponse{Metrics: map[string][]aura.PrometheusMetric{}}, nil
}

// GetMetricValue calls GetMetricValueFunc if set, otherwise returns 0.
func (f *FakePrometheusService) GetMetricValue(ctx context.Context, metrics *aura.PrometheusMetricsResponse, name string, labelFilters map[string]string) (float64, error) {
	if f.GetMetricValueFunc != nil {
		return f.GetMetricValueFunc(ctx, metrics, name, labelFilters)
	}
	return 0, nil
}

// GetInstanceHealth calls GetInstanceHealthFunc if set, otherwise returns a zero-value response.
func (f *FakePrometheusService) GetInstanceHealth(ctx context.Context, instanceID string, prometheusURL string) (*aura.PrometheusHealthMetrics, error) {
	if f.GetInstanceHealthFunc != nil {
		return f.GetInstanceHealthFunc(ctx, instanceID, prometheusURL)
	}
	return &aura.PrometheusHealthMetrics{}, nil
}
