# Aura API Client

## Overview

A Go package that enables the use of Neo4j Aura API in a friendly way e.g `client.Instances.List(ctx)` to return a list of instances in Aura.

Client Id and Secret are required and these can be obtained from the [Neo4j Aura Console](https://neo4j.com/docs/aura/api/authentication/).

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Context and Timeouts](#context-and-timeouts)
- [Tenant Operations](#tenant-operations)
- [Instance Operations](#instance-operations)
- [Snapshot Operations](#snapshot-operations)
- [CMEK Operations](#cmek-operations)
- [GDS Session Operations](#gds-session-operations)
- [Prometheus Metrics Operations](#prometheus-metrics-operations)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)
- [Migration from v1.x](#migration-from-v1x)

---

## Installation

```bash
go get github.com/LackOfMorals/aura-client
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    aura "github.com/LackOfMorals/aura-client"
)

func main() {
    client, err := aura.NewClient(
        aura.WithCredentials("your-client-id", "your-client-secret"),
    )
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    ctx := context.Background()

    instances, err := client.Instances.List(ctx)
    if err != nil {
        log.Fatalf("Failed to list instances: %v", err)
    }

    for _, instance := range instances.Data {
        log.Printf("Instance: %s (ID: %s)\n", instance.Name, instance.Id)
    }
}
```

---

## Configuration

### Simple Configuration

```go
client, err := aura.NewClient(
    aura.WithCredentials("client-id", "client-secret"),
)
```

### Advanced Configuration

```go
client, err := aura.NewClient(
    aura.WithCredentials("client-id", "client-secret"),
    aura.WithTimeout(60 * time.Second),
    aura.WithMaxRetry(5),
)
```

### Custom Logger

```go
import "log/slog"

opts := &slog.HandlerOptions{Level: slog.LevelDebug}
handler := slog.NewTextHandler(os.Stderr, opts)
logger := slog.New(handler)

client, err := aura.NewClient(
    aura.WithCredentials("client-id", "client-secret"),
    aura.WithLogger(logger),
)
```

### Targeting a Different Base URL

Use `WithBaseURL` to point the client at a staging or sandbox environment:

```go
client, err := aura.NewClient(
    aura.WithCredentials("client-id", "client-secret"),
    aura.WithBaseURL("https://api.staging.neo4j.io"),
)
```

---

## Context and Timeouts

Every service method accepts a `context.Context` as its first argument. This is the standard Go pattern and gives you full control over cancellation and deadlines on a per-call basis.

The client is configured with a default timeout (120 seconds, overridable with `WithTimeout`). This timeout is applied as a ceiling on each call — if the context you pass already has a shorter deadline, that shorter deadline wins.

### Basic usage

```go
ctx := context.Background()
instances, err := client.Instances.List(ctx)
```

### Per-call deadline

```go
// This specific call must complete within 10 seconds
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

instance, err := client.Instances.Get(ctx, "instance-id")
```

### Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

// Cancel all in-flight calls (e.g. on OS signal or user action)
go func() {
    <-shutdownSignal
    cancel()
}()

instances, err := client.Instances.List(ctx)
if err != nil {
    if ctx.Err() == context.Canceled {
        log.Println("Request was cancelled")
    }
}
```

### Distributed tracing

Because context flows through every call, you can attach trace spans from any OpenTelemetry-compatible library:

```go
ctx, span := tracer.Start(r.Context(), "list-instances")
defer span.End()

instances, err := client.Instances.List(ctx)
```

---

## Tenant Operations

### List All Tenants

```go
ctx := context.Background()

tenants, err := client.Tenants.List(ctx)
if err != nil {
    log.Fatalf("Error: %v", err)
}

for _, tenant := range tenants.Data {
    fmt.Printf("Tenant: %s (ID: %s)\n", tenant.Name, tenant.Id)
}
```

### Get Tenant Details

```go
ctx := context.Background()

tenant, err := client.Tenants.Get(ctx, "your-tenant-id")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Tenant: %s\n", tenant.Data.Name)
fmt.Printf("Available instance configurations:\n")

for _, config := range tenant.Data.InstanceConfigurations {
    fmt.Printf("  - %s in %s: %s memory, Type: %s\n",
        config.CloudProvider,
        config.RegionName,
        config.Memory,
        config.Type,
    )
}
```

---

## Instance Operations

### List All Instances

```go
ctx := context.Background()

instances, err := client.Instances.List(ctx)
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Found %d instances:\n", len(instances.Data))
for _, instance := range instances.Data {
    fmt.Printf("  - %s (ID: %s) on %s\n",
        instance.Name,
        instance.Id,
        instance.CloudProvider,
    )
}
```

### Get Instance Details

```go
ctx := context.Background()

instance, err := client.Instances.Get(ctx, "your-instance-id")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Instance: %s\n", instance.Data.Name)
fmt.Printf("Status: %s\n", instance.Data.Status)
fmt.Printf("Connection URL: %s\n", instance.Data.ConnectionUrl)
fmt.Printf("Memory: %s\n", instance.Data.Memory)
fmt.Printf("Type: %s\n", instance.Data.Type)
fmt.Printf("Region: %s\n", instance.Data.Region)
```

### Create a New Instance

```go
ctx := context.Background()

config := &aura.CreateInstanceConfigData{
    Name:          "my-neo4j-db",
    TenantId:      "your-tenant-id",
    CloudProvider: "gcp",
    Region:        "europe-west1",
    Type:          "enterprise-db",
    Version:       "5",
    Memory:        "8GB",
}

instance, err := client.Instances.Create(ctx, config)
if err != nil {
    log.Fatalf("Error creating instance: %v", err)
}

fmt.Printf("Instance created!\n")
fmt.Printf("  ID: %s\n", instance.Data.Id)
fmt.Printf("  Connection URL: %s\n", instance.Data.ConnectionUrl)
fmt.Printf("  Username: %s\n", instance.Data.Username)
fmt.Printf("  Password: %s\n", instance.Data.Password)

// ⚠️ IMPORTANT: Save these credentials securely!
// The password is only shown once during creation.
```

### Update an Instance

```go
ctx := context.Background()

updateData := &aura.UpdateInstanceData{
    Name:   "my-renamed-instance",
    Memory: "16GB",
}

instance, err := client.Instances.Update(ctx, "instance-id", updateData)
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Instance updated: %s with %s memory\n",
    instance.Data.Name,
    instance.Data.Memory,
)
```

### Pause an Instance

```go
ctx := context.Background()

instance, err := client.Instances.Pause(ctx, "instance-id")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Instance paused. Status: %s\n", instance.Data.Status)
```

### Resume an Instance

```go
ctx := context.Background()

instance, err := client.Instances.Resume(ctx, "instance-id")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Instance resumed. Status: %s\n", instance.Data.Status)
```

### Delete an Instance

```go
ctx := context.Background()

// ⚠️ WARNING: This is irreversible!
instance, err := client.Instances.Delete(ctx, "instance-to-delete")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Instance %s deleted\n", instance.Data.Id)
```

### Overwrite Instance from Another Instance

```go
ctx := context.Background()

result, err := client.Instances.Overwrite(ctx, "target-instance-id", "source-instance-id", "")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Overwrite initiated: %s\n", result.Data)
// Note: This is asynchronous. Monitor instance status.
```

### Overwrite Instance from Snapshot

```go
ctx := context.Background()

result, err := client.Instances.Overwrite(ctx, "target-instance-id", "", "snapshot-id")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Overwrite from snapshot initiated\n")
```

---

## Snapshot Operations

### List Snapshots

```go
ctx := context.Background()

// Empty date string returns today's snapshots
snapshots, err := client.Snapshots.List(ctx, "your-instance-id", "")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Found %d snapshots:\n", len(snapshots.Data))
for _, snapshot := range snapshots.Data {
    fmt.Printf("  - ID: %s, Profile: %s, Status: %s\n",
        snapshot.SnapshotId,
        snapshot.Profile,
        snapshot.Status,
    )
}
```

### List Snapshots for a Specific Date

```go
ctx := context.Background()

snapshots, err := client.Snapshots.List(ctx, "your-instance-id", "2024-01-15")
if err != nil {
    log.Fatalf("Error: %v", err)
}

for _, snapshot := range snapshots.Data {
    fmt.Printf("  - %s at %s\n", snapshot.SnapshotId, snapshot.Timestamp)
}
```

### Get Snapshot Details

```go
ctx := context.Background()

snapshot, err := client.Snapshots.Get(ctx, "your-instance-id", "your-snapshot-id")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Instance ID: %s\nSnapshot ID: %s\nStatus: %s\nTimestamp: %s\n",
    snapshot.Data.InstanceId,
    snapshot.Data.SnapshotId,
    snapshot.Data.Status,
    snapshot.Data.Timestamp,
)
```

### Create an On-Demand Snapshot

```go
ctx := context.Background()

snapshot, err := client.Snapshots.Create(ctx, "your-instance-id")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Snapshot creation initiated. Snapshot ID: %s\n", snapshot.Data.SnapshotId)
// Note: Snapshot creation is asynchronous. Poll List() to check completion status.
```

### Restore from a Snapshot

```go
ctx := context.Background()

result, err := client.Snapshots.Restore(ctx, "your-instance-id", "your-snapshot-id")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Instance ID: %s\nStatus: %s\n", result.Data.Id, result.Data.Status)
```

---

## CMEK Operations

### List Customer Managed Encryption Keys

```go
ctx := context.Background()

// Pass an empty string to list all CMEKs regardless of tenant
cmeks, err := client.Cmek.List(ctx, "")
if err != nil {
    log.Fatalf("Error: %v", err)
}

for _, cmek := range cmeks.Data {
    fmt.Printf("  - %s (ID: %s) in tenant %s\n", cmek.Name, cmek.Id, cmek.TenantId)
}
```

### Filter CMEKs by Tenant

```go
ctx := context.Background()

cmeks, err := client.Cmek.List(ctx, "your-tenant-id")
if err != nil {
    log.Fatalf("Error: %v", err)
}

for _, cmek := range cmeks.Data {
    fmt.Printf("  - %s\n", cmek.Name)
}
```

---

## GDS Session Operations

### List Graph Data Science Sessions

```go
ctx := context.Background()

sessions, err := client.GraphAnalytics.List(ctx)
if err != nil {
    log.Fatalf("Error: %v", err)
}

for _, session := range sessions.Data {
    fmt.Printf("  - %s (ID: %s)\n", session.Name, session.Id)
    fmt.Printf("    Memory: %s, Status: %s\n", session.Memory, session.Status)
    fmt.Printf("    Instance: %s, Expires: %s\n", session.InstanceId, session.Expiry)
}
```

---

## Prometheus Metrics Operations

Each Aura instance exposes Prometheus metrics for monitoring.

### Get the Prometheus URL for an Instance

```go
ctx := context.Background()

instance, err := client.Instances.Get(ctx, "your-instance-id")
if err != nil {
    log.Fatalf("Error: %v", err)
}

prometheusURL := instance.Data.MetricsURL
```

### Get Instance Health Metrics

```go
ctx := context.Background()

health, err := client.Prometheus.GetInstanceHealth(ctx, "your-instance-id", prometheusURL)
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Health Status: %s\n", health.OverallStatus)
fmt.Printf("CPU Usage: %.2f%%\n", health.Resources.CPUUsagePercent)
fmt.Printf("Memory Usage: %.2f%%\n", health.Resources.MemoryUsagePercent)
fmt.Printf("Queries/sec: %.2f\n", health.Query.QueriesPerSecond)
fmt.Printf("Active Connections: %d/%d (%.1f%%)\n",
    health.Connections.ActiveConnections,
    health.Connections.MaxConnections,
    health.Connections.UsagePercent,
)

if len(health.Issues) > 0 {
    fmt.Println("\nIssues detected:")
    for _, issue := range health.Issues {
        fmt.Printf("  - %s\n", issue)
    }
}

if len(health.Recommendations) > 0 {
    fmt.Println("\nRecommendations:")
    for _, rec := range health.Recommendations {
        fmt.Printf("  - %s\n", rec)
    }
}
```

For more detailed information on Prometheus operations, see the [Prometheus documentation](./docs/prometheus.md).

---

## Error Handling

### Basic Error Handling

```go
ctx := context.Background()

instance, err := client.Instances.Get(ctx, "instance-id")
if err != nil {
    log.Printf("Error: %v\n", err)
    return
}
```

### Typed API Errors

```go
ctx := context.Background()

instance, err := client.Instances.Get(ctx, "non-existent-id")
if err != nil {
    if apiErr, ok := err.(*aura.Error); ok {
        fmt.Printf("API Error %d: %s\n", apiErr.StatusCode, apiErr.Message)

        switch {
        case apiErr.IsNotFound():
            fmt.Println("Instance not found")
        case apiErr.IsUnauthorized():
            fmt.Println("Authentication failed - check credentials")
        case apiErr.IsBadRequest():
            fmt.Println("Invalid request parameters")
        }

        if apiErr.HasMultipleErrors() {
            fmt.Println("All errors:")
            for _, msg := range apiErr.AllErrors() {
                fmt.Printf("  - %s\n", msg)
            }
        }
        return
    }

    log.Printf("Unexpected error: %v\n", err)
    return
}
```

### Context Errors

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

instances, err := client.Instances.List(ctx)
if err != nil {
    switch ctx.Err() {
    case context.DeadlineExceeded:
        log.Println("Request timed out")
    case context.Canceled:
        log.Println("Request was cancelled")
    default:
        log.Printf("Error: %v\n", err)
    }
    return
}
```

---

## Best Practices

### 1. Secure Credential Management

```go
clientID := os.Getenv("AURA_CLIENT_ID")
clientSecret := os.Getenv("AURA_CLIENT_SECRET")

if clientID == "" || clientSecret == "" {
    log.Fatal("Missing AURA credentials in environment")
}

client, err := aura.NewClient(
    aura.WithCredentials(clientID, clientSecret),
)
```

### 2. Save Instance Credentials Immediately After Creation

```go
ctx := context.Background()

instance, err := client.Instances.Create(ctx, config)
if err != nil {
    log.Fatal(err)
}

// ⚠️ CRITICAL: Save these immediately — they are only shown once!
credentials := map[string]string{
    "instance_id":    instance.Data.Id,
    "connection_url": instance.Data.ConnectionUrl,
    "username":       instance.Data.Username,
    "password":       instance.Data.Password,
}
// Store in a secrets manager. Do NOT log passwords in production.
```

### 3. Polling for Async Operations

```go
ctx := context.Background()

instanceID := newInstance.Data.Id

for range 30 {
    inst, err := client.Instances.Get(ctx, instanceID)
    if err != nil {
        log.Printf("Error checking status: %v", err)
    } else if inst.Data.Status == aura.StatusRunning {
        fmt.Println("Instance is ready!")
        break
    } else {
        fmt.Printf("Status: %s, waiting...\n", inst.Data.Status)
    }
    time.Sleep(10 * time.Second)
}
```

### 4. Graceful Shutdown

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigChan
    fmt.Println("\nShutting down gracefully...")
    cancel()
}()

// Pass ctx to any in-flight calls — they will be cancelled on signal
instances, err := client.Instances.List(ctx)
```

### 5. Retry Logic for Transient Failures

```go
func retryOperation(maxRetries int, fn func() error) error {
    var err error
    for i := range maxRetries {
        err = fn()
        if err == nil {
            return nil
        }

        if apiErr, ok := err.(*aura.Error); ok {
            // Don't retry client errors (4xx except 429 Too Many Requests)
            if apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != 429 {
                return err
            }
        }

        wait := time.Duration(math.Pow(2, float64(i))) * time.Second
        fmt.Printf("Attempt %d failed, retrying in %v...\n", i+1, wait)
        time.Sleep(wait)
    }
    return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}

// Usage
ctx := context.Background()
err := retryOperation(3, func() error {
    _, err := client.Instances.List(ctx)
    return err
})
```

---

## Complete Example Application

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    aura "github.com/LackOfMorals/aura-client"
)

func main() {
    clientID := os.Getenv("AURA_CLIENT_ID")
    clientSecret := os.Getenv("AURA_CLIENT_SECRET")
    tenantID := os.Getenv("AURA_TENANT_ID")

    if clientID == "" || clientSecret == "" {
        log.Fatal("Missing required environment variables")
    }

    client, err := aura.NewClient(
        aura.WithCredentials(clientID, clientSecret),
        aura.WithTimeout(120 * time.Second),
    )
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    ctx := context.Background()

    fmt.Println("=== Current Instances ===")
    instances, err := client.Instances.List(ctx)
    if err != nil {
        log.Fatalf("Failed to list instances: %v", err)
    }

    for _, inst := range instances.Data {
        fmt.Printf("- %s: %s (%s)\n", inst.Name, inst.Id, inst.CloudProvider)
    }

    if tenantID != "" {
        fmt.Println("\n=== Tenant Configuration ===")
        tenant, err := client.Tenants.Get(ctx, tenantID)
        if err != nil {
            log.Printf("Warning: Could not get tenant: %v", err)
        } else {
            fmt.Printf("Tenant: %s\n", tenant.Data.Name)
            fmt.Printf("Available configurations: %d\n", len(tenant.Data.InstanceConfigurations))
        }
    }

    fmt.Println("\n✓ Client is working correctly!")
}
```

Run with:
```bash
export AURA_CLIENT_ID="your-client-id"
export AURA_CLIENT_SECRET="your-client-secret"
export AURA_TENANT_ID="your-tenant-id"
go run main.go
```

---

## Migration from v1.x

### Breaking Changes in v2.0

#### 1. `context.Context` is now required on every service call

The most significant change. Contexts are no longer stored inside the client at construction time — they flow through each individual call instead. This follows the standard Go convention and enables per-call cancellation, deadlines, and distributed tracing.

**Before (v1.x):**
```go
client, _ := aura.NewClient(
    aura.WithCredentials(id, secret),
    aura.   // ❌ Removed
)

instances, err := client.Instances.List()
tenant, err := client.Tenants.Get(tenantID)
```

**After (v2.0):**
```go
client, _ := aura.NewClient(
    aura.WithCredentials(id, secret),
    // No WithContext — pass ctx to each call instead
)

ctx := context.Background()

instances, err := client.Instances.List(ctx)
tenant, err := client.Tenants.Get(ctx, tenantID)
```

The quickest way to find all call sites is:

```bash
grep -rn "client\.\(Instances\|Tenants\|Snapshots\|Cmek\|GraphAnalytics\|Prometheus\)\." ./
```

#### 2. `WithContext` option removed

`WithContext` has been removed from `NewClient`. Pass a context directly to each service method instead (see above).

#### 3. `WithBaseURL` option added

A new `WithBaseURL` option is available for targeting non-production environments:

```go
client, _ := aura.NewClient(
    aura.WithCredentials(id, secret),
    aura.WithBaseURL("https://api.staging.neo4j.io"),
)
```

### Migration Steps

#### Step 1: Update the dependency

```bash
go get github.com/LackOfMorals/aura-client@v2.0.0
go mod tidy
```

#### Step 2: Remove `WithContext` from `NewClient`

```go
// Before
client, _ := aura.NewClient(
    aura.WithCredentials(id, secret),
    aura.  // remove this line
)

// After
client, _ := aura.NewClient(
    aura.WithCredentials(id, secret),
)
```

#### Step 3: Add `ctx` to every service call

Add `ctx` as the first argument to every method call. If you don't have a specific context, use `context.Background()`:

```go
ctx := context.Background()

// Before
instances, err := client.Instances.List()
instance, err := client.Instances.Get(id)
instance, err := client.Instances.Create(config)
instance, err := client.Instances.Delete(id)
instance, err := client.Instances.Pause(id)
instance, err := client.Instances.Resume(id)
instance, err := client.Instances.Update(id, data)
result, err  := client.Instances.Overwrite(id, srcID, snapID)

tenants, err := client.Tenants.List()
tenant, err  := client.Tenants.Get(id)
metrics, err := client.Tenants.GetMetrics(id)

snapshots, err := client.Snapshots.List(id, date)
snapshot, err  := client.Snapshots.Get(id, snapID)
snapshot, err  := client.Snapshots.Create(id)
result, err    := client.Snapshots.Restore(id, snapID)

cmeks, err := client.Cmek.List(tenantID)

sessions, err := client.GraphAnalytics.List()
session, err  := client.GraphAnalytics.Get(id)
session, err  := client.GraphAnalytics.Create(config)
estimate, err := client.GraphAnalytics.Estimate(req)
result, err   := client.GraphAnalytics.Delete(id)

raw, err    := client.Prometheus.FetchRawMetrics(url)
val, err    := client.Prometheus.GetMetricValue(raw, name, filters)
health, err := client.Prometheus.GetInstanceHealth(id, url)

// After
instances, err := client.Instances.List(ctx)
instance, err := client.Instances.Get(ctx, id)
instance, err := client.Instances.Create(ctx, config)
instance, err := client.Instances.Delete(ctx, id)
instance, err := client.Instances.Pause(ctx, id)
instance, err := client.Instances.Resume(ctx, id)
instance, err := client.Instances.Update(ctx, id, data)
result, err  := client.Instances.Overwrite(ctx, id, srcID, snapID)

tenants, err := client.Tenants.List(ctx)
tenant, err  := client.Tenants.Get(ctx, id)
metrics, err := client.Tenants.GetMetrics(ctx, id)

snapshots, err := client.Snapshots.List(ctx, id, date)
snapshot, err  := client.Snapshots.Get(ctx, id, snapID)
snapshot, err  := client.Snapshots.Create(ctx, id)
result, err    := client.Snapshots.Restore(ctx, id, snapID)

cmeks, err := client.Cmek.List(ctx, tenantID)

sessions, err := client.GraphAnalytics.List(ctx)
session, err  := client.GraphAnalytics.Get(ctx, id)
session, err  := client.GraphAnalytics.Create(ctx, config)
estimate, err := client.GraphAnalytics.Estimate(ctx, req)
result, err   := client.GraphAnalytics.Delete(ctx, id)

raw, err    := client.Prometheus.FetchRawMetrics(ctx, url)
val, err    := client.Prometheus.GetMetricValue(ctx, raw, name, filters)
health, err := client.Prometheus.GetInstanceHealth(ctx, id, url)
```

#### Step 4: Verify

```bash
go build ./...
go test ./...
```

### Quick Migration Checklist

- [ ] `go get github.com/LackOfMorals/aura-client@v2.0.0` and `go mod tidy`
- [ ] Remove `aura.WithContext(...)` from all `NewClient` calls
- [ ] Add `ctx` as first argument to every service method call
- [ ] Ensure `context` is imported wherever service calls are made
- [ ] `go build ./...` — fix any remaining compilation errors
- [ ] `go test ./...`

---

## Additional Resources

- [Neo4j Aura API Documentation](https://neo4j.com/docs/aura/platform/api/)
- [GitHub Repository](https://github.com/LackOfMorals/aura-client)
- [Report Issues](https://github.com/LackOfMorals/aura-client/issues)
- [Prometheus Metrics Guide](./docs/prometheus.md)

---

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

See [LICENSE](LICENSE) file for details.
