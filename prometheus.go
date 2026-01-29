package aura

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	httpClient "github.com/LackOfMorals/aura-client/internal/httpClient"
	"github.com/LackOfMorals/aura-client/internal/utils"
)

// Prometheus Metrics Types

// PrometheusQueryResponse represents a response from the Prometheus instant query API
type PrometheusQueryResponse struct {
	Status string                  `json:"status"`
	Data   PrometheusQueryData     `json:"data"`
	Error  string                  `json:"error,omitempty"`
	ErrorType string               `json:"errorType,omitempty"`
}

// PrometheusQueryData contains the result data from a Prometheus query
type PrometheusQueryData struct {
	ResultType string              `json:"resultType"`
	Result     []PrometheusResult  `json:"result"`
}

// PrometheusResult represents a single metric result
type PrometheusResult struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}

// PrometheusRangeQueryResponse represents a response from the Prometheus range query API
type PrometheusRangeQueryResponse struct {
	Status string                      `json:"status"`
	Data   PrometheusRangeQueryData    `json:"data"`
	Error  string                      `json:"error,omitempty"`
	ErrorType string                   `json:"errorType,omitempty"`
}

// PrometheusRangeQueryData contains the result data from a Prometheus range query
type PrometheusRangeQueryData struct {
	ResultType string                   `json:"resultType"`
	Result     []PrometheusRangeResult  `json:"result"`
}

// PrometheusRangeResult represents a single metric result with time series data
type PrometheusRangeResult struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
}

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

// prometheusService handles Prometheus metrics operations
type prometheusService struct {
	httpClient httpClient.HTTPService
	ctx        context.Context
	logger     *slog.Logger
}

