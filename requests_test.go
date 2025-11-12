package aura

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestAPIError_Error verifies error message formatting
func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name        string
		apiErr      *APIError
		expectedMsg string
	}{
		{
			name: "error without details",
			apiErr: &APIError{
				StatusCode: 404,
				Message:    "Not Found",
				Details:    nil,
			},
			expectedMsg: "API error (status 404): Not Found",
		},
		{
			name: "error with single detail",
			apiErr: &APIError{
				StatusCode: 400,
				Message:    "Bad Request",
				Details: []APIErrorDetail{
					{Message: "Invalid parameter"},
				},
			},
			expectedMsg: "API error (status 400): Bad Request - Invalid parameter",
		},
		{
			name: "error with multiple details",
			apiErr: &APIError{
				StatusCode: 422,
				Message:    "Validation Error",
				Details: []APIErrorDetail{
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
	apiErr := &APIError{
		StatusCode: 400,
		Message:    "Multiple errors occurred",
		Details: []APIErrorDetail{
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
		apiErr   *APIError
		expected bool
	}{
		{
			name: "no details",
			apiErr: &APIError{
				StatusCode: 404,
				Message:    "Not Found",
				Details:    nil,
			},
			expected: false,
		},
		{
			name: "single detail",
			apiErr: &APIError{
				StatusCode: 400,
				Message:    "Error",
				Details: []APIErrorDetail{
					{Message: "Single error"},
				},
			},
			expected: false,
		},
		{
			name: "multiple details",
			apiErr: &APIError{
				StatusCode: 422,
				Message:    "Error",
				Details: []APIErrorDetail{
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
		apiErr := &APIError{StatusCode: tt.statusCode}
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
		apiErr := &APIError{StatusCode: tt.statusCode}
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
		apiErr := &APIError{StatusCode: tt.statusCode}
		if apiErr.IsBadRequest() != tt.expected {
			t.Errorf("statusCode %d: expected IsBadRequest() = %v, got %v",
				tt.statusCode, tt.expected, apiErr.IsBadRequest())
		}
	}
}

// TestParseAPIError_EmptyResponse verifies handling of empty response
func TestParseAPIError_EmptyResponse(t *testing.T) {
	apiErr := parseAPIError([]byte{}, 500)

	if apiErr.StatusCode != 500 {
		t.Errorf("expected statusCode 500, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != http.StatusText(500) {
		t.Errorf("expected message '%s', got '%s'", http.StatusText(500), apiErr.Message)
	}
	if len(apiErr.Details) != 0 {
		t.Errorf("expected no details, got %d", len(apiErr.Details))
	}
}

// TestParseAPIError_ValidJSON_ErrorsField verifies parsing with 'errors' field
func TestParseAPIError_ValidJSON_ErrorsField(t *testing.T) {
	responseBody := `{
		"message": "Validation failed",
		"errors": [
			{"message": "Field required", "field": "name"},
			{"message": "Invalid format", "field": "email"}
		]
	}`

	apiErr := parseAPIError([]byte(responseBody), 400)

	if apiErr.StatusCode != 400 {
		t.Errorf("expected statusCode 400, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Validation failed" {
		t.Errorf("expected message 'Validation failed', got '%s'", apiErr.Message)
	}
	if len(apiErr.Details) != 2 {
		t.Errorf("expected 2 error details, got %d", len(apiErr.Details))
	}
	if apiErr.Details[0].Message != "Field required" {
		t.Errorf("expected first detail message 'Field required', got '%s'", apiErr.Details[0].Message)
	}
	if apiErr.Details[0].Field != "name" {
		t.Errorf("expected first detail field 'name', got '%s'", apiErr.Details[0].Field)
	}
}

// TestParseAPIError_ValidJSON_DetailsField verifies parsing with 'details' field
func TestParseAPIError_ValidJSON_DetailsField(t *testing.T) {
	responseBody := `{
		"message": "Operation failed",
		"details": [
			{"message": "Insufficient resources", "reason": "quota_exceeded"}
		]
	}`

	apiErr := parseAPIError([]byte(responseBody), 503)

	if apiErr.StatusCode != 503 {
		t.Errorf("expected statusCode 503, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Operation failed" {
		t.Errorf("expected message 'Operation failed', got '%s'", apiErr.Message)
	}
	if len(apiErr.Details) != 1 {
		t.Errorf("expected 1 error detail, got %d", len(apiErr.Details))
	}
	if apiErr.Details[0].Reason != "quota_exceeded" {
		t.Errorf("expected reason 'quota_exceeded', got '%s'", apiErr.Details[0].Reason)
	}
}

// TestParseAPIError_InvalidJSON verifies handling of malformed JSON
func TestParseAPIError_InvalidJSON(t *testing.T) {
	responseBody := `{"invalid json`

	apiErr := parseAPIError([]byte(responseBody), 500)

	// Should still create APIError with status code and default message
	if apiErr.StatusCode != 500 {
		t.Errorf("expected statusCode 500, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != http.StatusText(500) {
		t.Errorf("expected default message, got '%s'", apiErr.Message)
	}
}

// TestAPIError_TypeAssertion verifies error can be type asserted
func TestAPIError_TypeAssertion(t *testing.T) {
	var err error = &APIError{
		StatusCode: 404,
		Message:    "Not Found",
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatal("failed to type assert error to *APIError")
	}

	if !apiErr.IsNotFound() {
		t.Error("expected IsNotFound() to be true")
	}
}

// TestAPIError_JSONMarshaling verifies APIError can be marshaled to JSON
func TestAPIError_JSONMarshaling(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 400,
		Message:    "Test error",
		Details: []APIErrorDetail{
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

	var unmarshaled APIError
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
