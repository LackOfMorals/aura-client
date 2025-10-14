package auraAPIClient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	httpClient "github.com/LackOfMorals/aura-api-client/auraAPIClient/internal/httpClient"
)

// mockHTTPService implements httpClient.HTTPService for testing makeAuthenticatedRequest
type mockHTTPService struct {
	// captured inputs
	gotCtx      context.Context
	gotEndpoint string
	gotMethod   string
	gotHeader   map[string]string
	gotBody     string

	// configured outputs
	payload any
	retErr  error
}

func (m *mockHTTPService) MakeRequest(ctx context.Context, endpoint string, method string, header map[string]string, body string) (*httpClient.HTTPResponse, error) {
	m.gotCtx = ctx
	m.gotEndpoint = endpoint
	m.gotMethod = method
	m.gotHeader = header
	m.gotBody = body

	if m.retErr != nil {
		return nil, m.retErr
	}

	var b []byte
	if m.payload != nil {
		b, _ = json.Marshal(m.payload)
	}
	return &httpClient.HTTPResponse{ResponsePayload: &b, RequestResponse: &http.Response{StatusCode: http.StatusOK}}, nil
}

func TestNewAuraAPIActionsService_Constructs(t *testing.T) {
	svc := NewAuraAPIActionsService("id", "sec")

	if svc == nil {
		t.Fatal("expected non-nil service")
	}

	if svc.auraAPIBaseURL != BaseURL {
		t.Errorf("base url: want %s got %s", BaseURL, svc.auraAPIBaseURL)
	}
	if svc.auraAPIVersion != ApiVersion {
		t.Errorf("version: want %s got %s", ApiVersion, svc.auraAPIVersion)
	}
	if svc.auraAPITimeout != ApiTimeout || svc.timeout != ApiTimeout {
		t.Errorf("timeouts: want %v got %v/%v", ApiTimeout, svc.auraAPITimeout, svc.timeout)
	}

	// sub-services are initialized
	if svc.Auth == nil || svc.Tenants == nil || svc.Instances == nil || svc.Snapshots == nil {
		t.Errorf("expected all sub-services to be initialized")
	}
	// http client is set
	if svc.http == nil {
		t.Errorf("expected http client to be initialized")
	}
}

func TestMakeAuthenticatedRequest_Success(t *testing.T) {
	// Arrange
	svc := NewAuraAPIActionsService("client", "secret")
	mock := &mockHTTPService{payload: struct {
		Message string `json:"message"`
		OK      bool   `json:"ok"`
	}{Message: "hi", OK: true}}
	svc.http = mock

	token := &AuthAPIToken{Type: "Bearer", Token: "tok"}
	ctx := context.Background()
	endpoint := svc.auraAPIVersion + "/instances"
	body := "{}"

	// Act
	resp, err := makeAuthenticatedRequest[struct {
		Message string `json:"message"`
		OK      bool   `json:"ok"`
	}](ctx, svc, token, endpoint, http.MethodPost, "application/json", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert: response unmarshalled
	if resp == nil || resp.Message != "hi" || !resp.OK {
		t.Fatalf("unexpected response: %+v", resp)
	}

	// Assert: mock captured expected inputs
	if mock.gotEndpoint != endpoint {
		t.Errorf("endpoint: want %s got %s", endpoint, mock.gotEndpoint)
	}
	if mock.gotMethod != http.MethodPost {
		t.Errorf("method: want %s got %s", http.MethodPost, mock.gotMethod)
	}
	if ct := mock.gotHeader["Content-Type"]; ct != "application/json" {
		t.Errorf("content-type: want application/json got %s", ct)
	}
	if ua := mock.gotHeader["User-Agent"]; ua != userAgent {
		t.Errorf("user-agent: want %s got %s", userAgent, ua)
	}
	if auth := mock.gotHeader["Authorization"]; auth != "Bearer tok" {
		t.Errorf("authorization: want %s got %s", "Bearer tok", auth)
	}
	if !reflect.DeepEqual(mock.gotBody, body) {
		t.Errorf("body: want %q got %q", body, mock.gotBody)
	}
}

func TestMakeAuthenticatedRequest_ErrorPropagates(t *testing.T) {
	svc := NewAuraAPIActionsService("client", "secret")
	mock := &mockHTTPService{retErr: errors.New("boom")}
	svc.http = mock

	token := &AuthAPIToken{Type: "Bearer", Token: "tok"}
	_, err := makeAuthenticatedRequest[struct{}](context.Background(), svc, token, svc.auraAPIVersion+"/x", http.MethodGet, "application/json", "")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestMakeAuthenticatedRequest_ContextCanceled(t *testing.T) {
	svc := NewAuraAPIActionsService("client", "secret")
	// mock that would fail if called
	mock := &mockHTTPService{retErr: nil}
	svc.http = mock

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel

	token := &AuthAPIToken{Type: "Bearer", Token: "tok"}
	_, err := makeAuthenticatedRequest[struct{}](ctx, svc, token, svc.auraAPIVersion+"/x", http.MethodGet, "application/json", "")
	if err == nil {
		t.Fatalf("expected context error, got nil")
	}
}

func TestCheckDate(t *testing.T) {
	cases := []struct {
		in string
		ok bool
	}{
		{"2024-01-31", true},
		{"2024-13-01", false},
		{"", false},
		{"20240131", false},
	}
	for _, c := range cases {
		err := checkDate(c.in)
		if (err == nil) != c.ok {
			t.Errorf("checkDate(%q) ok=%v got err=%v", c.in, c.ok, err)
		}
	}
}
