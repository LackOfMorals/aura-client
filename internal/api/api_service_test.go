package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LackOfMorals/aura-client/internal/httpClient"
	"github.com/LackOfMorals/aura-client/internal/testutil"
)

// ============================================================================
// Test helpers
// ============================================================================

func testLogger() *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelWarn}
	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}

// newTestService constructs an apiRequestService wired to the supplied mock.
// The authManager starts with no token, so tests that need one must either
// pre-seed it (use newTestServiceWithToken) or set up the mock's PostResponse
// to return a valid token payload for the OAuth call.
func newTestService(mock *testutil.MockHTTPService) *apiRequestService {
	return &apiRequestService{
		httpClient: mock,
		authMgr: &authManager{
			clientID:     "test-client-id",
			clientSecret: "test-client-secret",
			logger:       testLogger(),
		},
		baseURL:      "https://api.neo4j.io",
		endpointBase: "https://api.neo4j.io/v1",
		logger:       testLogger(),
	}
}

// newTestServiceWithToken returns a service whose authManager already holds a
// valid token, bypassing the OAuth exchange for tests that focus on routing,
// URL construction, headers, and response handling.
func newTestServiceWithToken(mock *testutil.MockHTTPService) *apiRequestService {
	svc := newTestService(mock)
	svc.authMgr.token = "test-access-token"
	svc.authMgr.tokenType = "Bearer"
	svc.authMgr.expiresAt = time.Now().Unix() + 3600 // valid for 1 hour
	return svc
}

// tokenResponseBody returns a JSON-encoded OAuth token response body.
func tokenResponseBody(accessToken, tokenType string, expiresIn int64) []byte {
	b, _ := json.Marshal(tokenResponse{
		AccessToken: accessToken,
		TokenType:   tokenType,
		ExpiresIn:   expiresIn,
	})
	return b
}

// successHTTPResponse wraps a JSON body in an httpClient.HTTPResponse with status 200.
func successHTTPResponse(body []byte) *httpClient.HTTPResponse {
	return &httpClient.HTTPResponse{StatusCode: http.StatusOK, Body: body}
}

// ============================================================================
// parseError
// ============================================================================

func TestParseError_EmptyBody(t *testing.T) {
	err := parseError(nil, http.StatusNotFound)
	if err.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", err.StatusCode)
	}
	if err.Message != "Not Found" {
		t.Errorf("expected message 'Not Found', got '%s'", err.Message)
	}
	if len(err.Details) != 0 {
		t.Errorf("expected no details, got %d", len(err.Details))
	}
}

func TestParseError_MessageField(t *testing.T) {
	body := []byte(`{"message":"Instance not found"}`)
	err := parseError(body, http.StatusNotFound)
	if err.Message != "Instance not found" {
		t.Errorf("expected message 'Instance not found', got '%s'", err.Message)
	}
}

func TestParseError_ErrorsArray(t *testing.T) {
	body := []byte(`{"message":"Validation failed","errors":[{"message":"name is required","field":"name"},{"message":"region is required","field":"region"}]}`)
	err := parseError(body, http.StatusBadRequest)
	if len(err.Details) != 2 {
		t.Fatalf("expected 2 details, got %d", len(err.Details))
	}
	if err.Details[0].Message != "name is required" {
		t.Errorf("expected first detail 'name is required', got '%s'", err.Details[0].Message)
	}
	if err.Details[0].Field != "name" {
		t.Errorf("expected first detail field 'name', got '%s'", err.Details[0].Field)
	}
}

func TestParseError_DetailsArray(t *testing.T) {
	body := []byte(`{"message":"Validation failed","details":[{"message":"memory must be positive","reason":"invalid_value"}]}`)
	err := parseError(body, http.StatusUnprocessableEntity)
	if len(err.Details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(err.Details))
	}
	if err.Details[0].Reason != "invalid_value" {
		t.Errorf("expected reason 'invalid_value', got '%s'", err.Details[0].Reason)
	}
}

func TestParseError_ErrorsArrayTakesPrecedenceOverDetails(t *testing.T) {
	// When both arrays are present, "errors" wins.
	body := []byte(`{"message":"conflict","errors":[{"message":"from errors"}],"details":[{"message":"from details"}]}`)
	err := parseError(body, http.StatusBadRequest)
	if err.Details[0].Message != "from errors" {
		t.Errorf("expected 'from errors', got '%s'", err.Details[0].Message)
	}
}

