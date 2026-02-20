package aura

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/LackOfMorals/aura-client/internal/utils"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// FetchRawMetrics fetches and parses raw Prometheus metrics from an Aura metrics endpoint
func (p *prometheusService) FetchRawMetrics(ctx context.Context, prometheusURL string) (*PrometheusMetricsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	p.logger.DebugContext(ctx, "fetching raw Prometheus metrics", slog.String("url", prometheusURL))

	if prometheusURL == "" {
		return nil, fmt.Errorf("prometheus URL cannot be empty")
	}

	resp, err := p.api.Get(ctx, prometheusURL)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to fetch raw metrics", slog.String("error", err.Error()))
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		p.logger.ErrorContext(ctx, "metrics endpoint returned non-200 status",
			slog.Int("statusCode", resp.StatusCode))
		return nil, fmt.Errorf("metrics endpoint failed with status %d", resp.StatusCode)
	}

	metrics, err := p.parsePrometheusMetrics(resp.Body)
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to parse metrics", slog.String("error", err.Error()))
		return nil, err
	}

	p.logger.DebugContext(ctx, "raw metrics fetched successfully",
		slog.Int("metricCount", len(metrics.Metrics)))

	return metrics, nil
}

// parsePrometheusMetrics parses Prometheus metrics using the official client library
func (p *prometheusService) parsePrometheusMetrics(data []byte) (*PrometheusMetricsResponse, error) {
	result := &PrometheusMetricsResponse{
		Metrics: make(map[string][]PrometheusMetric),
	}

	reader := strings.NewReader(string(data))
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(reader)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to parse Prometheus metrics: %w", err)
	}

	for name, mf := range metricFamilies {
		for _, m := range mf.Metric {
			metric := PrometheusMetric{
				Name:   name,
				Labels: make(map[string]string),
			}

			for _, label := range m.Label {
				if label.Name != nil && label.Value != nil {
					metric.Labels[*label.Name] = *label.Value
				}
			}

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
					metric.Value = *m.Summary.SampleSum
				}
			case dto.MetricType_HISTOGRAM:
				if m.Histogram != nil && m.Histogram.SampleSum != nil {
					metric.Value = *m.Histogram.SampleSum
				}
			}

			if m.TimestampMs != nil {
				metric.Timestamp = *m.TimestampMs
			}

			result.Metrics[name] = append(result.Metrics[name], metric)
		}
	}

	return result, nil
}

