package aura

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
	"github.com/LackOfMorals/aura-client/internal/utils"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
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
	api    api.RequestService
	ctx    context.Context
	logger *slog.Logger
}

// FetchRawMetrics fetches and parses raw Prometheus metrics from an Aura metrics endpoint
// using the official Prometheus client library for robust parsing
func (p *prometheusService) FetchRawMetrics(prometheusURL string) (*PrometheusMetricsResponse, error) {
	p.logger.DebugContext(p.ctx, "fetching raw Prometheus metrics", slog.String("url", prometheusURL))

	if prometheusURL == "" {
		return nil, fmt.Errorf("prometheus URL cannot be empty")
	}

	// Fetch the raw metrics
	resp, err := p.api.Get(p.ctx, prometheusURL)
	if err != nil {
		p.logger.ErrorContext(p.ctx, "failed to fetch raw metrics", slog.String("error", err.Error()))
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		p.logger.ErrorContext(p.ctx, "metrics endpoint returned non-200 status",
			slog.Int("statusCode", resp.StatusCode))
		return nil, fmt.Errorf("metrics endpoint failed with status %d", resp.StatusCode)
	}

	// Parse using official Prometheus library
	metrics, err := p.parsePrometheusMetrics(resp.Body)
	if err != nil {
		p.logger.ErrorContext(p.ctx, "failed to parse metrics", slog.String("error", err.Error()))
		return nil, err
	}

	p.logger.DebugContext(p.ctx, "raw metrics fetched successfully",
		slog.Int("metricCount", len(metrics.Metrics)))

	return metrics, nil
}

// parsePrometheusMetrics parses Prometheus metrics using the official client library
func (p *prometheusService) parsePrometheusMetrics(data []byte) (*PrometheusMetricsResponse, error) {
	result := &PrometheusMetricsResponse{
		Metrics: make(map[string][]PrometheusMetric),
	}

	// Create a text parser using the official Prometheus library
	reader := strings.NewReader(string(data))
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(reader)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to parse Prometheus metrics: %w", err)
	}

	// Convert Prometheus metric families to our simplified format
	for name, mf := range metricFamilies {
		for _, m := range mf.Metric {
			metric := PrometheusMetric{
				Name:   name,
				Labels: make(map[string]string),
			}

			// Extract labels
			for _, label := range m.Label {
				if label.Name != nil && label.Value != nil {
					metric.Labels[*label.Name] = *label.Value
				}
			}

			// Extract value based on metric type
			switch mf.GetType() {
			case dto.MetricType_COUNTER:
				if m.Counter != nil && m.Counter.Value != nil {
					metric.Value = *m.Counter.Value
				}
			case dto.MetricType_GAUGE:
				if m.Gauge != nil && m.Gauge.Value != nil {
					metric.Value = *m.Gauge.Value
				}
			case dto.MetricType_UNTYPED:
				if m.Untyped != nil && m.Untyped.Value != nil {
					metric.Value = *m.Untyped.Value
				}
			case dto.MetricType_SUMMARY:
				if m.Summary != nil && m.Summary.SampleSum != nil {
					// For summaries, use the sum
					metric.Value = *m.Summary.SampleSum
				}
			case dto.MetricType_HISTOGRAM:
				if m.Histogram != nil && m.Histogram.SampleSum != nil {
					// For histograms, use the sum
					metric.Value = *m.Histogram.SampleSum
				}
			}

			// Extract timestamp if available
			if m.TimestampMs != nil {
				metric.Timestamp = *m.TimestampMs
			}

			result.Metrics[name] = append(result.Metrics[name], metric)
		}
	}

	return result, nil
}

