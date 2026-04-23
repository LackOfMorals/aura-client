// Package testutil provides mock implementations shared across internal test packages.
//
// MockRequestService was removed from this file because it imported
// internal/api, creating an import cycle when api's own test files imported
// testutil. The top-level service tests define their own lightweight mocks
// (mockAPIService in test_helpers.go) which are sufficient for their needs.
package testutil
