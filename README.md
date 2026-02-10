# Aura API Client

## Overview

A Go package that enables the use of Neo4j Aura API in a friendly way e.g `instances.List()` to return a list of instances in Aura. 

Client Id and Secret are required and these can be obtained from the [Neo4j Aura Console](https://neo4j.com/docs/aura/api/authentication/).

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
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

### Basic Setup

```go
package main

import (
    "log"
    aura "github.com/LackOfMorals/aura-client"
)

func main() {
    // Create client with credentials
    client, err := aura.NewClient(
        aura.WithCredentials("your-client-id", "your-client-secret"),
    )
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // List all instances
    instances, err := client.Instances.List()
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
    aura.WithContext(context.Background()),
    aura.WithMaxRetry(5),
)
```

### Custom Logger

```go
import "log/slog"

// Create custom logger with debug level
opts := &slog.HandlerOptions{Level: slog.LevelDebug}
handler := slog.NewTextHandler(os.Stderr, opts)
logger := slog.New(handler)

client, err := aura.NewClient(
    aura.WithCredentials("client-id", "client-secret"),
    aura.WithLogger(logger),
)
```

---

## Tenant Operations

### List All Tenants

```go
tenants, err := client.Tenants.List()
if err != nil {
    log.Fatalf("Error: %v", err)
}

for _, tenant := range tenants.Data {
    fmt.Printf("Tenant: %s (ID: %s)\n", tenant.Name, tenant.Id)
}
```

### Get Tenant Details

```go
tenantID := "your-tenant-id"
tenant, err := client.Tenants.Get(tenantID)
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
instances, err := client.Instances.List()
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
instanceID := "your-instance-id"
instance, err := client.Instances.Get(instanceID)
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
config := &aura.CreateInstanceConfigData{
    Name:          "my-neo4j-db",
    TenantId:      "your-tenant-id",
    CloudProvider: "gcp",
    Region:        "europe-west1",
    Type:          "enterprise-db",
    Version:       "5",
    Memory:        "8GB",
}

instance, err := client.Instances.Create(config)
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
updateData := &aura.UpdateInstanceData{
    Name:   "my-renamed-instance",
    Memory: "16GB",  // Scale up memory
}

instance, err := client.Instances.Update("instance-id", updateData)
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
instance, err := client.Instances.Pause("instance-id")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Instance paused. Status: %s\n", instance.Data.Status)
```

### Resume an Instance

```go
instance, err := client.Instances.Resume("instance-id")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Instance resumed. Status: %s\n", instance.Data.Status)
```

### Delete an Instance

```go
// ⚠️ WARNING: This is irreversible!
instanceID := "instance-to-delete"

instance, err := client.Instances.Delete(instanceID)
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Instance %s deleted\n", instance.Data.Id)
```

### Overwrite Instance from Another Instance

```go
// Restore targetInstance from sourceInstance
targetID := "target-instance-id"
sourceID := "source-instance-id"

result, err := client.Instances.Overwrite(targetID, sourceID, "")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Overwrite initiated: %s\n", result.Data)
// Note: This is asynchronous. Monitor instance status.
```

### Overwrite Instance from Snapshot

```go
targetID := "target-instance-id"
snapshotID := "snapshot-id"

result, err := client.Instances.Overwrite(targetID, "", snapshotID)
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Overwrite from snapshot initiated\n")
```

---

## Snapshot Operations

### List Snapshots

```go
instanceID := "your-instance-id"

// Empty date string returns today's snapshots
snapshots, err := client.Snapshots.List(instanceID, "")
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

### List Snapshots for Specific Date

```go
instanceID := "your-instance-id"
date := "2024-01-15"  // Format: YYYY-MM-DD

snapshots, err := client.Snapshots.List(instanceID, date)
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Snapshots for %s:\n", date)
for _, snapshot := range snapshots.Data {
    fmt.Printf("  - %s at %s\n", 
        snapshot.SnapshotId, 
        snapshot.Timestamp,
    )
}
```

### Get the details of a Snapshot

```go
instanceID := "your-instance-id"
snapshotID := "your-snapshot-id"

snapshot, err := client.Snapshots.Get(instanceID, snapshotID)
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Snapshot details: \n Instance ID: %s \n Snapshot ID: %s \n Status: %s \n Timestamp: %s ", 
    snapshot.Data.InstanceId, 
    snapshot.Data.SnapshotId, 
    snapshot.Data.Status,
    snapshot.Data.Timestamp,
)
```

### Create an On-Demand Snapshot

```go
instanceID := "your-instance-id"

