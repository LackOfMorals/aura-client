package aura

import (
	"encoding/json"
	"testing"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// TestAPIError_Error verifies error message formatting
func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name        string
		apiErr      *api.APIError
		expectedMsg string
	}{
		{
			name: "error without details",
			apiErr: &api.APIError{
				StatusCode: 404,
				Message:    "Not Found",
				Details:    nil,
			},
			expectedMsg: "API error (status 404): Not Found",
		},
		{
			name: "error with single detail",
			apiErr: &api.APIError{
				StatusCode: 400,
				Message:    "Bad Request",
				Details: []api.APIErrorDetail{
					{Message: "Invalid parameter"},
				},
			},
			expectedMsg: "API error (status 400): Bad Request - Invalid parameter",
		},
		{
			name: "error with multiple details",
			apiErr: &api.APIError{
				StatusCode: 422,
				Message:    "Validation Error",
				Details: []api.APIErrorDetail{
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

// TestAPIError_AllErrors verifies getting all error messages
func TestAPIError_AllErrors(t *testing.T) {
	apiErr := &api.APIError{
		StatusCode: 400,
		Message:    "Multiple errors occurred",
		Details: []api.APIErrorDetail{
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

// TestAPIError_HasMultipleErrors verifies multiple error detection
func TestAPIError_HasMultipleErrors(t *testing.T) {
	tests := []struct {
		name     string
		apiErr   *api.APIError
		expected bool
	}{
		{
			name: "no details",
			apiErr: &api.APIError{
				StatusCode: 404,
				Message:    "Not Found",
				Details:    nil,
			},
			expected: false,
		},
		{
			name: "single detail",
			apiErr: &api.APIError{
				StatusCode: 400,
				Message:    "Error",
				Details: []api.APIErrorDetail{
					{Message: "Single error"},
				},
			},
			expected: false,
		},
		{
			name: "multiple details",
			apiErr: &api.APIError{
				StatusCode: 422,
				Message:    "Error",
				Details: []api.APIErrorDetail{
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

// TestAPIError_IsNotFound verifies 404 detection
func TestAPIError_IsNotFound(t *testing.T) {
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
		apiErr := &api.APIError{StatusCode: tt.statusCode}
		if apiErr.IsNotFound() != tt.expected {
			t.Errorf("statusCode %d: expected IsNotFound() = %v, got %v",
				tt.statusCode, tt.expected, apiErr.IsNotFound())
		}
	}
}

// TestAPIError_IsUnauthorized verifies 401 detection
func TestAPIError_IsUnauthorized(t *testing.T) {
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
		apiErr := &api.APIError{StatusCode: tt.statusCode}
		if apiErr.IsUnauthorized() != tt.expected {
			t.Errorf("statusCode %d: expected IsUnauthorized() = %v, got %v",
				tt.statusCode, tt.expected, apiErr.IsUnauthorized())
		}
	}
}

// TestAPIError_IsBadRequest verifies 400 detection
func TestAPIError_IsBadRequest(t *testing.T) {
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
		apiErr := &api.APIError{StatusCode: tt.statusCode}
		if apiErr.IsBadRequest() != tt.expected {
			t.Errorf("statusCode %d: expected IsBadRequest() = %v, got %v",
				tt.statusCode, tt.expected, apiErr.IsBadRequest())
		}
	}
}

// TestAPIError_TypeAssertion verifies error can be type asserted
func TestAPIError_TypeAssertion(t *testing.T) {
	var err error = &api.APIError{
		StatusCode: 404,
		Message:    "Not Found",
	}

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatal("failed to type assert error to *APIError")
	}

	if !apiErr.IsNotFound() {
		t.Error("expected IsNotFound() to be true")
	}
}

// TestAPIError_JSONMarshaling verifies APIError can be marshaled to JSON
func TestAPIError_JSONMarshaling(t *testing.T) {
	apiErr := &api.APIError{
		StatusCode: 400,
		Message:    "Test error",
		Details: []api.APIErrorDetail{
			{
				Message: "Detail message",
				Reason:  "test_reason",
				Field:   "test_field",
			},
		},
	}

	data, err := json.Marshal(apiErr)
	if err != nil {
		t.Fatalf("failed to marshal APIError: %v", err)
	}

	var unmarshaled api.APIError
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal APIError: %v", err)
	}

	if unmarshaled.StatusCode != apiErr.StatusCode {
		t.Errorf("expected statusCode %d, got %d", apiErr.StatusCode, unmarshaled.StatusCode)
	}
	if unmarshaled.Message != apiErr.Message {
		t.Errorf("expected message '%s', got '%s'", apiErr.Message, unmarshaled.Message)
	}
}

// TestAPIErrorAlias verifies the type alias works correctly
func TestAPIErrorAlias(t *testing.T) {
	// Test that the type alias works
	var err APIError = api.APIError{
		StatusCode: 404,
		Message:    "Test",
	}

	if err.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", err.StatusCode)
	}
}