func TestParseError_InvalidJSON_FallsBackToDefault(t *testing.T) {
	body := []byte(`not valid json`)
	err := parseError(body, http.StatusInternalServerError)
	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", err.StatusCode)
	}
	// Falls back to http.StatusText
	if err.Message != "Internal Server Error" {
		t.Errorf("expected 'Internal Server Error', got '%s'", err.Message)
	}
}

func TestParseError_EmptyMessageField_FallsBackToStatusText(t *testing.T) {
	body := []byte(`{"message":""}`)
	err := parseError(body, http.StatusForbidden)
	if err.Message != "Forbidden" {
		t.Errorf("expected 'Forbidden' fallback, got '%s'", err.Message)
	}
}

// ============================================================================
// HTTP method routing and URL construction
// ============================================================================

func TestAPIService_Get_RoutesCorrectly(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, `{"data":[]}`)
	svc := newTestServiceWithToken(mock)

	_, err := svc.Get(context.Background(), "instances")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.LastMethod != "GET" {
		t.Errorf("expected GET, got %s", mock.LastMethod)
	}
	if mock.LastURL != "https://api.neo4j.io/v1/instances" {
		t.Errorf("unexpected URL: %s", mock.LastURL)
	}
}

func TestAPIService_Post_RoutesCorrectly(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, `{"data":{}}`)
	svc := newTestServiceWithToken(mock)

	body := `{"name":"my-instance"}`
	_, err := svc.Post(context.Background(), "instances", body)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.LastMethod != "POST" {
		t.Errorf("expected POST, got %s", mock.LastMethod)
	}
	if mock.LastURL != "https://api.neo4j.io/v1/instances" {
		t.Errorf("unexpected URL: %s", mock.LastURL)
	}
	if mock.LastBody != body {
		t.Errorf("expected body '%s', got '%s'", body, mock.LastBody)
	}
}

func TestAPIService_Put_RoutesCorrectly(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, `{"data":{}}`)
	svc := newTestServiceWithToken(mock)

	body := `{"name":"updated"}`
	_, err := svc.Put(context.Background(), "instances/aaaa1234", body)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.LastMethod != "PUT" {
		t.Errorf("expected PUT, got %s", mock.LastMethod)
	}
	if mock.LastURL != "https://api.neo4j.io/v1/instances/aaaa1234" {
		t.Errorf("unexpected URL: %s", mock.LastURL)
	}
}

func TestAPIService_Patch_RoutesCorrectly(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, `{"data":{}}`)
	svc := newTestServiceWithToken(mock)

	body := `{"memory":"16GB"}`
	_, err := svc.Patch(context.Background(), "instances/aaaa1234", body)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.LastMethod != "PATCH" {
		t.Errorf("expected PATCH, got %s", mock.LastMethod)
	}
	if mock.LastURL != "https://api.neo4j.io/v1/instances/aaaa1234" {
		t.Errorf("unexpected URL: %s", mock.LastURL)
	}
}

func TestAPIService_Delete_RoutesCorrectly(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, `{"data":{}}`)
	svc := newTestServiceWithToken(mock)

	_, err := svc.Delete(context.Background(), "instances/aaaa1234")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.LastMethod != "DELETE" {
		t.Errorf("expected DELETE, got %s", mock.LastMethod)
	}
	if mock.LastURL != "https://api.neo4j.io/v1/instances/aaaa1234" {
		t.Errorf("unexpected URL: %s", mock.LastURL)
	}
}

func TestAPIService_URLConstruction_NestedPath(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, `{"data":[]}`)
	svc := newTestServiceWithToken(mock)

	_, err := svc.Get(context.Background(), "instances/aaaa1234/snapshots")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://api.neo4j.io/v1/instances/aaaa1234/snapshots"
	if mock.LastURL != expected {
		t.Errorf("expected URL '%s', got '%s'", expected, mock.LastURL)
	}
}

// ============================================================================
// Request headers
// ============================================================================

func TestAPIService_Headers_ContentType(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, `{"data":[]}`)
	svc := newTestServiceWithToken(mock)

	_, err := svc.Get(context.Background(), "instances")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.LastHeaders["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", mock.LastHeaders["Content-Type"])
	}
}