snapshot, err := client.Snapshots.Create(instanceID)
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Snapshot creation initiated!\n")
fmt.Printf("Snapshot ID: %s\n", snapshot.Data.SnapshotId)

// Note: Snapshot creation is asynchronous
// Poll with List() to check completion status
```

### Restore from a snapshot

```go
instanceID := "your-instance-id"
snapshotID := "your-snapshot-id"

result, err := client.Snapshots.Restore(instanceID, snapshotID)
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Snapshot details: \n Instance ID: %s \n Status: %s", 
    result.Data.InstanceId, 
    result.Data.Status,
)
```

---

## CMEK Operations

### List Customer Managed Encryption Keys

```go
// List all CMEKs
cmeks, err := client.Cmek.List("")
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Found %d CMEK(s):\n", len(cmeks.Data))
for _, cmek := range cmeks.Data {
    fmt.Printf("  - %s (ID: %s) in tenant %s\n",
        cmek.Name,
        cmek.Id,
        cmek.TenantId,
    )
}
```

### Filter CMEKs by Tenant

```go
tenantID := "your-tenant-id"
cmeks, err := client.Cmek.List(tenantID)
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("CMEKs in tenant %s:\n", tenantID)
for _, cmek := range cmeks.Data {
    fmt.Printf("  - %s\n", cmek.Name)
}
```

---

## GDS Session Operations

### List Graph Data Science Sessions

```go
sessions, err := client.GraphAnalytics.List()
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Found %d GDS session(s):\n", len(sessions.Data))
for _, session := range sessions.Data {
    fmt.Printf("  - %s (ID: %s)\n", session.Name, session.Id)
    fmt.Printf("    Memory: %s, Status: %s\n", 
        session.Memory, 
        session.Status,
    )
    fmt.Printf("    Instance: %s\n", session.InstanceId)
    fmt.Printf("    Expires: %s\n", session.Expiry)
}
```

---

## Prometheus Metrics Operations

### Query Prometheus Metrics

Each Aura instance exposes Prometheus metrics for monitoring. The client provides a convenient way to query these metrics.

```go
// Get instance details to retrieve the Prometheus URL
instanceID := "your-instance-id"
instance, err := client.Instances.Get(instanceID)
if err != nil {
    log.Fatalf("Error: %v", err)
}

prometheusURL := instance.Data.MetricsURL
// e.g., "https://c9f0d13a.metrics.neo4j.io/prometheus"
```

### Get Instance Health Metrics

```go
// Get comprehensive health metrics for an instance
health, err := client.Prometheus.GetInstanceHealth(instanceID, prometheusURL)
if err != nil {
    log.Fatalf("Error: %v", err)
}

fmt.Printf("Instance Health Status: %s\n", health.OverallStatus)
fmt.Printf("CPU Usage: %.2f%%\n", health.Resources.CPUUsagePercent)
fmt.Printf("Memory Usage: %.2f%%\n", health.Resources.MemoryUsagePercent)
fmt.Printf("Queries/sec: %.2f\n", health.Query.QueriesPerSecond)
fmt.Printf("Active Connections: %d/%d (%.1f%%)\n",
    health.Connections.ActiveConnections,
    health.Connections.MaxConnections,
    health.Connections.UsagePercent)

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
instance, err := client.Instances.Get("instance-id")
if err != nil {
    log.Printf("Error: %v\n", err)
    return
}
```

### Advanced Error Handling with Custom API Errors

```go
instance, err := client.Instances.Get("non-existent-id")
if err != nil {
    // Type assert to Error for detailed information
    if apiErr, ok := err.(*api.Error); ok {
        fmt.Printf("API Error %d: %s\n", 
            apiErr.StatusCode, 
            apiErr.Message,
        )
        
        // Check specific error types
        switch {
        case apiErr.IsNotFound():
            fmt.Println("Instance not found")
        case apiErr.IsUnauthorized():
            fmt.Println("Authentication failed - check credentials")
        case apiErr.IsBadRequest():
            fmt.Println("Invalid request parameters")
        }
        
        // Handle multiple errors
        if apiErr.HasMultipleErrors() {
            fmt.Println("Multiple errors occurred:")
            for _, errMsg := range apiErr.AllErrors() {
                fmt.Printf("  - %s\n", errMsg)
            }
        }
        
        return
    }
    
    // Some other error type
    log.Printf("Unexpected error: %v\n", err)
    return
}

fmt.Printf("Success: %s\n", instance.Data.Name)
```

### Context-Based Timeout Handling

```go
// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Initialize client with context
client, err := aura.NewClient(
    aura.WithCredentials("client-id", "client-secret"),
    aura.WithContext(ctx),
)
if err != nil {
    log.Fatal(err)
}

