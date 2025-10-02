# Aura API Client

This a hobby project to create a package that can be used to access the Aura API in Go.  

Your mileage will vary 


## Usage





## Tests

### httpCLient

**Test coverage**
1. Constructor Tests
TestNewHTTPRequestService - Verifies service creation and field initialization

2. Basic Request Tests
TestMakeRequest_Success - Tests successful GET request
TestMakeRequest_WithHeaders - Validates custom headers are sent
TestMakeRequest_WithBody - Tests POST request with body payload

3. Error Handling Tests
TestMakeRequest_4xxError - Tests 404 and client errors
TestMakeRequest_5xxError - Tests 500 and server errors
TestMakeRequest_InvalidURL - Tests handling of invalid domains

4. HTTP Method Tests
TestMakeRequest_AllHTTPMethods - Tests GET, POST, PUT, PATCH, DELETE

5. Response Handling Tests
TestMakeRequest_JSONResponse - Tests JSON deserialization
TestCheckResponse_Success - Tests status codes 200-299
TestCheckResponse_Errors - Tests various error status codes

Running the Tests

```bash
# Run all tests
go test -v

# Run with coverage
go test -cover

# Generate coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### auraClient

**Test Coverage**
1. Service Constructor
TestNewAuraAPIActionsService - Validates proper initialization of all service fields

2. Authentication Tests
TestGetAuthToken_Success - Tests successful OAuth token retrieval
TestGetAuthToken_Unauthorized - Tests handling of invalid credentials

3. Tenant Management Tests
TestListTenants_Success - Tests listing multiple tenants
TestListTenants_EmptyList - Tests empty tenant list response
TestGetTenant_Success - Tests retrieving tenant details with configurations
TestGetTenant_NotFound - Tests 404 error handling

4. Instance Management Tests
TestListInstances_Success - Tests listing multiple instances
TestListInstances_EmptyList - Tests empty instance list
TestListInstances_Unauthorized - Tests expired/invalid token handling

5. Model Structure Tests
TestAuthAPIToken_Structure - Validates token model
TestTenantInstanceConfiguration_Structure - Validates configuration model
TestListInstanceData_Structure - Validates instance data model

```bash
# Run all tests
go test -v

# Run specific test
go test -v -run TestGetAuthToken_Success

# With coverage
go test -cover

# Coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```