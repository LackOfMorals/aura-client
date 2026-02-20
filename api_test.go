package aura

import (
	"encoding/json"
	"testing"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// TestError_Error verifies error message formatting
func TestError_Error(t *testing.T) {
	tests := []struct {
		name        string
		apiErr      *api.Error
		expectedMsg string
	}{
		{
			name: "error without details",
			apiErr: &api.Error{
				StatusCode: 404,
				Message:    "Not Found",
				Details:    nil,
			},
			expectedMsg: "API error (status 404): Not Found",
		},
		{
			name: "error with single detail",
			apiErr: &api.Error{
				StatusCode: 400,
				Message:    "Bad Request",
				Details: []api.ErrorDetail{
					{Message: "Invalid parameter"},
				},
			},
			expectedMsg: "API error (status 400): Bad Request - Invalid parameter",
		},
		{
			name: "error with multiple details",
			apiErr: &api.Error{
				StatusCode: 422,
				Message:    "Validation Error",
				Details: []api.ErrorDetail{
					{Message: "Field 'name' is required"},
					{Message: "Field 'region' is invalid"},
					{Message: "Field 'memory' must be positive"},
				},
			},
			expectedMsg: "API error (status 422): Validation Error - Field 'name' is required (and 2 more error(s))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.apiErr.Error() != tt.expectedMsg {
				t.Errorf("expected '%s', got '%s'", tt.expectedMsg, tt.apiErr.Error())
			}
		})
	}
}

// TestError_AllErrors verifies getting all error messages
func TestError_AllErrors(t *testing.T) {
	apiErr := &api.Error{
		StatusCode: 400,
		Message:    "Multiple errors occurred",
		Details: []api.ErrorDetail{
			{Message: "Error 1"},
			{Message: "Error 2"},
			{Message: "Error 3"},
		},
	}

	allErrs := apiErr.AllErrors()

	if len(allErrs) != 4 { // Main message + 3 details
		t.Errorf("expected 4 errors, got %d", len(allErrs))
	}

	if allErrs[0] != "Multiple errors occurred" {
		t.Errorf("expected first error to be main message, got '%s'", allErrs[0])
	}
	if allErrs[1] != "Error 1" {
		t.Errorf("expected second error to be 'Error 1', got '%s'", allErrs[1])
	}
}

// TestError_HasMultipleErrors verifies multiple error detection
func TestError_HasMultipleErrors(t *testing.T) {
	tests := []struct {
		name     string
		apiErr   *api.Error
		expected bool
	}{
		{
			name: "no details",
			apiErr: &api.Error{
				StatusCode: 404,
				Message:    "Not Found",
				Details:    nil,
			},
			expected: false,
		},
		{
			name: "single detail",
			apiErr: &api.Error{
				StatusCode: 400,
				Message:    "Error",
				Details: []api.ErrorDetail{
					{Message: "Single error"},
				},
			},
			expected: false,
		},
		{
			name: "multiple details",
			apiErr: &api.Error{
				StatusCode: 422,
				Message:    "Error",
				Details: []api.ErrorDetail{
					{Message: "Error 1"},
					{Message: "Error 2"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.apiErr.HasMultipleErrors() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tt.apiErr.HasMultipleErrors())
			}
		})
	}
}

// TestError_IsNotFound verifies 404 detection
func TestError_IsNotFound(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{statusCode: 404, expected: true},
		{statusCode: 200, expected: false},
		{statusCode: 400, expected: false},
		{statusCode: 401, expected: false},
		{statusCode: 500, expected: false},
	}

	for _, tt := range tests {
		apiErr := &api.Error{StatusCode: tt.statusCode}
		if apiErr.IsNotFound() != tt.expected {
			t.Errorf("statusCode %d: expected IsNotFound() = %v, got %v",
				tt.statusCode, tt.expected, apiErr.IsNotFound())
		}
	}
}

// TestError_IsUnauthorized verifies 401 detection
func TestError_IsUnauthorized(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{statusCode: 401, expected: true},
		{statusCode: 200, expected: false},
		{statusCode: 403, expected: false},
		{statusCode: 404, expected: false},
		{statusCode: 500, expected: false},
	}

	for _, tt := range tests {
		apiErr := &api.Error{StatusCode: tt.statusCode}
		if apiErr.IsUnauthorized() != tt.expected {
			t.Errorf("statusCode %d: expected IsUnauthorized() = %v, got %v",
				tt.statusCode, tt.expected, apiErr.IsUnauthorized())
		}
	}
}

// TestError_IsBadRequest verifies 400 detection
func TestError_IsBadRequest(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{statusCode: 400, expected: true},
		{statusCode: 200, expected: false},
		{statusCode: 401, expected: false},
		{statusCode: 404, expected: false},
		{statusCode: 422, expected: false},
	}

	for _, tt := range tests {
		apiErr := &api.Error{StatusCode: tt.statusCode}
		if apiErr.IsBadRequest() != tt.expected {
			t.Errorf("statusCode %d: expected IsBadRequest() = %v, got %v",
				tt.statusCode, tt.expected, apiErr.IsBadRequest())
		}
	}
}

// TestError_TypeAssertion verifies error can be type asserted
func TestError_TypeAssertion(t *testing.T) {
	var err error = &api.Error{
		StatusCode: 404,
		Message:    "Not Found",
	}

	apiErr, ok := err.(*api.Error)
	if !ok {
		t.Fatal("failed to type assert error to *Error")
	}

	if !apiErr.IsNotFound() {
		t.Error("expected IsNotFound() to be true")
	}
}

// TestError_JSONMarshaling verifies Error can be marshaled to JSON
func TestError_JSONMarshaling(t *testing.T) {
	apiErr := &api.Error{
		StatusCode: 400,
		Message:    "Test error",
		Details: []api.ErrorDetail{
			{
				Message: "Detail message",
				Reason:  "test_reason",
				Field:   "test_field",
			},
		},
	}

	data, err := json.Marshal(apiErr)
	if err != nil {
		t.Fatalf("failed to marshal Error: %v", err)
	}

	var unmarshaled api.Error
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal Error: %v", err)
	}

	if unmarshaled.StatusCode != apiErr.StatusCode {
		t.Errorf("expected statusCode %d, got %d", apiErr.StatusCode, unmarshaled.StatusCode)
	}
	if unmarshaled.Message != apiErr.Message {
		t.Errorf("expected message '%s', got '%s'", apiErr.Message, unmarshaled.Message)
	}
}

// TestErrorAlias verifies the type alias works correctly
func TestErrorAlias(t *testing.T) {
	// Test that the type alias works
	var err Error = api.Error{
		StatusCode: 404,
		Message:    "Test",
	}

	if err.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", err.StatusCode)
	}
}