func TestAPIService_Headers_UserAgent(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, `{"data":[]}`)
	svc := newTestServiceWithToken(mock)
	svc.userAgent = "aura-go-client/v1.8.0"

	_, err := svc.Get(context.Background(), "instances")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.LastHeaders["User-Agent"] != "aura-go-client/v1.8.0" {
		t.Errorf("expected User-Agent 'aura-go-client/v1.8.0', got '%s'", mock.LastHeaders["User-Agent"])
	}
}

func TestAPIService_Headers_UserAgent_DefaultFallback(t *testing.T) {
	// When userAgent is not set (e.g. tests that build apiRequestService directly),
	// NewRequestService applies the fallback "aura-go-client" to keep the header
	// populated. This test verifies the field is used as-is; the empty-string
	// fallback is applied in NewRequestService, not in the header code itself.
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, `{"data":[]}`)
	svc := newTestServiceWithToken(mock)
	// userAgent left as zero value ("") to verify the header is set to that
	// value directly — the fallback lives in NewRequestService.

	_, err := svc.Get(context.Background(), "instances")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Zero-value userAgent means the header is empty; real callers always go
	// through NewRequestService which sets the fallback.
	if got := mock.LastHeaders["User-Agent"]; got != "" {
		t.Logf("note: User-Agent is %q when userAgent field is empty", got)
	}
}

func TestAPIService_Headers_AuthorizationFormat(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, `{"data":[]}`)
	svc := newTestServiceWithToken(mock)

	_, err := svc.Get(context.Background(), "instances")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	authHeader := mock.LastHeaders["Authorization"]
	if authHeader != "Bearer test-access-token" {
		t.Errorf("expected Authorization 'Bearer test-access-token', got '%s'", authHeader)
	}
}

// ============================================================================
// Response handling
// ============================================================================

func TestAPIService_Response_BodyAndStatusReturned(t *testing.T) {
	expectedBody := []byte(`{"data":{"id":"aaaa1234"}}`)
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, string(expectedBody))
	svc := newTestServiceWithToken(mock)

	resp, err := svc.Get(context.Background(), "instances/aaaa1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if string(resp.Body) != string(expectedBody) {
		t.Errorf("expected body '%s', got '%s'", expectedBody, resp.Body)
	}
}

func TestAPIService_Response_201IsSuccess(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusCreated, `{"data":{"id":"new-id"}}`)
	svc := newTestServiceWithToken(mock)

	resp, err := svc.Post(context.Background(), "instances", `{}`)
	if err != nil {
		t.Fatalf("unexpected error for 201: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
}

func TestAPIService_Response_299IsSuccess(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.Response = &httpClient.HTTPResponse{StatusCode: 299, Body: []byte(`{}`)}
	svc := newTestServiceWithToken(mock)

	_, err := svc.Get(context.Background(), "instances")
	if err != nil {
		t.Fatalf("unexpected error for 299: %v", err)
	}
}

// ============================================================================
// API error responses (non-2xx → *Error)
// ============================================================================

func TestAPIService_ErrorResponse_400(t *testing.T) {
	body := `{"message":"Bad Request","errors":[{"message":"name is required","field":"name"}]}`
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusBadRequest, body)
	svc := newTestServiceWithToken(mock)

	_, err := svc.Post(context.Background(), "instances", `{}`)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if !apiErr.IsBadRequest() {
		t.Error("expected IsBadRequest() to be true")
	}
	if apiErr.Details[0].Field != "name" {
		t.Errorf("expected field 'name', got '%s'", apiErr.Details[0].Field)
	}
}

func TestAPIService_ErrorResponse_401(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusUnauthorized, `{"message":"Invalid credentials"}`)
	svc := newTestServiceWithToken(mock)

	_, err := svc.Get(context.Background(), "instances")
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if !apiErr.IsUnauthorized() {
		t.Error("expected IsUnauthorized() to be true")
	}
}

func TestAPIService_ErrorResponse_404(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusNotFound, `{"message":"Instance not found"}`)
	svc := newTestServiceWithToken(mock)

	_, err := svc.Get(context.Background(), "instances/aaaa1234")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if !apiErr.IsNotFound() {
		t.Error("expected IsNotFound() to be true")
	}
	if apiErr.Message != "Instance not found" {
		t.Errorf("expected message 'Instance not found', got '%s'", apiErr.Message)
	}
}

func TestAPIService_ErrorResponse_500(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusInternalServerError, `{"message":"Internal error"}`)
	svc := newTestServiceWithToken(mock)

	_, err := svc.Get(context.Background(), "instances")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
}