// GetInstanceHealth retrieves comprehensive health metrics for an instance
func (p *prometheusService) GetInstanceHealth(ctx context.Context, instanceID string, prometheusURL string) (*PrometheusHealthMetrics, error) {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	p.logger.DebugContext(ctx, "getting instance health metrics", slog.String("instanceID", instanceID))

	if err := utils.ValidateInstanceID(instanceID); err != nil {
		return nil, err
	}

	if prometheusURL == "" {
		return nil, fmt.Errorf("prometheus URL cannot be empty")
	}

	// FetchRawMetrics will create its own child timeout from ctx, which is fine —
	// the outer timeout here acts as a ceiling for the entire GetInstanceHealth call.
	rawMetrics, err := p.FetchRawMetrics(ctx, prometheusURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}

	metrics := &PrometheusHealthMetrics{
		InstanceID:      instanceID,
		Timestamp:       time.Now(),
		Issues:          []string{},
		Recommendations: []string{},
	}

	if cpuUsage, err := p.GetMetricValue(ctx, rawMetrics, "neo4j_aura_cpu_usage", nil); err == nil {
		if cpuLimit, err := p.GetMetricValue(ctx, rawMetrics, "neo4j_aura_cpu_limit", nil); err == nil && cpuLimit > 0 {
			metrics.Resources.CPUUsagePercent = (cpuUsage / cpuLimit) * 100
		}
	} else {
		p.logger.WarnContext(ctx, "failed to get CPU usage", slog.String("error", err.Error()))
	}

	if heapRatio, err := p.GetMetricValue(ctx, rawMetrics, "neo4j_dbms_vm_heap_used_ratio", nil); err == nil {
		metrics.Resources.MemoryUsagePercent = heapRatio * 100
	} else {
		p.logger.WarnContext(ctx, "failed to get memory usage", slog.String("error", err.Error()))
	}

	if successCount, err := p.GetMetricValue(ctx, rawMetrics, "neo4j_db_query_execution_success_total", nil); err == nil {
		metrics.Query.QueriesPerSecond = successCount
	} else {
		p.logger.WarnContext(ctx, "failed to get query count", slog.String("error", err.Error()))
	}

	if latency, err := p.GetMetricValue(ctx, rawMetrics, "neo4j_db_query_execution_internal_latency_q50", nil); err == nil {
		metrics.Query.AvgLatencyMS = latency
	} else {
		p.logger.WarnContext(ctx, "failed to get query latency", slog.String("error", err.Error()))
	}

	if idle, err := p.GetMetricValue(ctx, rawMetrics, "neo4j_dbms_bolt_connections_idle", nil); err == nil {
		if running, err := p.GetMetricValue(ctx, rawMetrics, "neo4j_dbms_bolt_connections_running", nil); err == nil {
			metrics.Connections.ActiveConnections = int(idle + running)
		}
	}

	// MaxConnections assumes the standard Aura limit; this may vary by plan.
	metrics.Connections.MaxConnections = 100
	if metrics.Connections.MaxConnections > 0 {
		metrics.Connections.UsagePercent = float64(metrics.Connections.ActiveConnections) / float64(metrics.Connections.MaxConnections) * 100
	}

	if hitRate, err := p.GetMetricValue(ctx, rawMetrics, "neo4j_dbms_page_cache_hit_ratio_per_minute", nil); err == nil {
		metrics.Storage.PageCacheHitRate = hitRate * 100
	} else {
		p.logger.WarnContext(ctx, "failed to get page cache hit rate", slog.String("error", err.Error()))
	}

	metrics.OverallStatus = p.assessHealth(metrics)

	p.logger.InfoContext(ctx, "instance health metrics retrieved",
		slog.String("instanceID", instanceID),
		slog.String("status", metrics.OverallStatus))

	return metrics, nil
}

// GetMetricValue retrieves a specific metric value by name and optional label filters.
// When no filters are provided it averages across all series for that metric name.
func (p *prometheusService) GetMetricValue(_ context.Context, metrics *PrometheusMetricsResponse, name string, labelFilters map[string]string) (float64, error) {
	metricList, ok := metrics.Metrics[name]
	if !ok {
		return 0, fmt.Errorf("metric %s not found", name)
	}

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

	var sum float64
	for _, m := range matchingMetrics {
		sum += m.Value
	}
	return sum / float64(len(matchingMetrics)), nil
}

// assessHealth analyzes metrics and determines overall health status
func (p *prometheusService) assessHealth(metrics *PrometheusHealthMetrics) string {
	status := "healthy"

	if metrics.Resources.CPUUsagePercent > 80 {
		metrics.Issues = append(metrics.Issues, fmt.Sprintf("High CPU usage: %.1f%%", metrics.Resources.CPUUsagePercent))
		metrics.Recommendations = append(metrics.Recommendations, "Consider scaling to a larger instance size")
		status = "warning"
	}

	if metrics.Resources.MemoryUsagePercent > 85 {
		metrics.Issues = append(metrics.Issues, fmt.Sprintf("High memory usage: %.1f%%", metrics.Resources.MemoryUsagePercent))
		metrics.Recommendations = append(metrics.Recommendations, "Consider scaling to a larger memory instance")
		if status == "healthy" {
			status = "warning"
		}
	}

	if metrics.Connections.UsagePercent > 80 {
		metrics.Issues = append(metrics.Issues, fmt.Sprintf("High connection usage: %.1f%%", metrics.Connections.UsagePercent))
		metrics.Recommendations = append(metrics.Recommendations, "Review connection pooling configuration in your application")
		if status == "healthy" {
			status = "warning"
		}
	}

	if metrics.Storage.PageCacheHitRate < 50 && metrics.Storage.PageCacheHitRate > 0 {
		metrics.Issues = append(metrics.Issues, fmt.Sprintf("Low page cache hit rate: %.1f%%", metrics.Storage.PageCacheHitRate))
		metrics.Recommendations = append(metrics.Recommendations, "Consider increasing page cache size for better performance")
		if status == "healthy" {
			status = "warning"
		}
	}

	return status
}
