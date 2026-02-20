package aura

import (
	"log/slog"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// Prometheus Metrics Types

// PrometheusHealthMetrics contains parsed health metrics for an instance
type PrometheusHealthMetrics struct {
	InstanceID      string            `json:"instance_id"`
	Timestamp       time.Time         `json:"timestamp"`
	Resources       ResourceMetrics   `json:"resources"`
	Query           QueryMetrics      `json:"query"`
	Connections     ConnectionMetrics `json:"connections"`
	Storage         StorageMetrics    `json:"storage"`
	OverallStatus   string            `json:"overall_status"`
	Issues          []string          `json:"issues"`
	Recommendations []string          `json:"recommendations"`
}

// ResourceMetrics contains CPU and memory usage
type ResourceMetrics struct {
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
}

// QueryMetrics contains query performance statistics
type QueryMetrics struct {
	QueriesPerSecond float64 `json:"queries_per_second"`
	AvgLatencyMS     float64 `json:"avg_latency_ms"`
}

// ConnectionMetrics contains connection pool information
type ConnectionMetrics struct {
	ActiveConnections int     `json:"active_connections"`
	MaxConnections    int     `json:"max_connections"`
	UsagePercent      float64 `json:"usage_percent"`
}

// StorageMetrics contains storage usage information
type StorageMetrics struct {
	PageCacheHitRate float64 `json:"page_cache_hit_rate,omitempty"`
}

// PrometheusMetric represents a single parsed metric from Prometheus exposition format
type PrometheusMetric struct {
	Name      string
	Labels    map[string]string
	Value     float64
	Timestamp int64
}

// PrometheusMetricsResponse contains parsed metrics from the raw endpoint
type PrometheusMetricsResponse struct {
	Metrics map[string][]PrometheusMetric
}

// prometheusService handles Prometheus metrics operations
type prometheusService struct {
	api     api.RequestService
	timeout time.Duration
	logger  *slog.Logger
}