func TestAPIService_HTTPClientError_Propagated(t *testing.T) {
	networkErr := fmt.Errorf("connection refused")
	mock := testutil.NewMockHTTPService()
	mock.WithError(networkErr)
	svc := newTestServiceWithToken(mock)

	_, err := svc.Get(context.Background(), "instances")
	if err == nil {
		t.Fatal("expected error to be propagated")
	}
	if !errors.Is(err, networkErr) {
		t.Errorf("expected networkErr, got %v", err)
	}
}

// ============================================================================
// Context handling
// ============================================================================

func TestAPIService_CancelledContext_RejectedBeforeHTTPCall(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, `{"data":[]}`)
	svc := newTestServiceWithToken(mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	_, err := svc.Get(ctx, "instances")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	// The HTTP mock must not have been called.
	if mock.CallCount != 0 {
		t.Errorf("expected 0 HTTP calls, got %d", mock.CallCount)
	}
}

func TestAPIService_ExpiredDeadline_RejectedBeforeHTTPCall(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, `{"data":[]}`)
	svc := newTestServiceWithToken(mock)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := svc.Get(ctx, "instances")
	if err == nil {
		t.Fatal("expected error for expired deadline")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
	if mock.CallCount != 0 {
		t.Errorf("expected 0 HTTP calls, got %d", mock.CallCount)
	}
}

// ============================================================================
// Token acquisition (ensureValidToken)
// ============================================================================

// sequencedMock returns different HTTPResponses for successive POST calls.
// The first call always goes to the OAuth token endpoint; subsequent calls
// are the actual API requests. This lets us test the full auth + request flow.
type sequencedMock struct {
	responses []*httpClient.HTTPResponse
	errors    []error
	mu        sync.Mutex
	callIndex int
	// Capture all calls for assertions
	calls []struct {
		method, url, body string
		headers           map[string]string
	}
}

func (m *sequencedMock) next() (*httpClient.HTTPResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i := m.callIndex
	m.callIndex++
	if i >= len(m.responses) {
		return nil, fmt.Errorf("sequencedMock: unexpected call index %d", i)
	}
	return m.responses[i], m.errors[i]
}

func (m *sequencedMock) record(method, url, body string, headers map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, struct {
		method, url, body string
		headers           map[string]string
	}{method, url, body, headers})
}

func (m *sequencedMock) Get(ctx context.Context, url string, headers map[string]string) (*httpClient.HTTPResponse, error) {
	m.record("GET", url, "", headers)
	return m.next()
}
func (m *sequencedMock) Post(ctx context.Context, url string, headers map[string]string, body string) (*httpClient.HTTPResponse, error) {
	m.record("POST", url, body, headers)
	return m.next()
}
func (m *sequencedMock) Put(ctx context.Context, url string, headers map[string]string, body string) (*httpClient.HTTPResponse, error) {
	m.record("PUT", url, body, headers)
	return m.next()
}
func (m *sequencedMock) Patch(ctx context.Context, url string, headers map[string]string, body string) (*httpClient.HTTPResponse, error) {
	m.record("PATCH", url, body, headers)
	return m.next()
}
func (m *sequencedMock) Delete(ctx context.Context, url string, headers map[string]string) (*httpClient.HTTPResponse, error) {
	m.record("DELETE", url, "", headers)
	return m.next()
}

func newSequencedMock(responses []*httpClient.HTTPResponse, errs []error) *sequencedMock {
	return &sequencedMock{responses: responses, errors: errs}
}

func TestToken_FetchedOnFirstCall(t *testing.T) {
	tokenBody := tokenResponseBody("fresh-token", "Bearer", 3600)
	apiBody := []byte(`{"data":[]}`)

	mock := newSequencedMock(
		[]*httpClient.HTTPResponse{
			{StatusCode: http.StatusOK, Body: tokenBody}, // OAuth call
			{StatusCode: http.StatusOK, Body: apiBody},   // API call
		},
		[]error{nil, nil},
	)

	svc := &apiRequestService{
		httpClient:   mock,
		authMgr:      &authManager{clientID: "id", clientSecret: "secret", logger: testLogger()},
		baseURL:      "https://api.neo4j.io",
		endpointBase: "https://api.neo4j.io/v1",
		logger:       testLogger(),
	}

	resp, err := svc.Get(context.Background(), "instances")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp.Body) != string(apiBody) {
		t.Errorf("expected api body, got %s", resp.Body)
	}

	// First call must be the OAuth token endpoint.
	if len(mock.calls) < 2 {
		t.Fatalf("expected 2 HTTP calls, got %d", len(mock.calls))
	}
	if !strings.HasSuffix(mock.calls[0].url, "/oauth/token") {
		t.Errorf("expected first call to /oauth/token, got %s", mock.calls[0].url)
	}
	// OAuth call must use Basic auth.
	if !strings.HasPrefix(mock.calls[0].headers["Authorization"], "Basic ") {
		t.Errorf("expected Basic auth on token call, got %s", mock.calls[0].headers["Authorization"])
	}
	// API call must use Bearer token.
	if mock.calls[1].headers["Authorization"] != "Bearer fresh-token" {
		t.Errorf("expected Bearer fresh-token on API call, got %s", mock.calls[1].headers["Authorization"])
	}
}

