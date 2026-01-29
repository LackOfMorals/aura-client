# Prometheus Client for Neo4j Aura

This package provides a Go client for querying Prometheus metrics from Neo4j Aura instances.

## Features

- **Instant Queries**: Execute instant queries against Prometheus endpoints
- **Range Queries**: Query metrics over time ranges
- **Health Monitoring**: Get comprehensive health metrics for instances
- **Auto-parsing**: Automatically parse and validate Prometheus responses

## Installation

```bash
go get github.com/LackOfMorals/aura-client
```

## Usage

### Basic Setup

```go
import aura "github.com/LackOfMorals/aura-client"

client, err := aura.NewClient(
    aura.WithCredentials("client-id", "client-secret"),
)
if err != nil {
    log.Fatal(err)
}
```

### Getting the Prometheus URL

Each Aura instance has a Prometheus metrics endpoint. You can get the URL from the instance details:

```go
instance, err := client.Instances.Get("instance-id")
if err != nil {
    log.Fatal(err)
}

prometheusURL := instance.Data.MetricsURL
// e.g., "https://c9f0d13a.metrics.neo4j.io/prometheus"
```

### Instant Queries

Execute a Prometheus instant query to get the current value of a metric:

```go
resp, err := client.Prometheus.Query(prometheusURL, "up")
if err != nil {
    log.Fatal(err)
}

for _, result := range resp.Data.Result {
    fmt.Printf("Metric: %v, Value: %v\n", result.Metric, result.Value)
}
```

### Range Queries

Query metrics over a time range:

```go
end := time.Now()
start := end.Add(-1 * time.Hour)

resp, err := client.Prometheus.QueryRange(
    prometheusURL,
    "rate(process_cpu_seconds_total[5m])",
    start,
    end,
    "5m", // step
)
if err != nil {
    log.Fatal(err)
}

for _, result := range resp.Data.Result {
    fmt.Printf("Metric: %v\n", result.Metric)
    fmt.Printf("Data points: %d\n", len(result.Values))
}
```

### Instance Health Monitoring

Get comprehensive health metrics for an instance:

```go
health, err := client.Prometheus.GetInstanceHealth(instanceID, prometheusURL)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Status: %s\n", health.OverallStatus)
fmt.Printf("CPU Usage: %.2f%%\n", health.Resources.CPUUsagePercent)
fmt.Printf("Memory Usage: %.2f%%\n", health.Resources.MemoryUsagePercent)
fmt.Printf("Queries/sec: %.2f\n", health.Query.QueriesPerSecond)
fmt.Printf("Connections: %d/%d\n", 
    health.Connections.ActiveConnections,
    health.Connections.MaxConnections)

if len(health.Issues) > 0 {
    fmt.Println("Issues detected:")
    for _, issue := range health.Issues {
        fmt.Printf("  - %s\n", issue)
    }
}

if len(health.Recommendations) > 0 {
    fmt.Println("Recommendations:")
    for _, rec := range health.Recommendations {
        fmt.Printf("  - %s\n", rec)
    }
}
```

## Common Neo4j Metrics

Here are some useful Prometheus queries for Neo4j Aura:

### Database Metrics

```go
// Transaction rate
"rate(neo4j_transaction_started_total[5m])"

// Store size
"neo4j_store_size_total"

// Database operations
"rate(neo4j_database_system_check_point_events_total[5m])"
```

### Performance Metrics

```go
// CPU usage
"rate(process_cpu_seconds_total[5m]) * 100"

// Memory usage
"process_resident_memory_bytes / process_virtual_memory_bytes * 100"

// Query latency
"rate(neo4j_database_system_check_point_duration_seconds_total[5m]) / rate(neo4j_database_system_check_point_events_total[5m]) * 1000"
```

### Cache Metrics

```go
// Page cache hit rate
"rate(neo4j_page_cache_hits_total[5m]) / (rate(neo4j_page_cache_hits_total[5m]) + rate(neo4j_page_cache_faults_total[5m])) * 100"

// Page cache size
"neo4j_page_cache_bytes_total"
```

### Connection Metrics

```go
// Active connections
"bolt_connections_opened - bolt_connections_closed"

// Connection rate
"rate(bolt_connections_opened[5m])"
```

## API Reference

### PrometheusService Interface

```go
type PrometheusService interface {
    // Query executes an instant query
    Query(prometheusURL string, query string) (*PrometheusQueryResponse, error)
    
    // QueryRange executes a range query
    QueryRange(prometheusURL string, query string, start, end time.Time, step string) (*PrometheusRangeQueryResponse, error)
    
    // GetInstanceHealth retrieves comprehensive health metrics
    GetInstanceHealth(instanceID string, prometheusURL string) (*PrometheusHealthMetrics, error)
}
```

### Response Types

#### PrometheusQueryResponse

```go
type PrometheusQueryResponse struct {
    Status    string              `json:"status"`
    Data      PrometheusQueryData `json:"data"`
    Error     string              `json:"error,omitempty"`
    ErrorType string              `json:"errorType,omitempty"`
}

type PrometheusQueryData struct {
    ResultType string             `json:"resultType"`
    Result     []PrometheusResult `json:"result"`
}

type PrometheusResult struct {
    Metric map[string]string `json:"metric"`
    Value  []interface{}     `json:"value"` // [timestamp, value]
}
```

#### PrometheusHealthMetrics

```go
type PrometheusHealthMetrics struct {
    InstanceID      string                 `json:"instance_id"`
    Timestamp       time.Time              `json:"timestamp"`
    Resources       ResourceMetrics        `json:"resources"`
    Query           QueryMetrics           `json:"query"`
    Connections     ConnectionMetrics      `json:"connections"`
    Storage         StorageMetrics         `json:"storage"`
    OverallStatus   string                 `json:"overall_status"`
    Issues          []string               `json:"issues"`
    Recommendations []string               `json:"recommendations"`
}
```

## Health Status

The `GetInstanceHealth` method returns an overall health status:

- **healthy**: All metrics are within normal ranges
- **warning**: One or more metrics are outside normal ranges

### Health Checks

The health assessment checks:

1. **CPU Usage**: Warning if > 80%
2. **Memory Usage**: Warning if > 85%
3. **Connection Pool**: Warning if > 80% utilization
4. **Page Cache**: Warning if hit rate < 50%

### Recommendations

The system provides actionable recommendations based on the detected issues:

- High CPU/Memory: Suggests scaling to larger instance
- High connections: Suggests reviewing connection pooling
- Low cache hit rate: Suggests increasing page cache size

## Error Handling

All Prometheus operations return standard Go errors. Always check for errors:

```go
resp, err := client.Prometheus.Query(prometheusURL, "up")
if err != nil {
    // Handle error
    log.Printf("Query failed: %v", err)
    return
}

// Use response
```

## Best Practices

1. **Cache Prometheus URLs**: The metrics URL doesn't change for an instance
2. **Use Appropriate Step Sizes**: For range queries, choose step sizes that match your data resolution needs
3. **Monitor Regularly**: Set up periodic health checks to catch issues early
4. **Handle Errors Gracefully**: Prometheus endpoints may be temporarily unavailable

## Examples

See the [examples directory](../example/prometheus_example.go) for complete working examples.

## Authentication

The Prometheus client uses the same authentication as the Aura API client. The Prometheus endpoints use the same credentials (Client ID and Client Secret) that you use for the Aura API.

The client automatically handles authentication, token management, and retries.

## License

See the main package [LICENSE](../LICENSE) file.
