package api

import (
	"context"
)

// MockRequestService is a mock implementation of RequestService for testing
type MockRequestService struct {
	// Default response for all requests
	Response *Response
	Error    error

	// Method-specific responses
	GetResponse    *Response
	GetError       error
	PostResponse   *Response
	PostError      error
	PutResponse    *Response
	PutError       error
	PatchResponse  *Response
	PatchError     error
	DeleteResponse *Response
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

// NewMockRequestService creates a new mock API request service
func NewMockRequestService() *MockRequestService {
	return &MockRequestService{
		CallHistory: make([]MockAPICall, 0),
	}
}

// Get implements RequestService.Get
func (m *MockRequestService) Get(ctx context.Context, endpoint string) (*Response, error) {
	m.recordCall("GET", endpoint, "")
	if m.GetResponse != nil || m.GetError != nil {
		return m.GetResponse, m.GetError
	}
	return m.Response, m.Error
}

// Post implements RequestService.Post
func (m *MockRequestService) Post(ctx context.Context, endpoint string, body string) (*Response, error) {
	m.recordCall("POST", endpoint, body)
	if m.PostResponse != nil || m.PostError != nil {
		return m.PostResponse, m.PostError
	}
	return m.Response, m.Error
}

// Put implements RequestService.Put
func (m *MockRequestService) Put(ctx context.Context, endpoint string, body string) (*Response, error) {
	m.recordCall("PUT", endpoint, body)
	if m.PutResponse != nil || m.PutError != nil {
		return m.PutResponse, m.PutError
	}
	return m.Response, m.Error
}

// Patch implements RequestService.Patch
func (m *MockRequestService) Patch(ctx context.Context, endpoint string, body string) (*Response, error) {
	m.recordCall("PATCH", endpoint, body)
	if m.PatchResponse != nil || m.PatchError != nil {
		return m.PatchResponse, m.PatchError
	}
	return m.Response, m.Error
}

// Delete implements RequestService.Delete
func (m *MockRequestService) Delete(ctx context.Context, endpoint string) (*Response, error) {
	m.recordCall("DELETE", endpoint, "")
	if m.DeleteResponse != nil || m.DeleteError != nil {
		return m.DeleteResponse, m.DeleteError
	}
	return m.Response, m.Error
}

// recordCall records the call details for later assertions
func (m *MockRequestService) recordCall(method, endpoint, body string) {
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
func (m *MockRequestService) Reset() {
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
func (m *MockRequestService) WithResponse(statusCode int, body string) *MockRequestService {
	m.Response = &Response{
		StatusCode: statusCode,
		Body:       []byte(body),
	}
	return m
}

// WithError sets the default error for all methods
func (m *MockRequestService) WithError(err error) *MockRequestService {
	m.Error = err
	return m
}

// WithGetResponse sets a specific response for GET requests
func (m *MockRequestService) WithGetResponse(statusCode int, body string) *MockRequestService {
	m.GetResponse = &Response{
		StatusCode: statusCode,
		Body:       []byte(body),
	}
	return m
}

// WithPostResponse sets a specific response for POST requests
func (m *MockRequestService) WithPostResponse(statusCode int, body string) *MockRequestService {
	m.PostResponse = &Response{
		StatusCode: statusCode,
		Body:       []byte(body),
	}
	return m
}

// WithError sets an API error response
func (m *MockRequestService) WithAPIError(statusCode int, message string) *MockRequestService {
	m.Error = &Error{
		StatusCode: statusCode,
		Message:    message,
	}
	return m
}
