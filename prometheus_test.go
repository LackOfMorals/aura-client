package aura

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
	"github.com/LackOfMorals/aura-client/internal/httpClient"
)

func TestPrometheusService_Query(t *testing.T) {
	// Create a prometheus service
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	httpSvc := httpClient.NewHTTPService("https://api.neo4j.io/", 30*time.Second, 3, logger)
	apiSvc := api.NewAPIRequestService(httpSvc, api.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		APIVersion:   "v1",
		Timeout:      30 * time.Second,
	}, logger)
	
	promSvc := &prometheusService{
		api:    apiSvc,
		ctx:    context.Background(),
		logger: logger,
	}

	// Test with invalid URL (should fail gracefully)
	t.Run("EmptyURL", func(t *testing.T) {
		_, err := promSvc.Query("", "up")
		if err == nil {
			t.Error("Expected error for empty URL, got nil")
		}
	})

	// Note: Testing actual Prometheus queries would require a real instance
	// and valid credentials, which is typically done in integration tests
}

func TestPrometheusService_QueryRange(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	httpSvc := httpClient.NewHTTPService("https://api.neo4j.io/", 30*time.Second, 3, logger)
	apiSvc := api.NewAPIRequestService(httpSvc, api.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		APIVersion:   "v1",
		Timeout:      30 * time.Second,
	}, logger)
	
	promSvc := &prometheusService{
		api:    apiSvc,
		ctx:    context.Background(),
		logger: logger,
	}

	t.Run("EmptyURL", func(t *testing.T) {
		start := time.Now().Add(-1 * time.Hour)
		end := time.Now()
		_, err := promSvc.QueryRange("", "up", start, end, "1m")
		if err == nil {
			t.Error("Expected error for empty URL, got nil")
		}
	})
}

func TestPrometheusService_GetInstanceHealth(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	httpSvc := httpClient.NewHTTPService("https://api.neo4j.io/", 30*time.Second, 3, logger)
	apiSvc := api.NewAPIRequestService(httpSvc, api.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		APIVersion:   "v1",
		Timeout:      30 * time.Second,
	}, logger)
	
	promSvc := &prometheusService{
		api:    apiSvc,
		ctx:    context.Background(),
		logger: logger,
	}

	t.Run("InvalidInstanceID", func(t *testing.T) {
		_, err := promSvc.GetInstanceHealth("", "https://example.com/prometheus")
		if err == nil {
			t.Error("Expected error for empty instance ID, got nil")
		}
	})

	t.Run("EmptyPrometheusURL", func(t *testing.T) {
		_, err := promSvc.GetInstanceHealth("abc123", "")
		if err == nil {
			t.Error("Expected error for empty Prometheus URL, got nil")
		}
	})
}

func TestAssessHealth(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	httpSvc := httpClient.NewHTTPService("https://api.neo4j.io/", 30*time.Second, 3, logger)
	apiSvc := api.NewAPIRequestService(httpSvc, api.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		APIVersion:   "v1",
		Timeout:      30 * time.Second,
	}, logger)
	
	promSvc := &prometheusService{
		api:    apiSvc,
		ctx:    context.Background(),
		logger: logger,
	}

	tests := []struct {
		name           string
		metrics        *PrometheusHealthMetrics
		expectedStatus string
		expectIssues   bool
	}{
		{
			name: "Healthy system",
			metrics: &PrometheusHealthMetrics{
				Resources: ResourceMetrics{
					CPUUsagePercent:    50,
					MemoryUsagePercent: 60,
				},
				Connections: ConnectionMetrics{
					ActiveConnections: 30,
					MaxConnections:    100,
					UsagePercent:      30,
				},
				Storage: StorageMetrics{
					PageCacheHitRate: 90,
				},
			},
			expectedStatus: "healthy",
			expectIssues:   false,
		},
		{
			name: "High CPU",
			metrics: &PrometheusHealthMetrics{
				Resources: ResourceMetrics{
					CPUUsagePercent:    85,
					MemoryUsagePercent: 60,
				},
				Connections: ConnectionMetrics{
					ActiveConnections: 30,
					MaxConnections:    100,
					UsagePercent:      30,
				},
				Storage: StorageMetrics{
					PageCacheHitRate: 90,
				},
			},
			expectedStatus: "warning",
			expectIssues:   true,
		},
		{
			name: "High Memory",
			metrics: &PrometheusHealthMetrics{
				Resources: ResourceMetrics{
					CPUUsagePercent:    50,
					MemoryUsagePercent: 90,
				},
				Connections: ConnectionMetrics{
					ActiveConnections: 30,
					MaxConnections:    100,
					UsagePercent:      30,
				},
				Storage: StorageMetrics{
					PageCacheHitRate: 90,
				},
			},
			expectedStatus: "warning",
			expectIssues:   true,
		},
		{
			name: "Low Page Cache",
			metrics: &PrometheusHealthMetrics{
				Resources: ResourceMetrics{
					CPUUsagePercent:    50,
					MemoryUsagePercent: 60,
				},
				Connections: ConnectionMetrics{
					ActiveConnections: 30,
					MaxConnections:    100,
					UsagePercent:      30,
				},
				Storage: StorageMetrics{
					PageCacheHitRate: 30,
				},
			},
			expectedStatus: "warning",
			expectIssues:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.metrics.Issues = []string{}
			tt.metrics.Recommendations = []string{}
			
			status := promSvc.assessHealth(tt.metrics)
			
			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}
			
			if tt.expectIssues && len(tt.metrics.Issues) == 0 {
				t.Error("Expected issues to be reported, but none were found")
			}
			
			if !tt.expectIssues && len(tt.metrics.Issues) > 0 {
				t.Errorf("Expected no issues, but found: %v", tt.metrics.Issues)
			}
		})
	}
}