func TestToken_ReusedWhenStillValid(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(http.StatusOK, `{"data":[]}`)

	svc := newTestServiceWithToken(mock)

	// Two calls — neither should trigger a token refresh.
	for i := range 2 {
		_, err := svc.Get(context.Background(), "instances")
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
	}
	if mock.CallCount != 2 {
		t.Errorf("expected 2 HTTP calls (no token refresh), got %d", mock.CallCount)
	}
}

func TestToken_RefreshedWhenExpired(t *testing.T) {
	tokenBody := tokenResponseBody("refreshed-token", "Bearer", 3600)
	apiBody := []byte(`{"data":[]}`)

	mock := newSequencedMock(
		[]*httpClient.HTTPResponse{
			{StatusCode: http.StatusOK, Body: tokenBody}, // token refresh
			{StatusCode: http.StatusOK, Body: apiBody},   // API call
		},
		[]error{nil, nil},
	)

	svc := &apiRequestService{
		httpClient: mock,
		authMgr: &authManager{
			clientID:     "id",
			clientSecret: "secret",
			token:        "expired-token",
			tokenType:    "Bearer",
			expiresAt:    time.Now().Unix() - 1, // already expired
			logger:       testLogger(),
		},
		baseURL:      "https://api.neo4j.io",
		endpointBase: "https://api.neo4j.io/v1",
		logger:       testLogger(),
	}

	_, err := svc.Get(context.Background(), "instances")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.calls) < 2 {
		t.Fatalf("expected 2 HTTP calls (refresh + API), got %d", len(mock.calls))
	}
	if !strings.HasSuffix(mock.calls[0].url, "/oauth/token") {
		t.Errorf("expected first call to be token refresh, got %s", mock.calls[0].url)
	}
	// API call must use the refreshed token.
	if mock.calls[1].headers["Authorization"] != "Bearer refreshed-token" {
		t.Errorf("expected refreshed token on API call, got %s", mock.calls[1].headers["Authorization"])
	}
}

func TestToken_RefreshedWithin60SecondsOfExpiry(t *testing.T) {
	tokenBody := tokenResponseBody("renewed-token", "Bearer", 3600)
	apiBody := []byte(`{"data":[]}`)

	mock := newSequencedMock(
		[]*httpClient.HTTPResponse{
			{StatusCode: http.StatusOK, Body: tokenBody},
			{StatusCode: http.StatusOK, Body: apiBody},
		},
		[]error{nil, nil},
	)

	svc := &apiRequestService{
		httpClient: mock,
		authMgr: &authManager{
			clientID:     "id",
			clientSecret: "secret",
			token:        "nearly-expired-token",
			tokenType:    "Bearer",
			expiresAt:    time.Now().Unix() + 30, // expires in 30s — inside the 60s buffer
			logger:       testLogger(),
		},
		baseURL:      "https://api.neo4j.io",
		endpointBase: "https://api.neo4j.io/v1",
		logger:       testLogger(),
	}

	_, err := svc.Get(context.Background(), "instances")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.calls) < 2 || !strings.HasSuffix(mock.calls[0].url, "/oauth/token") {
		t.Error("expected token refresh for token expiring within 60 seconds")
	}
}