// Query executes an instant query against a Prometheus endpoint
func (p *prometheusService) Query(prometheusURL string, query string) (*PrometheusQueryResponse, error) {
	p.logger.DebugContext(p.ctx, "executing Prometheus instant query", slog.String("query", query))

	if prometheusURL == "" {
		return nil, fmt.Errorf("prometheus URL cannot be empty")
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("time", fmt.Sprintf("%d", time.Now().Unix()))

	fullURL := prometheusURL + "/api/v1/query?" + params.Encode()

	resp, err := p.httpClient.Get(p.ctx, fullURL, nil)
	if err != nil {
		p.logger.ErrorContext(p.ctx, "failed to execute Prometheus query", slog.String("error", err.Error()))
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		p.logger.ErrorContext(p.ctx, "Prometheus query returned non-200 status", 
			slog.Int("statusCode", resp.StatusCode),
			slog.String("body", string(resp.Body)))
		return nil, fmt.Errorf("prometheus query failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var result PrometheusQueryResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		p.logger.ErrorContext(p.ctx, "failed to unmarshal Prometheus response", slog.String("error", err.Error()))
		return nil, err
	}

	if result.Status != "success" {
		p.logger.ErrorContext(p.ctx, "Prometheus query failed", 
			slog.String("status", result.Status),
			slog.String("error", result.Error),
			slog.String("errorType", result.ErrorType))
		return nil, fmt.Errorf("prometheus query failed: %s - %s", result.ErrorType, result.Error)
	}

	p.logger.DebugContext(p.ctx, "Prometheus query executed successfully", 
		slog.Int("resultCount", len(result.Data.Result)))
	return &result, nil
}

// QueryRange executes a range query against a Prometheus endpoint
func (p *prometheusService) QueryRange(prometheusURL string, query string, start, end time.Time, step string) (*PrometheusRangeQueryResponse, error) {
	p.logger.DebugContext(p.ctx, "executing Prometheus range query", 
		slog.String("query", query),
		slog.Time("start", start),
		slog.Time("end", end),
		slog.String("step", step))

	if prometheusURL == "" {
		return nil, fmt.Errorf("prometheus URL cannot be empty")
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("start", fmt.Sprintf("%d", start.Unix()))
	params.Set("end", fmt.Sprintf("%d", end.Unix()))
	params.Set("step", step)

	fullURL := prometheusURL + "/api/v1/query_range?" + params.Encode()

	resp, err := p.httpClient.Get(p.ctx, fullURL, nil)
	if err != nil {
		p.logger.ErrorContext(p.ctx, "failed to execute Prometheus range query", slog.String("error", err.Error()))
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		p.logger.ErrorContext(p.ctx, "Prometheus range query returned non-200 status", 
			slog.Int("statusCode", resp.StatusCode),
			slog.String("body", string(resp.Body)))
		return nil, fmt.Errorf("prometheus range query failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var result PrometheusRangeQueryResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		p.logger.ErrorContext(p.ctx, "failed to unmarshal Prometheus range response", slog.String("error", err.Error()))
		return nil, err
	}

	if result.Status != "success" {
		p.logger.ErrorContext(p.ctx, "Prometheus range query failed", 
			slog.String("status", result.Status),
			slog.String("error", result.Error),
			slog.String("errorType", result.ErrorType))
		return nil, fmt.Errorf("prometheus range query failed: %s - %s", result.ErrorType, result.Error)
	}

	p.logger.DebugContext(p.ctx, "Prometheus range query executed successfully", 
		slog.Int("resultCount", len(result.Data.Result)))
	return &result, nil
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

	// Query various metrics
	metrics := &PrometheusHealthMetrics{
		InstanceID: instanceID,
		Timestamp:  time.Now(),
		Issues:     []string{},
		Recommendations: []string{},
	}

	// CPU Usage
	if cpuUsage, err := p.querySingleMetric(prometheusURL, "rate(process_cpu_seconds_total[5m]) * 100"); err == nil {
		metrics.Resources.CPUUsagePercent = cpuUsage
	} else {
		p.logger.WarnContext(p.ctx, "failed to get CPU usage", slog.String("error", err.Error()))
	}

	// Memory Usage
	if memUsage, err := p.querySingleMetric(prometheusURL, "process_resident_memory_bytes / process_virtual_memory_bytes * 100"); err == nil {
		metrics.Resources.MemoryUsagePercent = memUsage
	} else {
		p.logger.WarnContext(p.ctx, "failed to get memory usage", slog.String("error", err.Error()))
	}

	// Query Rate (queries per second)
	if qps, err := p.querySingleMetric(prometheusURL, "rate(neo4j_database_system_check_point_events_total[5m])"); err == nil {
		metrics.Query.QueriesPerSecond = qps
	} else {
		p.logger.WarnContext(p.ctx, "failed to get query rate", slog.String("error", err.Error()))
	}

	// Average Query Latency
	if latency, err := p.querySingleMetric(prometheusURL, "rate(neo4j_database_system_check_point_duration_seconds_total[5m]) / rate(neo4j_database_system_check_point_events_total[5m]) * 1000"); err == nil {
		metrics.Query.AvgLatencyMS = latency
	} else {
		p.logger.WarnContext(p.ctx, "failed to get query latency", slog.String("error", err.Error()))
	}

	// Connection Pool Metrics
	if activeConns, err := p.querySingleMetric(prometheusURL, "bolt_connections_opened - bolt_connections_closed"); err == nil {
		metrics.Connections.ActiveConnections = int(activeConns)
	}

	// Max connections (typically 100 for Aura instances)
	metrics.Connections.MaxConnections = 100
	if metrics.Connections.MaxConnections > 0 {
		metrics.Connections.UsagePercent = float64(metrics.Connections.ActiveConnections) / float64(metrics.Connections.MaxConnections) * 100
	}

	// Page Cache Hit Rate
	if hitRate, err := p.querySingleMetric(prometheusURL, "rate(neo4j_page_cache_hits_total[5m]) / (rate(neo4j_page_cache_hits_total[5m]) + rate(neo4j_page_cache_faults_total[5m])) * 100"); err == nil {
		metrics.Storage.PageCacheHitRate = hitRate
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

// querySingleMetric is a helper function to query a single metric value
func (p *prometheusService) querySingleMetric(prometheusURL, query string) (float64, error) {
	resp, err := p.Query(prometheusURL, query)
	if err != nil {
		return 0, err
	}

	if len(resp.Data.Result) == 0 {
		return 0, fmt.Errorf("no results returned for query")
	}

	// Extract the value from the result
	if len(resp.Data.Result[0].Value) < 2 {
		return 0, fmt.Errorf("invalid result format")
	}

	valueStr, ok := resp.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("value is not a string")
	}

	var value float64
	if _, err := fmt.Sscanf(valueStr, "%f", &value); err != nil {
		return 0, fmt.Errorf("failed to parse value: %w", err)
	}

	return value, nil
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
