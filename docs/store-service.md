# Configuration Store Service

The Configuration Store service provides persistent storage for Neo4j Aura instance configurations using SQLite. This allows you to save, manage, and reuse instance configurations across sessions.

## Features

- **CRUD Operations**: Create, Read, Update, and Delete instance configurations
- **Persistent Storage**: Configurations are stored in a SQLite database
- **Label-based Access**: Reference configurations using simple string labels
- **List Support**: Retrieve all stored configuration labels
- **Integration**: Seamlessly create Aura instances from stored configurations

## Database Location

By default, the store database is created at:
- **Linux/macOS**: `~/.aura-client/store.db`
- **Windows**: `%USERPROFILE%\.aura-client\store.db`

You can customize this location using the `WithStorePath` option when creating the client.

## Usage

### Basic Setup

```go
import "github.com/LackOfMorals/aura-client"

client, err := aura.NewClient(
    aura.WithCredentials("client-id", "client-secret"),
)
if err != nil {
    log.Fatal(err)
}
```

### Custom Database Path

```go
client, err := aura.NewClient(
    aura.WithCredentials("client-id", "client-secret"),
    aura.WithStorePath("/custom/path/to/store.db"),
)
```

## Operations

### Create - Store a New Configuration

```go
config := &aura.CreateInstanceConfigData{
    Name:          "production-db",
    TenantId:      "my-tenant-id",
    CloudProvider: "gcp",
    Region:        "us-central1",
    Type:          "enterprise-db",
    Version:       "5",
    Memory:        "8GB",
}

err := client.Store.Create("prod-config", config)
if err != nil {
    // Handle error
}
```

**Errors:**
- `ErrConfigAlreadyExists`: Label already exists
- `ErrInvalidLabel`: Label is empty
- `ErrInvalidConfig`: Configuration is nil

### Read - Retrieve a Configuration

```go
config, err := client.Store.Read("prod-config")
if err != nil {
    // Handle error
}

fmt.Printf("Configuration: %s (%s)\n", config.Name, config.CloudProvider)
```

**Errors:**
- `ErrConfigNotFound`: Configuration doesn't exist
- `ErrInvalidLabel`: Label is empty

### Update - Modify an Existing Configuration

```go
updatedConfig := &aura.CreateInstanceConfigData{
    Name:          "production-db-v2",
    TenantId:      "my-tenant-id",
    CloudProvider: "gcp",
    Region:        "us-central1",
    Type:          "enterprise-db",
    Version:       "5",
    Memory:        "16GB", // Updated from 8GB
}

err := client.Store.Update("prod-config", updatedConfig)
if err != nil {
    // Handle error
}
```

**Errors:**
- `ErrConfigNotFound`: Configuration doesn't exist
- `ErrInvalidLabel`: Label is empty
- `ErrInvalidConfig`: Configuration is nil

### Delete - Remove a Configuration

```go
err := client.Store.Delete("prod-config")
if err != nil {
    // Handle error
}
```

**Errors:**
- `ErrConfigNotFound`: Configuration doesn't exist
- `ErrInvalidLabel`: Label is empty

### List - Get All Configuration Labels

```go
labels, err := client.Store.List()
if err != nil {
    // Handle error
}

for _, label := range labels {
    fmt.Printf("Configuration: %s\n", label)
}
```

Labels are returned in alphabetical order.

## Creating Instances from Stored Configurations

The most powerful feature is creating Aura instances directly from stored configurations:

```go
// Store a configuration
config := &aura.CreateInstanceConfigData{
    Name:          "my-database",
    TenantId:      "tenant-123",
    CloudProvider: "gcp",
    Region:        "us-central1",
    Type:          "enterprise-db",
    Version:       "5",
    Memory:        "8GB",
}
client.Store.Create("my-config", config)

// Later, create an instance using the stored configuration
instance, err := client.Instances.CreateFromStore("my-config")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created instance: %s (ID: %s)\n", instance.Data.Name, instance.Data.Id)
fmt.Printf("Connection URL: %s\n", instance.Data.ConnectionUrl)
```

## Error Handling

The store service uses custom error types for better error handling:

```go
config, err := client.Store.Read("nonexistent")
if err != nil {
    var storeErr *aura.StoreError
    if errors.As(err, &storeErr) {
        fmt.Printf("Store operation '%s' failed: %s\n", storeErr.Op, storeErr.Message)
        
        // Check specific error types
        if errors.Is(err, aura.ErrConfigNotFound) {
            fmt.Println("Configuration not found")
        }
    }
}
```

