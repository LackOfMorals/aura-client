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

**Test Coverage - instances**
1. Success Cases
TestListInstances_Success - Tests successful listing of instances
TestGetInstance_Success - Tests retrieving a specific instance
TestCreateInstance_Success - Tests creating a new instance
TestDeleteInstance_Success - Tests deleting an instance

2. Edge Cases

TestListInstances_EmptyList - Tests handling of empty instance lists

3. Context Support

TestListInstances_ContextCancellation - Tests that cancelled contexts are handled correctly
TestListInstances_ContextTimeout - Tests timeout behavior

4. Header Verification

TestMakeAuthenticatedRequest_UserAgentConstant - Verifies the user agent constant is used
TestMakeAuthenticatedRequest_AuthorizationHeader - Verifies correct authorization header format
TestMakeAuthenticatedRequest_ContentTypeHeader - Verifies JSON content type is set

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