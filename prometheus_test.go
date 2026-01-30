package aura

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/LackOfMorals/aura-client/internal/api"
	"github.com/LackOfMorals/aura-client/internal/httpClient"
	"time"
)

func TestPrometheusService_FetchRawMetrics(t *testing.T) {
	// Create a prometheus service
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	httpSvc := httpClient.NewHTTPService("https://api.neo4j.io/", "", 30*time.Second, 3, logger)
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
		_, err := promSvc.FetchRawMetrics("")
		if err == nil {
			t.Error("Expected error for empty URL, got nil")
		}
	})

	// Note: Testing actual Prometheus metrics would require a real instance
	// and valid credentials, which is typically done in integration tests
}

func TestPrometheusService_ParseMetricLine(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelWarn}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	httpSvc := httpClient.NewHTTPService("https://api.neo4j.io/", "", 30*time.Second, 3, logger)
	apiSvc := api.NewAPIRequestService(httpSvc, api.Config{
		ClientID:     "test",
		ClientSecret: "test",
		APIVersion:   "v1",
		Timeout:      30 * time.Second,
	}, logger)
	
	promSvc := &prometheusService{
		api:    apiSvc,
		ctx:    context.Background(),
		logger: logger,
	}

	tests := []struct {
		name          string
		line          string
		expectedName  string
		expectedValue float64
		expectError   bool
	}{
		{
			name:          "Valid metric with labels",
			line:          `neo4j_aura_cpu_usage{aggregation="MAX",availability_zone="europe-west2-c",instance_mode="PRIMARY",instance_id="c9f0d13a"} 0.023206 1769766720469`,
			expectedName:  "neo4j_aura_cpu_usage",
			expectedValue: 0.023206,
			expectError:   false,
		},
		{
			name:          "Valid metric simple",
			line:          `neo4j_database_count_node{database="neo4j",instance_id="c9f0d13a"} 171.000000 1769766720469`,
			expectedName:  "neo4j_database_count_node",
			expectedValue: 171.0,
			expectError:   false,
		},
		{
			name:        "Invalid metric no labels",
			line:        `neo4j_metric_value 123`,
			expectError: true,
		},
		{
			name:        "Invalid metric no value",
			line:        `neo4j_metric{label="value"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric, err := promSvc.parseMetricLine(tt.line)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			
			if metric.Name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, metric.Name)
			}
			
			if metric.Value != tt.expectedValue {
				t.Errorf("Expected value %f, got %f", tt.expectedValue, metric.Value)
			}
		})
	}
}

func TestPrometheusService_GetInstanceHealth(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	httpSvc := httpClient.NewHTTPService("https://api.neo4j.io/", "", 30*time.Second, 3, logger)
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

func TestPrometheusService_GetMetricValue(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelWarn}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	httpSvc := httpClient.NewHTTPService("https://api.neo4j.io/", "", 30*time.Second, 3, logger)
	apiSvc := api.NewAPIRequestService(httpSvc, api.Config{
		ClientID:     "test",
		ClientSecret: "test",
		APIVersion:   "v1",
		Timeout:      30 * time.Second,
	}, logger)
	
	promSvc := &prometheusService{
		api:    apiSvc,
		ctx:    context.Background(),
		logger: logger,
	}

	// Create test metrics
	testMetrics := &PrometheusMetricsResponse{
		Metrics: map[string][]PrometheusMetric{
			"test_metric": {
				{
					Name:  "test_metric",
					Value: 10.0,
					Labels: map[string]string{
						"zone": "zone-a",
						"type": "primary",
					},
				},
				{
					Name:  "test_metric",
					Value: 20.0,
					Labels: map[string]string{
						"zone": "zone-b",
						"type": "primary",
					},
				},
				{
					Name:  "test_metric",
					Value: 15.0,
					Labels: map[string]string{
						"zone": "zone-c",
						"type": "secondary",
					},
				},
			},
		},
	}

	t.Run("Average all metrics", func(t *testing.T) {
		value, err := promSvc.GetMetricValue(testMetrics, "test_metric", nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		expected := (10.0 + 20.0 + 15.0) / 3.0
		if value != expected {
			t.Errorf("Expected average %f, got %f", expected, value)
		}
	})

	t.Run("Filter by label", func(t *testing.T) {
		value, err := promSvc.GetMetricValue(testMetrics, "test_metric", map[string]string{"type": "primary"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		expected := (10.0 + 20.0) / 2.0
		if value != expected {
			t.Errorf("Expected average %f, got %f", expected, value)
		}
	})

	t.Run("Filter by multiple labels", func(t *testing.T) {
		value, err := promSvc.GetMetricValue(testMetrics, "test_metric", map[string]string{"zone": "zone-a", "type": "primary"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if value != 10.0 {
			t.Errorf("Expected value 10.0, got %f", value)
		}
	})

	t.Run("Metric not found", func(t *testing.T) {
		_, err := promSvc.GetMetricValue(testMetrics, "nonexistent_metric", nil)
		if err == nil {
			t.Error("Expected error for nonexistent metric, got nil")
		}
	})

	t.Run("No matching labels", func(t *testing.T) {
		_, err := promSvc.GetMetricValue(testMetrics, "test_metric", map[string]string{"zone": "zone-d"})
		if err == nil {
			t.Error("Expected error for non-matching label filter, got nil")
		}
	})
}

func TestAssessHealth(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	httpSvc := httpClient.NewHTTPService("https://api.neo4j.io/", "", 30*time.Second, 3, logger)
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
