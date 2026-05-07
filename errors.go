package aura

import (
	"github.com/LackOfMorals/aura-client/internal/api"
)

// Error represents an error response from the Aura API.
type Error = api.Error

// ErrorDetail represents individual error details.
type ErrorDetail = api.ErrorDetail

// Sentinel errors for common API failure modes. Callers can match them with
// errors.Is even when the underlying *Error is wrapped by service code:
//
//	_, err := client.Instances.Get(ctx, id)
//	if errors.Is(err, aura.ErrNotFound) {
//	    // handle missing instance
//	}
//
// To inspect status code, message, or detail fields, unwrap with errors.As:
//
//	var apiErr *aura.Error
//	if errors.As(err, &apiErr) {
//	    fmt.Println(apiErr.StatusCode, apiErr.Message)
//	}
var (
	ErrBadRequest      = api.ErrBadRequest
	ErrUnauthorized    = api.ErrUnauthorized
	ErrForbidden       = api.ErrForbidden
	ErrNotFound        = api.ErrNotFound
	ErrConflict        = api.ErrConflict
	ErrTooManyRequests = api.ErrTooManyRequests
	ErrInternalServer  = api.ErrInternalServer
	ErrServiceUnavail  = api.ErrServiceUnavail
)
