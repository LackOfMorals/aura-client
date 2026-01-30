package aura

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
	"github.com/LackOfMorals/aura-client/internal/utils"
)

// Prometheus Metrics Types

// PrometheusHealthMetrics contains parsed health metrics for an instance
type PrometheusHealthMetrics struct {
	InstanceID    string                 `json:"instance_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Resources     ResourceMetrics        `json:"resources"`
	Query         QueryMetrics           `json:"query"`
	Connections   ConnectionMetrics      `json:"connections"`
	Storage       StorageMetrics         `json:"storage"`
	OverallStatus string                 `json:"overall_status"`
	Issues        []string               `json:"issues"`
	Recommendations []string             `json:"recommendations"`
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
	api    api.APIRequestService
	ctx    context.Context
	logger *slog.Logger
}

// FetchRawMetrics fetches and parses raw Prometheus metrics from an Aura metrics endpoint
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

	// Parse the Prometheus exposition format
	metrics, err := p.parsePrometheusText(string(resp.Body))
	if err != nil {
		p.logger.ErrorContext(p.ctx, "failed to parse metrics", slog.String("error", err.Error()))
		return nil, err
	}

	p.logger.DebugContext(p.ctx, "raw metrics fetched successfully",
		slog.Int("metricCount", len(metrics.Metrics)))

	return metrics, nil
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

// parsePrometheusText parses Prometheus exposition format text
func (p *prometheusService) parsePrometheusText(text string) (*PrometheusMetricsResponse, error) {
	metrics := &PrometheusMetricsResponse{
		Metrics: make(map[string][]PrometheusMetric),
	}

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments (HELP and TYPE)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse metric line: metric_name{labels} value timestamp
		metric, err := p.parseMetricLine(line)
		if err != nil {
			p.logger.WarnContext(p.ctx, "failed to parse metric line",
				slog.String("line", line),
				slog.String("error", err.Error()))
			continue
		}

		metrics.Metrics[metric.Name] = append(metrics.Metrics[metric.Name], metric)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning metrics: %w", err)
	}

	return metrics, nil
}

// parseMetricLine parses a single metric line in Prometheus format
// Format: metric_name{label1="value1",label2="value2"} value timestamp
func (p *prometheusService) parseMetricLine(line string) (PrometheusMetric, error) {
	metric := PrometheusMetric{
		Labels: make(map[string]string),
	}

	// Find the opening brace for labels
	braceIdx := strings.Index(line, "{")
	if braceIdx == -1 {
		return metric, fmt.Errorf("invalid metric line: no labels")
	}

	metric.Name = strings.TrimSpace(line[:braceIdx])

	// Find the closing brace
	closeIdx := strings.Index(line, "}")
	if closeIdx == -1 {
		return metric, fmt.Errorf("invalid metric line: no closing brace")
	}

	// Parse labels
	labelsStr := line[braceIdx+1 : closeIdx]
	if err := p.parseLabels(labelsStr, metric.Labels); err != nil {
		return metric, fmt.Errorf("failed to parse labels: %w", err)
	}

	// Parse value and timestamp
	rest := strings.TrimSpace(line[closeIdx+1:])
	parts := strings.Fields(rest)
	if len(parts) < 1 {
		return metric, fmt.Errorf("invalid metric line: no value")
	}

	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return metric, fmt.Errorf("failed to parse value: %w", err)
	}
	metric.Value = value

	if len(parts) >= 2 {
		timestamp, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return metric, fmt.Errorf("failed to parse timestamp: %w", err)
		}
		metric.Timestamp = timestamp
	}

	return metric, nil
}

// parseLabels parses label key-value pairs from the Prometheus format
// Format: key1="value1",key2="value2"
func (p *prometheusService) parseLabels(labelsStr string, labels map[string]string) error {
	if labelsStr == "" {
		return nil
	}

	// Split by comma, but respect quoted values
	var current strings.Builder
	inQuote := false
	for i := 0; i < len(labelsStr); i++ {
		ch := labelsStr[i]
		if ch == '"' {
			inQuote = !inQuote
			current.WriteByte(ch)
		} else if ch == ',' && !inQuote {
			if err := p.parseLabel(current.String(), labels); err != nil {
				return err
			}
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}

	// Parse the last label
	if current.Len() > 0 {
		if err := p.parseLabel(current.String(), labels); err != nil {
			return err
		}
	}

	return nil
}

// parseLabel parses a single label key-value pair
// Format: key="value"
func (p *prometheusService) parseLabel(labelStr string, labels map[string]string) error {
	parts := strings.SplitN(labelStr, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid label format: %s", labelStr)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.Trim(strings.TrimSpace(parts[1]), "\"")
	labels[key] = value

	return nil
}
