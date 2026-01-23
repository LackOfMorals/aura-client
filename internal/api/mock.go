package api

import (
	"context"
)

// MockAPIRequestService is a mock implementation of APIRequestService for testing
type MockAPIRequestService struct {
	// Default response for all requests
	Response *APIResponse
	Error    error

	// Method-specific responses
	GetResponse    *APIResponse
	GetError       error
	PostResponse   *APIResponse
	PostError      error
	PutResponse    *APIResponse
	PutError       error
	PatchResponse  *APIResponse
	PatchError     error
	DeleteResponse *APIResponse
	DeleteError    error

	// Capture the last request for assertions
	LastMethod   string
	LastEndpoint string
	LastBody     string
	CallCount    int
	CallHistory  []MockAPICall
}

// MockAPICall represents a single call to the mock service
type MockAPICall struct {
	Method   string
	Endpoint string
	Body     string
}

// NewMockAPIRequestService creates a new mock API request service
func NewMockAPIRequestService() *MockAPIRequestService {
	return &MockAPIRequestService{
		CallHistory: make([]MockAPICall, 0),
	}
}

// Get implements APIRequestService.Get
func (m *MockAPIRequestService) Get(ctx context.Context, endpoint string) (*APIResponse, error) {
	m.recordCall("GET", endpoint, "")
	if m.GetResponse != nil || m.GetError != nil {
		return m.GetResponse, m.GetError
	}
	return m.Response, m.Error
}

// Post implements APIRequestService.Post
func (m *MockAPIRequestService) Post(ctx context.Context, endpoint string, body string) (*APIResponse, error) {
	m.recordCall("POST", endpoint, body)
	if m.PostResponse != nil || m.PostError != nil {
		return m.PostResponse, m.PostError
	}
	return m.Response, m.Error
}

// Put implements APIRequestService.Put
func (m *MockAPIRequestService) Put(ctx context.Context, endpoint string, body string) (*APIResponse, error) {
	m.recordCall("PUT", endpoint, body)
	if m.PutResponse != nil || m.PutError != nil {
		return m.PutResponse, m.PutError
	}
	return m.Response, m.Error
}

// Patch implements APIRequestService.Patch
func (m *MockAPIRequestService) Patch(ctx context.Context, endpoint string, body string) (*APIResponse, error) {
	m.recordCall("PATCH", endpoint, body)
	if m.PatchResponse != nil || m.PatchError != nil {
		return m.PatchResponse, m.PatchError
	}
	return m.Response, m.Error
}

// Delete implements APIRequestService.Delete
func (m *MockAPIRequestService) Delete(ctx context.Context, endpoint string) (*APIResponse, error) {
	m.recordCall("DELETE", endpoint, "")
	if m.DeleteResponse != nil || m.DeleteError != nil {
		return m.DeleteResponse, m.DeleteError
	}
	return m.Response, m.Error
}

// recordCall records the call details for later assertions
func (m *MockAPIRequestService) recordCall(method, endpoint, body string) {
	m.LastMethod = method
	m.LastEndpoint = endpoint
	m.LastBody = body
	m.CallCount++
	m.CallHistory = append(m.CallHistory, MockAPICall{
		Method:   method,
		Endpoint: endpoint,
		Body:     body,
	})
}

// Reset clears all recorded calls and responses
func (m *MockAPIRequestService) Reset() {
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
	m.LastEndpoint = ""
	m.LastBody = ""
	m.CallCount = 0
	m.CallHistory = make([]MockAPICall, 0)
}

// WithResponse sets the default response for all methods
func (m *MockAPIRequestService) WithResponse(statusCode int, body string) *MockAPIRequestService {
	m.Response = &APIResponse{
		StatusCode: statusCode,
		Body:       []byte(body),
	}
	return m
}

// WithError sets the default error for all methods
func (m *MockAPIRequestService) WithError(err error) *MockAPIRequestService {
	m.Error = err
	return m
}

// WithGetResponse sets a specific response for GET requests
func (m *MockAPIRequestService) WithGetResponse(statusCode int, body string) *MockAPIRequestService {
	m.GetResponse = &APIResponse{
		StatusCode: statusCode,
		Body:       []byte(body),
	}
	return m
}

// WithPostResponse sets a specific response for POST requests
func (m *MockAPIRequestService) WithPostResponse(statusCode int, body string) *MockAPIRequestService {
	m.PostResponse = &APIResponse{
		StatusCode: statusCode,
		Body:       []byte(body),
	}
	return m
}

// WithAPIError sets an API error response
func (m *MockAPIRequestService) WithAPIError(statusCode int, message string) *MockAPIRequestService {
	m.Error = &APIError{
		StatusCode: statusCode,
		Message:    message,
	}
	return m
}