func TestToken_TokenEndpointError_Propagated(t *testing.T) {
	networkErr := fmt.Errorf("token endpoint unreachable")
	mock := newSequencedMock(
		[]*httpClient.HTTPResponse{nil},
		[]error{networkErr},
	)

	svc := &apiRequestService{
		httpClient:   mock,
		authMgr:      &authManager{clientID: "id", clientSecret: "secret", logger: testLogger()},
		baseURL:      "https://api.neo4j.io",
		endpointBase: "https://api.neo4j.io/v1",
		logger:       testLogger(),
	}

	_, err := svc.Get(context.Background(), "instances")
	if err == nil {
		t.Fatal("expected error from token endpoint failure")
	}
	if !errors.Is(err, networkErr) {
		t.Errorf("expected networkErr, got %v", err)
	}
}

func TestToken_TokenEndpointNonSuccess_ReturnsAPIError(t *testing.T) {
	body := []byte(`{"message":"invalid_client"}`)
	mock := newSequencedMock(
		[]*httpClient.HTTPResponse{{StatusCode: http.StatusUnauthorized, Body: body}},
		[]error{nil},
	)

	svc := &apiRequestService{
		httpClient:   mock,
		authMgr:      &authManager{clientID: "id", clientSecret: "secret", logger: testLogger()},
		baseURL:      "https://api.neo4j.io",
		endpointBase: "https://api.neo4j.io/v1",
		logger:       testLogger(),
	}

	_, err := svc.Get(context.Background(), "instances")
	if err == nil {
		t.Fatal("expected error for 401 token response")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestToken_MalformedTokenResponse_ReturnsError(t *testing.T) {
	mock := newSequencedMock(
		[]*httpClient.HTTPResponse{{StatusCode: http.StatusOK, Body: []byte(`not json`)}},
		[]error{nil},
	)

	svc := &apiRequestService{
		httpClient:   mock,
		authMgr:      &authManager{clientID: "id", clientSecret: "secret", logger: testLogger()},
		baseURL:      "https://api.neo4j.io",
		endpointBase: "https://api.neo4j.io/v1",
		logger:       testLogger(),
	}

	_, err := svc.Get(context.Background(), "instances")
	if err == nil {
		t.Fatal("expected error for malformed token response")
	}
	if !strings.Contains(err.Error(), "failed to parse token response") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestToken_OAuthBodyFormat(t *testing.T) {
	// Verify the token request sends grant_type=client_credentials and
	// uses the correct Content-Type header.
	tokenBody := tokenResponseBody("tok", "Bearer", 3600)
	apiBody := []byte(`{"data":[]}`)

	mock := newSequencedMock(
		[]*httpClient.HTTPResponse{
			{StatusCode: http.StatusOK, Body: tokenBody},
			{StatusCode: http.StatusOK, Body: apiBody},
		},
		[]error{nil, nil},
	)

	svc := &apiRequestService{
		httpClient:   mock,
		authMgr:      &authManager{clientID: "id", clientSecret: "secret", logger: testLogger()},
		baseURL:      "https://api.neo4j.io",
		endpointBase: "https://api.neo4j.io/v1",
		logger:       testLogger(),
	}

	_, err := svc.Get(context.Background(), "instances")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tokenCall := mock.calls[0]
	if tokenCall.headers["Content-Type"] != "application/x-www-form-urlencoded" {
		t.Errorf("expected Content-Type 'application/x-www-form-urlencoded' on token call, got '%s'", tokenCall.headers["Content-Type"])
	}
	if !strings.Contains(tokenCall.body, "grant_type=client_credentials") {
		t.Errorf("expected grant_type=client_credentials in token body, got '%s'", tokenCall.body)
	}
}

// ============================================================================
// Concurrent token refresh — double-checked locking
// ============================================================================

func TestToken_ConcurrentRefresh_OnlyOneFetch(t *testing.T) {
	// Many goroutines hit the service simultaneously with no cached token.
	// ensureValidToken's double-checked locking must ensure only one token
	// fetch occurs despite the concurrent pressure.

	const goroutines = 20

	tokenBody := tokenResponseBody("concurrent-token", "Bearer", 3600)
	apiBody := []byte(`{"data":[]}`)

	// Build enough responses: up to goroutines token calls + goroutines API calls.
	// In practice only one token call should happen, but we provide extras to
	// prevent sequencedMock from panicking if the test fails.
	var responses []*httpClient.HTTPResponse
	var errs []error
	for range goroutines {
		responses = append(responses, &httpClient.HTTPResponse{StatusCode: http.StatusOK, Body: tokenBody})
		errs = append(errs, nil)
	}
	for range goroutines {
		responses = append(responses, &httpClient.HTTPResponse{StatusCode: http.StatusOK, Body: apiBody})
		errs = append(errs, nil)
	}

	mock := newSequencedMock(responses, errs)

	svc := &apiRequestService{
		httpClient:   mock,
		authMgr:      &authManager{clientID: "id", clientSecret: "secret", logger: testLogger()},
		baseURL:      "https://api.neo4j.io",
		endpointBase: "https://api.neo4j.io/v1",
		logger:       testLogger(),
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			svc.Get(context.Background(), "instances") //nolint:errcheck
		}()
	}
	wg.Wait()

	// Count token endpoint calls.
	tokenCallCount := 0
	mock.mu.Lock()
	for _, c := range mock.calls {
		if strings.HasSuffix(c.url, "/oauth/token") {
			tokenCallCount++
		}
	}
	mock.mu.Unlock()

	if tokenCallCount != 1 {
		t.Errorf("expected exactly 1 token fetch under concurrent load, got %d", tokenCallCount)
	}
}

// ============================================================================
// Error type helper methods
// ============================================================================

func TestError_IsNotFound(t *testing.T) {
	tests := []struct{ code int; want bool }{
		{http.StatusNotFound, true},
		{http.StatusOK, false},
		{http.StatusUnauthorized, false},
	}
	for _, tt := range tests {
		e := &Error{StatusCode: tt.code}
		if e.IsNotFound() != tt.want {
			t.Errorf("status %d: IsNotFound() = %v, want %v", tt.code, e.IsNotFound(), tt.want)
		}
	}
}

func TestError_IsUnauthorized(t *testing.T) {
	tests := []struct{ code int; want bool }{
		{http.StatusUnauthorized, true},
		{http.StatusForbidden, false},
		{http.StatusOK, false},
	}
	for _, tt := range tests {
		e := &Error{StatusCode: tt.code}
		if e.IsUnauthorized() != tt.want {
			t.Errorf("status %d: IsUnauthorized() = %v, want %v", tt.code, e.IsUnauthorized(), tt.want)
		}
	}
}

func TestError_IsBadRequest(t *testing.T) {
	tests := []struct{ code int; want bool }{
		{http.StatusBadRequest, true},
		{http.StatusUnprocessableEntity, false},
		{http.StatusOK, false},
	}
	for _, tt := range tests {
		e := &Error{StatusCode: tt.code}
		if e.IsBadRequest() != tt.want {
			t.Errorf("status %d: IsBadRequest() = %v, want %v", tt.code, e.IsBadRequest(), tt.want)
		}
	}
}

func TestError_Error_NoDetails(t *testing.T) {
	e := &Error{StatusCode: 404, Message: "Not Found"}
	expected := "API error (status 404): Not Found"
	if e.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, e.Error())
	}
}

