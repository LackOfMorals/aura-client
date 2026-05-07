# Architecture Overview

## Layered Service Architecture

The Aura API client follows a clean, layered architecture that separates concerns and promotes code reuse:

```
┌─────────────────────────────────────────────────────┐
│              Client (AuraAPIClient)                  │
│  - Tenants, Instances, Snapshots, CMEK, GDS,       │
│    Prometheus Services                              │
└───────────────────┬─────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│           API Service (RequestService)            │
│  - Handles OAuth authentication                      │
│  - Token management and refresh                      │
│  - Request/response handling                         │
│  - Supports both relative paths and full URLs       │
└───────────────────┬─────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│           HTTP Service (HTTPService)                 │
│  - Low-level HTTP operations                         │
│  - Retry logic (network errors only) and pooling    │
│  - Honours an injected *http.Client when supplied   │
│  - Timeout management via context deadline           │
└─────────────────────────────────────────────────────┘
```

## Key Design Principles

### 1. Automatic URL Handling

The API service detects whether an endpoint is a full URL or a relative path
and only joins the base URL + version when the caller passes a relative path:

```go
// Relative path — base URL and version are prepended.
resp := apiSvc.Get(ctx, "instances")
// → https://api.neo4j.io/v1/instances

// Full URL — used as-is, retaining authentication.
resp := apiSvc.Get(ctx, "https://prometheus.example.com/api/v1/query")
// → https://prometheus.example.com/api/v1/query
```

The HTTP service below it never reasons about base URLs — it executes whatever
URL the API layer hands it.

### 2. Unified Authentication

All services use the same API service layer for authentication:

```go
// Aura API endpoint (relative path)
resp := apiSvc.Get(ctx, "instances")
// → Authenticates → Prepends API version → Calls HTTP service

// Prometheus endpoint (full URL)
resp := apiSvc.Get(ctx, "https://c9f0d13a.metrics.neo4j.io/prometheus/api/v1/query?...")
// → Authenticates → Passes to HTTP service directly
```

Both benefit from:
- Automatic OAuth token management
- Token refresh when expired
- Consistent error handling
- Structured logging

### 3. Service Isolation

Each service is responsible for its domain:

**HTTP Service**
- URL construction (base URL + path OR full URL)
- HTTP communication
- Retry logic
- Connection pooling

**API Service**
- Authentication and authorization
- Token lifecycle management
- Request preparation
- Response validation

**Domain Services** (Instances, Prometheus, etc.)
- Business logic
- Request/response mapping
- Domain-specific validation

## Example: Prometheus Service Flow

When you query Prometheus metrics:

```go
// 1. Caller invokes the Prometheus service with its own context.
ctx := context.Background()
health, err := client.Prometheus.GetInstanceHealth(ctx, instanceID, prometheusURL)

// 2. The service applies the configured per-call timeout to ctx, validates
//    inputs, and forwards the (possibly absolute) URL to the API service.
resp, err := p.api.Get(ctx, prometheusURL)

// 3. The API service obtains/refreshes the OAuth token and attaches the
//    Authorization header along with caller-supplied default headers.
headers := map[string]string{
    "Authorization": "Bearer " + token,
    "User-Agent":    s.userAgent,
    "Content-Type":  "application/json",
}

// 4. The HTTP service runs the request through the configured *http.Client
//    (default or user-provided via WithHTTPClient) wrapped in the retry policy.
resp, err := s.httpClient.Get(ctx, fullURL, headers)

// 5. URL detection happens in the API service: full URLs are passed through;
//    relative endpoints are joined with the base URL plus version.
if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
    fullURL = endpoint
} else {
    fullURL = endpointBase + "/" + endpoint
}

// 6. Response flows back through the layers.
```

## Benefits of This Architecture

1. **Code Reuse**: No special methods needed for different URL types
2. **Consistency**: All services use the same interfaces
3. **Maintainability**: Changes to authentication or HTTP handling affect all services
4. **Testability**: Each layer can be mocked independently
5. **Flexibility**: Easy to add new services (internal or external)
6. **Separation of Concerns**: Each layer has a single, clear responsibility

## Adding New Services

To add a new service (e.g., for a third-party API):

```go
type MyService struct {
    api     api.RequestService  // shared API service, handles auth + retries
    timeout time.Duration       // per-call timeout ceiling (from client config)
    logger  *slog.Logger
}

// Context flows in per call. The service applies its configured timeout as
// a ceiling — if the caller's ctx already has a shorter deadline, that wins.
func (s *MyService) CallExternalAPI(ctx context.Context) error {
    if err := ctx.Err(); err != nil {
        return err
    }
    ctx, cancel := context.WithTimeout(ctx, s.timeout)
    defer cancel()

    // Pass the full URL — authentication and User-Agent are added by the
    // api layer. Relative endpoints are joined to the configured base URL.
    resp, err := s.api.Get(ctx, "https://external-api.com/endpoint")
    // ...
}
```

No special setup needed — authentication, default headers, and URL handling
all work automatically. Context is passed per call rather than stored on the
struct, which preserves the standard Go cancellation/tracing model.

## Backward Compatibility

This architecture maintains full backward compatibility:

- Relative paths work exactly as before (base URL prepended)
- Existing services (Instances, Tenants, etc.) are unaffected
- No breaking changes to public APIs
- Internal refactoring only

## Performance Considerations

- **Token Caching**: OAuth tokens are cached and only refreshed when needed
- **Connection Pooling**: HTTP client reuses connections
- **Concurrent Requests**: Thread-safe token management
- **Timeout Control**: Configurable timeouts at each layer
- **Retry Logic**: Automatic retries for transient failures

## Testing Strategy

Each layer has its own test coverage:

- **HTTP Service**: Tests URL handling, retries, timeouts
- **API Service**: Tests authentication, token refresh
- **Domain Services**: Tests business logic, data mapping

Integration tests verify the full stack works together.