### Common Error Types

- `ErrConfigNotFound`: The requested configuration doesn't exist
- `ErrConfigAlreadyExists`: Attempted to create a configuration with an existing label
- `ErrInvalidLabel`: Label is empty or invalid
- `ErrInvalidConfig`: Configuration data is nil or invalid

## Use Cases

### 1. Environment-Specific Configurations

```go
// Development
devConfig := &aura.CreateInstanceConfigData{
    Name:          "dev-db",
    TenantId:      "my-tenant",
    CloudProvider: "gcp",
    Region:        "us-west1",
    Type:          "professional-db",
    Memory:        "4GB",
}
client.Store.Create("dev", devConfig)

// Production
prodConfig := &aura.CreateInstanceConfigData{
    Name:          "prod-db",
    TenantId:      "my-tenant",
    CloudProvider: "gcp",
    Region:        "us-central1",
    Type:          "enterprise-db",
    Memory:        "16GB",
}
client.Store.Create("prod", prodConfig)

// Create instance for specific environment
instance, _ := client.Instances.CreateFromStore("dev")
```

### 2. Template Configurations

```go
// Store templates for different workload types
templates := map[string]*aura.CreateInstanceConfigData{
    "small-oltp": {
        Name:          "small-transaction-db",
        Type:          "professional-db",
        Memory:        "4GB",
        CloudProvider: "gcp",
        Region:        "us-central1",
    },
    "large-analytics": {
        Name:          "large-analytics-db",
        Type:          "enterprise-db",
        Memory:        "32GB",
        CloudProvider: "aws",
        Region:        "us-east-1",
    },
}

for label, config := range templates {
    client.Store.Create(label, config)
}
```

### 3. Configuration Versioning

```go
// Save different versions of a configuration
v1Config := &aura.CreateInstanceConfigData{
    Name:   "my-db-v1",
    Memory: "8GB",
    // ... other settings
}
client.Store.Create("my-config-v1", v1Config)

v2Config := &aura.CreateInstanceConfigData{
    Name:   "my-db-v2",
    Memory: "16GB", // Upgraded
    // ... other settings
}
client.Store.Create("my-config-v2", v2Config)
```

### 4. Disaster Recovery

```go
// Save production configuration for quick recovery
prodConfig, _ := getProdInstanceConfig()
client.Store.Create("prod-backup", prodConfig)

// Later, if disaster occurs
recovered, _ := client.Instances.CreateFromStore("prod-backup")
fmt.Printf("Recovered instance: %s\n", recovered.Data.Id)
```

## Best Practices

1. **Use descriptive labels**: Choose clear, meaningful labels like `prod-analytics` instead of `config1`

2. **Version your configurations**: Include version numbers in labels for tracking changes

3. **Regular backups**: The SQLite database file should be included in your backup strategy

4. **Validate configurations**: Ensure configurations are valid before storing them

5. **Clean up old configurations**: Regularly delete unused configurations to keep the store manageable

## Database Schema

The store uses a simple SQLite schema:

```sql
CREATE TABLE instance_configs (
    label TEXT PRIMARY KEY,
    config TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

- `label`: Unique identifier for the configuration
- `config`: JSON-encoded configuration data
- `created_at`: Timestamp when configuration was created
- `updated_at`: Timestamp of last update

## Thread Safety

The store service is thread-safe for concurrent reads but uses database-level locking for writes. For high-concurrency scenarios, consider using multiple database files or implementing your own synchronization.

## Limitations

- **Local storage only**: The SQLite database is stored locally
- **No synchronization**: Multiple clients don't automatically sync
- **Single database**: One database per client instance
- **Size limits**: Practical limit around 10,000 configurations

## Migration

If you need to migrate configurations between systems:

```go
// Export on source system
labels, _ := client.Store.List()
configs := make(map[string]*aura.CreateInstanceConfigData)
for _, label := range labels {
    configs[label], _ = client.Store.Read(label)
}

// Import on target system
for label, config := range configs {
    client.Store.Create(label, config)
}
```

## See Also

- [Instances API Documentation](instances.md)
- [Client Configuration](configuration.md)
- [Error Handling](errors.md)