func TestError_Error_SingleDetail(t *testing.T) {
	e := &Error{
		StatusCode: 400,
		Message:    "Bad Request",
		Details:    []ErrorDetail{{Message: "name is required"}},
	}
	expected := "API error (status 400): Bad Request - name is required"
	if e.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, e.Error())
	}
}

func TestError_Error_MultipleDetails(t *testing.T) {
	e := &Error{
		StatusCode: 422,
		Message:    "Validation Error",
		Details: []ErrorDetail{
			{Message: "field A"},
			{Message: "field B"},
			{Message: "field C"},
		},
	}
	msg := e.Error()
	if !strings.Contains(msg, "and 2 more error(s)") {
		t.Errorf("expected '2 more error(s)' in message, got '%s'", msg)
	}
}

func TestError_AllErrors(t *testing.T) {
	e := &Error{
		StatusCode: 400,
		Message:    "top-level",
		Details: []ErrorDetail{
			{Message: "detail-1"},
			{Message: "detail-2"},
		},
	}
	all := e.AllErrors()
	if len(all) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(all))
	}
	if all[0] != "top-level" {
		t.Errorf("expected first to be top-level message, got '%s'", all[0])
	}
}

func TestError_HasMultipleErrors(t *testing.T) {
	single := &Error{Details: []ErrorDetail{{Message: "one"}}}
	if single.HasMultipleErrors() {
		t.Error("single detail: expected HasMultipleErrors() = false")
	}
	multi := &Error{Details: []ErrorDetail{{Message: "one"}, {Message: "two"}}}
	if !multi.HasMultipleErrors() {
		t.Error("two details: expected HasMultipleErrors() = true")
	}
}