instances, err := client.Instances.List()
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("Request timed out after 30 seconds")
    } else {
        log.Printf("Error: %v\n", err)
    }
    return
}
```

---

## Best Practices

### 1. Secure Credential Management

```go
import "os"

// Load credentials from environment variables
clientID := os.Getenv("AURA_CLIENT_ID")
clientSecret := os.Getenv("AURA_CLIENT_SECRET")

if clientID == "" || clientSecret == "" {
    log.Fatal("Missing AURA credentials in environment")
}

client, err := aura.NewClient(
    aura.WithCredentials(clientID, clientSecret),
)
```

### 2. Save Instance Credentials Securely

```go
instance, err := client.Instances.Create(config)
if err != nil {
    log.Fatal(err)
}

// ⚠️ CRITICAL: Save these immediately - they're only shown once!
credentials := map[string]string{
    "instance_id":    instance.Data.Id,
    "connection_url": instance.Data.ConnectionUrl,
    "username":       instance.Data.Username,
    "password":       instance.Data.Password,
}

// Save to secure storage (e.g., environment variables, secrets manager)
// DO NOT log or print passwords in production!
```

### 3. Polling for Async Operations

```go
// After creating an instance, poll for readiness
instanceID := instance.Data.Id
maxAttempts := 30
waitTime := 10 * time.Second

for i := 0; i < maxAttempts; i++ {
    inst, err := client.Instances.Get(instanceID)
    if err != nil {
        log.Printf("Error checking status: %v", err)
        continue
    }
    
    if inst.Data.Status == "running" {
        fmt.Println("Instance is ready!")
        break
    }
    
    fmt.Printf("Status: %s, waiting...\n", inst.Data.Status)
    time.Sleep(waitTime)
}
```

### 4. Graceful Shutdown

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Listen for interrupt signals
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigChan
    fmt.Println("\nShutting down gracefully...")
    cancel()
}()

client, err := aura.NewClient(
    aura.WithCredentials(clientID, clientSecret),
    aura.WithContext(ctx),
)
```

### 5. Retry Logic for Transient Failures

```go
func retryOperation(maxRetries int, fn func() error) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = fn()
        if err == nil {
            return nil
        }
        
        // Check if error is retryable
        if apiErr, ok := err.(*api.Error); ok {
            // Don't retry client errors (4xx except 429)
            if apiErr.StatusCode >= 400 && 
               apiErr.StatusCode < 500 && 
               apiErr.StatusCode != 429 {
                return err
            }
        }
        
        // Exponential backoff
        waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
        fmt.Printf("Attempt %d failed, retrying in %v...\n", i+1, waitTime)
        time.Sleep(waitTime)
    }
    
    return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}

// Usage
err := retryOperation(3, func() error {
    _, err := client.Instances.List()
    return err
})
```

---

## Complete Example Application

```go
package main

import (
    "fmt"
    "log"
    "os"
    "time"
    
    aura "github.com/LackOfMorals/aura-client"
)

func main() {
    // Load credentials from environment
    clientID := os.Getenv("AURA_CLIENT_ID")
    clientSecret := os.Getenv("AURA_CLIENT_SECRET")
    tenantID := os.Getenv("AURA_TENANT_ID")
    
    if clientID == "" || clientSecret == "" {
        log.Fatal("Missing required environment variables")
    }
    
    // Create client
    client, err := aura.NewClient(
        aura.WithCredentials(clientID, clientSecret),
        aura.WithTimeout(120 * time.Second),
    )
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    // List existing instances
    fmt.Println("=== Current Instances ===")
    instances, err := client.Instances.List()
    if err != nil {
        log.Fatalf("Failed to list instances: %v", err)
    }
    
    for _, inst := range instances.Data {
        fmt.Printf("- %s: %s (%s)\n", 
            inst.Name, 
            inst.Id, 
            inst.CloudProvider,
        )
    }
    
    // Get tenant details
    if tenantID != "" {
        fmt.Println("\n=== Tenant Configuration ===")
        tenant, err := client.Tenants.Get(tenantID)
        if err != nil {
            log.Printf("Warning: Could not get tenant: %v", err)
        } else {
            fmt.Printf("Tenant: %s\n", tenant.Data.Name)
            fmt.Printf("Available configurations: %d\n", 
                len(tenant.Data.InstanceConfigurations),
            )
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

## Additional Resources

- [Neo4j Aura API Documentation](https://neo4j.com/docs/aura/platform/api/)
- [GitHub Repository](https://github.com/LackOfMorals/aura-client)
- [Report Issues](https://github.com/LackOfMorals/aura-client/issues)

---

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

See [LICENSE](LICENSE) file for details.
