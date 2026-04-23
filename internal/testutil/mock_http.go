// Package testutil provides mock implementations for use in tests across the
// aura-client module. Nothing in this package should be imported by production
// code — it exists solely to share test helpers between internal packages
// without putting mock types inside the packages they mock.
package testutil

import (
	"context"

	"github.com/LackOfMorals/aura-client/internal/httpClient"
)

// MockHTTPService is a test double for httpClient.HTTPService.
// It records every call made to it and can be pre-loaded with per-method
// responses so tests can exercise specific code paths without a network.
type MockHTTPService struct {
	// Default response returned when no method-specific response is set.
	Response *httpClient.HTTPResponse
	Error    error

	// Method-specific responses (take precedence over Response/Error).
	GetResponse    *httpClient.HTTPResponse
	GetError       error
	PostResponse   *httpClient.HTTPResponse
	PostError      error
	PutResponse    *httpClient.HTTPResponse
	PutError       error
	PatchResponse  *httpClient.HTTPResponse
	PatchError     error
	DeleteResponse *httpClient.HTTPResponse
	DeleteError    error

	// Call recording for assertions.
	LastMethod  string
	LastURL     string
	LastHeaders map[string]string
	LastBody    string
	CallCount   int
	CallHistory []MockHTTPCall
}

// MockHTTPCall records the details of a single call to MockHTTPService.
type MockHTTPCall struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

// NewMockHTTPService returns an empty MockHTTPService ready for use.
func NewMockHTTPService() *MockHTTPService {
	return &MockHTTPService{
		CallHistory: make([]MockHTTPCall, 0),
	}
}

// Get implements httpClient.HTTPService.
func (m *MockHTTPService) Get(ctx context.Context, url string, headers map[string]string) (*httpClient.HTTPResponse, error) {
	m.record("GET", url, headers, "")
	if m.GetResponse != nil || m.GetError != nil {
		return m.GetResponse, m.GetError
	}
	return m.Response, m.Error
}

// Post implements httpClient.HTTPService.
func (m *MockHTTPService) Post(ctx context.Context, url string, headers map[string]string, body string) (*httpClient.HTTPResponse, error) {
	m.record("POST", url, headers, body)
	if m.PostResponse != nil || m.PostError != nil {
		return m.PostResponse, m.PostError
	}
	return m.Response, m.Error
}

// Put implements httpClient.HTTPService.
func (m *MockHTTPService) Put(ctx context.Context, url string, headers map[string]string, body string) (*httpClient.HTTPResponse, error) {
	m.record("PUT", url, headers, body)
	if m.PutResponse != nil || m.PutError != nil {
		return m.PutResponse, m.PutError
	}
	return m.Response, m.Error
}

// Patch implements httpClient.HTTPService.
func (m *MockHTTPService) Patch(ctx context.Context, url string, headers map[string]string, body string) (*httpClient.HTTPResponse, error) {
	m.record("PATCH", url, headers, body)
	if m.PatchResponse != nil || m.PatchError != nil {
		return m.PatchResponse, m.PatchError
	}
	return m.Response, m.Error
}

// Delete implements httpClient.HTTPService.
func (m *MockHTTPService) Delete(ctx context.Context, url string, headers map[string]string) (*httpClient.HTTPResponse, error) {
	m.record("DELETE", url, headers, "")
	if m.DeleteResponse != nil || m.DeleteError != nil {
		return m.DeleteResponse, m.DeleteError
	}
	return m.Response, m.Error
}

func (m *MockHTTPService) record(method, url string, headers map[string]string, body string) {
	m.LastMethod = method
	m.LastURL = url
	m.LastHeaders = headers
	m.LastBody = body
	m.CallCount++
	m.CallHistory = append(m.CallHistory, MockHTTPCall{
		Method:  method,
		URL:     url,
		Headers: headers,
		Body:    body,
	})
}

// Reset clears all recorded calls and configured responses.
func (m *MockHTTPService) Reset() {
	m.Response = nil
	m.Error = nil
	m.GetResponse = nil
	m.GetError = nil
	m.PostResponse = nil
	m.PostError = nil
	m.PutResponse = nil
	m.PutError = nil
	m.PatchResponse = nil
	m.PatchError = nil
	m.DeleteResponse = nil
	m.DeleteError = nil
	m.LastMethod = ""
	m.LastURL = ""
	m.LastHeaders = nil
	m.LastBody = ""
	m.CallCount = 0
	m.CallHistory = make([]MockHTTPCall, 0)
}

// WithResponse configures a default response for all HTTP methods.
func (m *MockHTTPService) WithResponse(statusCode int, body string) *MockHTTPService {
	m.Response = &httpClient.HTTPResponse{
		StatusCode: statusCode,
		Body:       []byte(body),
	}
	return m
}

// WithError configures a default error for all HTTP methods.
func (m *MockHTTPService) WithError(err error) *MockHTTPService {
	m.Error = err
	return m
}

// WithGetResponse configures a response returned only for GET requests.
func (m *MockHTTPService) WithGetResponse(statusCode int, body string) *MockHTTPService {
	m.GetResponse = &httpClient.HTTPResponse{
		StatusCode: statusCode,
		Body:       []byte(body),
	}
	return m
}

// WithPostResponse configures a response returned only for POST requests.
func (m *MockHTTPService) WithPostResponse(statusCode int, body string) *MockHTTPService {
	m.PostResponse = &httpClient.HTTPResponse{
		StatusCode: statusCode,
		Body:       []byte(body),
	}
	return m
}