// GetInstanceHealth retrieves comprehensive health metrics for an instance
func (p *prometheusService) GetInstanceHealth(instanceID string, prometheusURL string) (*PrometheusHealthMetrics, error) {
	p.logger.DebugContext(p.ctx, "getting instance health metrics", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		return nil, err
	}

	if prometheusURL == "" {
		return nil, fmt.Errorf("prometheus URL cannot be empty")
	}

	// Fetch raw metrics
	rawMetrics, err := p.FetchRawMetrics(prometheusURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}

	// Initialize health metrics
	metrics := &PrometheusHealthMetrics{
		InstanceID:      instanceID,
		Timestamp:       time.Now(),
		Issues:          []string{},
		Recommendations: []string{},
	}

	// CPU Usage - from neo4j_aura_cpu_usage and neo4j_aura_cpu_limit
	if cpuUsage, err := p.GetMetricValue(rawMetrics, "neo4j_aura_cpu_usage", nil); err == nil {
		// Get CPU limit
		if cpuLimit, err := p.GetMetricValue(rawMetrics, "neo4j_aura_cpu_limit", nil); err == nil && cpuLimit > 0 {
			metrics.Resources.CPUUsagePercent = (cpuUsage / cpuLimit) * 100
		}
	} else {
		p.logger.WarnContext(p.ctx, "failed to get CPU usage", slog.String("error", err.Error()))
	}

	// Memory Usage - from neo4j_dbms_vm_heap_used_ratio (already a ratio 0-1)
	if heapRatio, err := p.GetMetricValue(rawMetrics, "neo4j_dbms_vm_heap_used_ratio", nil); err == nil {
		metrics.Resources.MemoryUsagePercent = heapRatio * 100
	} else {
		p.logger.WarnContext(p.ctx, "failed to get memory usage", slog.String("error", err.Error()))
	}

	// Query metrics - from neo4j_db_query_execution_success_total
	if successCount, err := p.GetMetricValue(rawMetrics, "neo4j_db_query_execution_success_total", nil); err == nil {
		// This is a counter total, not a rate
		metrics.Query.QueriesPerSecond = successCount
	} else {
		p.logger.WarnContext(p.ctx, "failed to get query count", slog.String("error", err.Error()))
	}

	// Query Latency - from neo4j_db_query_execution_internal_latency_q50
	if latency, err := p.GetMetricValue(rawMetrics, "neo4j_db_query_execution_internal_latency_q50", nil); err == nil {
		metrics.Query.AvgLatencyMS = latency
	} else {
		p.logger.WarnContext(p.ctx, "failed to get query latency", slog.String("error", err.Error()))
	}

	// Connection Pool Metrics - from neo4j_dbms_bolt_connections_*
	if idle, err := p.GetMetricValue(rawMetrics, "neo4j_dbms_bolt_connections_idle", nil); err == nil {
		if running, err := p.GetMetricValue(rawMetrics, "neo4j_dbms_bolt_connections_running", nil); err == nil {
			metrics.Connections.ActiveConnections = int(idle + running)
		}
	}

	// Max connections (typically 100 for Aura instances)
	metrics.Connections.MaxConnections = 100
	if metrics.Connections.MaxConnections > 0 {
		metrics.Connections.UsagePercent = float64(metrics.Connections.ActiveConnections) / float64(metrics.Connections.MaxConnections) * 100
	}

	// Page Cache Hit Rate - from neo4j_dbms_page_cache_hit_ratio_per_minute (already a ratio 0-1)
	if hitRate, err := p.GetMetricValue(rawMetrics, "neo4j_dbms_page_cache_hit_ratio_per_minute", nil); err == nil {
		metrics.Storage.PageCacheHitRate = hitRate * 100
	} else {
		p.logger.WarnContext(p.ctx, "failed to get page cache hit rate", slog.String("error", err.Error()))
	}

	// Assess overall health and generate recommendations
	metrics.OverallStatus = p.assessHealth(metrics)

	p.logger.InfoContext(p.ctx, "instance health metrics retrieved",
		slog.String("instanceID", instanceID),
		slog.String("status", metrics.OverallStatus))

	return metrics, nil
}

// GetMetricValue retrieves a specific metric value by name and optional label filters
func (p *prometheusService) GetMetricValue(metrics *PrometheusMetricsResponse, name string, labelFilters map[string]string) (float64, error) {
	metricList, ok := metrics.Metrics[name]
	if !ok {
		return 0, fmt.Errorf("metric %s not found", name)
	}

	// If no filters, average across all instances
	if len(labelFilters) == 0 {
		if len(metricList) == 0 {
			return 0, fmt.Errorf("no values for metric %s", name)
		}
		var sum float64
		for _, m := range metricList {
			sum += m.Value
		}
		return sum / float64(len(metricList)), nil
	}

	// Filter by labels
	var matchingMetrics []PrometheusMetric
	for _, m := range metricList {
		match := true
		for key, value := range labelFilters {
			if m.Labels[key] != value {
				match = false
				break
			}
		}
		if match {
			matchingMetrics = append(matchingMetrics, m)
		}
	}

	if len(matchingMetrics) == 0 {
		return 0, fmt.Errorf("no matching metrics found for %s with filters %v", name, labelFilters)
	}

	// Average across matching metrics
	var sum float64
	for _, m := range matchingMetrics {
		sum += m.Value
	}
	return sum / float64(len(matchingMetrics)), nil
}

// assessHealth analyzes metrics and determines overall health status
func (p *prometheusService) assessHealth(metrics *PrometheusHealthMetrics) string {
	status := "healthy"

	// Check CPU usage
	if metrics.Resources.CPUUsagePercent > 80 {
		metrics.Issues = append(metrics.Issues, fmt.Sprintf("High CPU usage: %.1f%%", metrics.Resources.CPUUsagePercent))
		metrics.Recommendations = append(metrics.Recommendations, "Consider scaling to a larger instance size")
		status = "warning"
	}

	// Check memory usage
	if metrics.Resources.MemoryUsagePercent > 85 {
		metrics.Issues = append(metrics.Issues, fmt.Sprintf("High memory usage: %.1f%%", metrics.Resources.MemoryUsagePercent))
		metrics.Recommendations = append(metrics.Recommendations, "Consider scaling to a larger memory instance")
		if status == "healthy" {
			status = "warning"
		}
	}

	// Check connection pool
	if metrics.Connections.UsagePercent > 80 {
		metrics.Issues = append(metrics.Issues, fmt.Sprintf("High connection usage: %.1f%%", metrics.Connections.UsagePercent))
		metrics.Recommendations = append(metrics.Recommendations, "Review connection pooling configuration in your application")
		if status == "healthy" {
			status = "warning"
		}
	}

	// Check page cache hit rate
	if metrics.Storage.PageCacheHitRate < 50 && metrics.Storage.PageCacheHitRate > 0 {
		metrics.Issues = append(metrics.Issues, fmt.Sprintf("Low page cache hit rate: %.1f%%", metrics.Storage.PageCacheHitRate))
		metrics.Recommendations = append(metrics.Recommendations, "Consider increasing page cache size for better performance")
		if status == "healthy" {
			status = "warning"
		}
	}

	return status
}
