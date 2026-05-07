package auratest_test

import (
	"context"
	"errors"
	"testing"

	aura "github.com/LackOfMorals/aura-client"
	"github.com/LackOfMorals/aura-client/auratest"
)

// Compile-time checks: every fake must satisfy its interface.
var (
	_ aura.TenantService     = (*auratest.FakeTenantService)(nil)
	_ aura.InstanceService   = (*auratest.FakeInstanceService)(nil)
	_ aura.SnapshotService   = (*auratest.FakeSnapshotService)(nil)
	_ aura.CmekService       = (*auratest.FakeCmekService)(nil)
	_ aura.GDSSessionService = (*auratest.FakeGDSSessionService)(nil)
	_ aura.PrometheusService = (*auratest.FakePrometheusService)(nil)
)

func TestFake_NilFunc_ReturnsZeroValue(t *testing.T) {
	ctx := context.Background()

	t.Run("FakeTenantService", func(t *testing.T) {
		svc := &auratest.FakeTenantService{}
		if _, err := svc.List(ctx); err != nil {
			t.Errorf("List: %v", err)
		}
		if _, err := svc.Get(ctx, "tid"); err != nil {
			t.Errorf("Get: %v", err)
		}
		if _, err := svc.GetMetrics(ctx, "tid"); err != nil {
			t.Errorf("GetMetrics: %v", err)
		}
	})

	t.Run("FakeInstanceService", func(t *testing.T) {
		svc := &auratest.FakeInstanceService{}
		if _, err := svc.List(ctx); err != nil {
			t.Errorf("List: %v", err)
		}
		if _, err := svc.Get(ctx, "iid"); err != nil {
			t.Errorf("Get: %v", err)
		}
	})

	t.Run("FakeSnapshotService", func(t *testing.T) {
		svc := &auratest.FakeSnapshotService{}
		if _, err := svc.List(ctx, "iid", nil); err != nil {
			t.Errorf("List: %v", err)
		}
	})

	t.Run("FakeCmekService", func(t *testing.T) {
		svc := &auratest.FakeCmekService{}
		if _, err := svc.List(ctx, ""); err != nil {
			t.Errorf("List: %v", err)
		}
	})

	t.Run("FakeGDSSessionService", func(t *testing.T) {
		svc := &auratest.FakeGDSSessionService{}
		if _, err := svc.List(ctx); err != nil {
			t.Errorf("List: %v", err)
		}
	})

	t.Run("FakePrometheusService", func(t *testing.T) {
		svc := &auratest.FakePrometheusService{}
		if _, err := svc.FetchRawMetrics(ctx, "http://example.com"); err != nil {
			t.Errorf("FetchRawMetrics: %v", err)
		}
	})
}

func TestFake_CustomFunc_IsCalled(t *testing.T) {
	ctx := context.Background()
	sentinel := errors.New("injected")

	svc := &auratest.FakeInstanceService{
		ListFunc: func(_ context.Context) (*aura.ListInstancesResponse, error) {
			return nil, sentinel
		},
	}
	_, err := svc.List(ctx)
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
