package httpClient

import (
	"context"
)

// MockHTTPService is a mock implementation of HTTPService for testing
type MockHTTPService struct {
	// Response to return for all requests
	Response *HTTPResponse
	// Error to return for all requests
	Error error

	// Method-specific responses (takes precedence over Response)
	GetResponse    *HTTPResponse
	GetError       error
	PostResponse   *HTTPResponse
	PostError      error
	PutResponse    *HTTPResponse
	PutError       error
	PatchResponse  *HTTPResponse
	PatchError     error
	DeleteResponse *HTTPResponse
	DeleteError    error

	// Capture the last request for assertions
	LastMethod   string
	LastURL      string
	LastHeaders  map[string]string
	LastBody     string
	CallCount    int
	CallHistory  []MockCall
}

// MockCall represents a single call to the mock service
type MockCall struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

// NewMockHTTPService creates a new mock HTTP service
func NewMockHTTPService() *MockHTTPService {
	return &MockHTTPService{
		CallHistory: make([]MockCall, 0),
	}
}

// Get implements HTTPService.Get
func (m *MockHTTPService) Get(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error) {
	m.recordCall("GET", url, headers, "")
	if m.GetResponse != nil || m.GetError != nil {
		return m.GetResponse, m.GetError
	}
	return m.Response, m.Error
}

// Post implements HTTPService.Post
func (m *MockHTTPService) Post(ctx context.Context, url string, headers map[string]string, body string) (*HTTPResponse, error) {
	m.recordCall("POST", url, headers, body)
	if m.PostResponse != nil || m.PostError != nil {
		return m.PostResponse, m.PostError
	}
	return m.Response, m.Error
}

// Put implements HTTPService.Put
func (m *MockHTTPService) Put(ctx context.Context, url string, headers map[string]string, body string) (*HTTPResponse, error) {
	m.recordCall("PUT", url, headers, body)
	if m.PutResponse != nil || m.PutError != nil {
		return m.PutResponse, m.PutError
	}
	return m.Response, m.Error
}

// Patch implements HTTPService.Patch
func (m *MockHTTPService) Patch(ctx context.Context, url string, headers map[string]string, body string) (*HTTPResponse, error) {
	m.recordCall("PATCH", url, headers, body)
	if m.PatchResponse != nil || m.PatchError != nil {
		return m.PatchResponse, m.PatchError
	}
	return m.Response, m.Error
}

// Delete implements HTTPService.Delete
func (m *MockHTTPService) Delete(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error) {
	m.recordCall("DELETE", url, headers, "")
	if m.DeleteResponse != nil || m.DeleteError != nil {
		return m.DeleteResponse, m.DeleteError
	}
	return m.Response, m.Error
}

// recordCall records the call details for later assertions
func (m *MockHTTPService) recordCall(method, url string, headers map[string]string, body string) {
	m.LastMethod = method
	m.LastURL = url
	m.LastHeaders = headers
	m.LastBody = body
	m.CallCount++
	m.CallHistory = append(m.CallHistory, MockCall{
		Method:  method,
		URL:     url,
		Headers: headers,
		Body:    body,
	})
}

// Reset clears all recorded calls and responses
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
	m.CallHistory = make([]MockCall, 0)
}

// WithResponse sets the default response for all methods
func (m *MockHTTPService) WithResponse(statusCode int, body string) *MockHTTPService {
	m.Response = &HTTPResponse{
		StatusCode: statusCode,
		Body:       []byte(body),
	}
	return m
}

// WithError sets the default error for all methods
func (m *MockHTTPService) WithError(err error) *MockHTTPService {
	m.Error = err
	return m
}

// WithGetResponse sets a specific response for GET requests
func (m *MockHTTPService) WithGetResponse(statusCode int, body string) *MockHTTPService {
	m.GetResponse = &HTTPResponse{
		StatusCode: statusCode,
		Body:       []byte(body),
	}
	return m
}

// WithPostResponse sets a specific response for POST requests
func (m *MockHTTPService) WithPostResponse(statusCode int, body string) *MockHTTPService {
	m.PostResponse = &HTTPResponse{
		StatusCode: statusCode,
		Body:       []byte(body),
	}
	return m
}
