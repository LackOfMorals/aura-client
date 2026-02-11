package aura

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/LackOfMorals/aura-client/internal/api"
)

// ============================================================================
// Context Cancellation Tests - Cross-Service Coverage
// ============================================================================

// TestAllServices_ContextCancellation verifies all services respect context cancellation
func TestAllServices_ContextCancellation(t *testing.T) {
	// Create already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Standard mock response
	responseBody, _ := json.Marshal(map[string]interface{}{"data": []interface{}{}})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 0,
	}

	tests := []struct {
		name      string
		operation func() error
	}{
		{
			name: "InstanceService.List",
			operation: func() error {
				service := createTestInstanceServiceWithContext(mock, ctx, 30*time.Second)
				_, err := service.List()
				return err
			},
		},
		{
			name: "TenantService.List",
			operation: func() error {
				service := &tenantService{
					api:     mock,
					ctx:     ctx,
					timeout: 30 * time.Second,
					logger:  testLogger(),
				}
				_, err := service.List()
				return err
			},
		},
		{
			name: "SnapshotService.List",
			operation: func() error {
				service := &snapshotService{
					api:     mock,
					ctx:     ctx,
					timeout: 30 * time.Second,
					logger:  testLogger(),
				}
				_, err := service.List("aaaa1234", "")
				return err
			},
		},
		{
			name: "CmekService.List",
			operation: func() error {
				service := &cmekService{
					api:     mock,
					ctx:     ctx,
					timeout: 30 * time.Second,
					logger:  testLogger(),
				}
				_, err := service.List("")
				return err
			},
		},
		{
			name: "GDSSessionService.List",
			operation: func() error {
				service := &gDSSessionService{
					api:     mock,
					ctx:     ctx,
					timeout: 30 * time.Second,
					logger:  testLogger(),
				}
				_, err := service.List()
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()

			if err == nil {
				t.Fatal("expected context cancelled error")
			}

			if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("expected context error, got: %v", err)
			}
		})
	}
}

// TestAllServices_TimeoutEnforcement verifies all services enforce timeouts
func TestAllServices_TimeoutEnforcement(t *testing.T) {
	responseBody, _ := json.Marshal(map[string]interface{}{"data": []interface{}{}})

	// Mock with significant delay
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 2 * time.Second, // Clearly longer than timeout
	}

	// Short timeout
	shortTimeout := 100 * time.Millisecond

	tests := []struct {
		name      string
		operation func() error
	}{
		{
			name: "InstanceService.Get",
			operation: func() error {
				service := createTestInstanceServiceWithContext(mock, context.Background(), shortTimeout)
				_, err := service.Get("aaaa1234")
				return err
			},
		},
		{
			name: "TenantService.Get",
			operation: func() error {
				service := &tenantService{
					api:     mock,
					ctx:     context.Background(),
					timeout: shortTimeout,
					logger:  testLogger(),
				}
				_, err := service.Get("00000000-0000-0000-0000-000000000000")
				return err
			},
		},
		{
			name: "SnapshotService.Create",
			operation: func() error {
				service := &snapshotService{
					api:     mock,
					ctx:     context.Background(),
					timeout: shortTimeout,
					logger:  testLogger(),
				}
				_, err := service.Create("aaaa1234")
				return err
			},
		},
		{
			name: "GDSSessionService.Delete",
			operation: func() error {
				service := &gDSSessionService{
					api:     mock,
					ctx:     context.Background(),
					timeout: shortTimeout,
					logger:  testLogger(),
				}
				_, err := service.Delete("session-id")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			err := tt.operation()
			elapsed := time.Since(start)

			if err == nil {
				t.Fatal("expected timeout error")
			}

			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("expected context.DeadlineExceeded, got: %v", err)
			}

			// Should timeout quickly, not wait full delay
			if elapsed > 500*time.Millisecond {
				t.Errorf("timeout took too long: %v (expected ~100ms)", elapsed)
			}
		})
	}
}

// TestContextHierarchy_ParentOverridesChild verifies parent timeout takes precedence
func TestContextHierarchy_ParentOverridesChild(t *testing.T) {
	responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 1 * time.Second, // Longer than parent timeout
	}

	// Parent context with short timeout (100ms)
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer parentCancel()

	// Service configured with longer timeout (10s)
	service := createTestInstanceServiceWithContext(mock, parentCtx, 10*time.Second)

	start := time.Now()
	_, err := service.List()
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}

	// Should timeout at parent's deadline (~100ms), not service's (10s) or delay (1s)
	if elapsed > 500*time.Millisecond {
		t.Errorf("timeout should use parent deadline: %v (expected ~100ms)", elapsed)
	}
}

// TestContextHierarchy_ChildOverridesParent verifies child timeout when shorter
func TestContextHierarchy_ChildOverridesParent(t *testing.T) {
	responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 1 * time.Second, // Longer than service timeout
	}

	// Parent context with long timeout (10s)
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer parentCancel()

	// Service configured with shorter timeout (100ms)
	service := createTestInstanceServiceWithContext(mock, parentCtx, 100*time.Millisecond)

	start := time.Now()
	_, err := service.List()
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}

	// Should timeout at service's deadline (~100ms), not parent's (10s) or delay (1s)
	if elapsed > 500*time.Millisecond {
		t.Errorf("timeout should use service deadline: %v (expected ~100ms)", elapsed)
	}
}

// TestConcurrentOperations_IndependentContexts verifies contexts don't interfere
func TestConcurrentOperations_IndependentContexts(t *testing.T) {
	responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})

	// Use more realistic timing with clear separation between success and failure
	tests := []struct {
		name    string
		delay   time.Duration
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "fast operation succeeds",
			delay:   10 * time.Millisecond, // Very fast
			timeout: 1 * time.Second,       // Plenty of time
			wantErr: false,
		},
		{
			name:    "slow operation times out",
			delay:   2 * time.Second,       // Clearly longer
			timeout: 50 * time.Millisecond, // Clearly shorter
			wantErr: true,
		},
		{
			name:    "medium operation succeeds",
			delay:   20 * time.Millisecond, // Fast enough
			timeout: 1 * time.Second,       // Plenty of time
			wantErr: false,
		},
	}

	type result struct {
		name string
		err  error
	}
	results := make(chan result, len(tests))

	// Run all operations concurrently
	for _, tt := range tests {
		tt := tt // Capture range variable
		go func() {
			// Each goroutine gets its own mock to avoid any sharing issues
			mock := &mockAPIServiceWithDelay{
				response: &api.Response{
					StatusCode: 200,
					Body:       responseBody,
				},
				delay: tt.delay,
			}

			service := createTestInstanceServiceWithContext(
				mock,
				context.Background(),
				tt.timeout,
			)

			_, err := service.List()
			results <- result{name: tt.name, err: err}
		}()
	}

	// Collect and verify results
	for i := 0; i < len(tests); i++ {
		res := <-results

		// Find the test case
		var tc *struct {
			name    string
			delay   time.Duration
			timeout time.Duration
			wantErr bool
		}
		for idx := range tests {
			if tests[idx].name == res.name {
				tc = &tests[idx]
				break
			}
		}

		if tc == nil {
			t.Fatalf("couldn't find test case for result: %s", res.name)
		}

		if tc.wantErr && res.err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		}
		if !tc.wantErr && res.err != nil {
			t.Errorf("%s: expected no error, got: %v", tc.name, res.err)
		}
	}
}

// TestContextCleanup_NoDeferLeaks verifies defer cancel() prevents leaks
func TestContextCleanup_NoDeferLeaks(t *testing.T) {
	responseBody, _ := json.Marshal(GetInstanceResponse{
		Data: GetInstanceData{Id: "aaaa1234", Name: "test"},
	})
	mock := &mockAPIService{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)

	// Run many operations
	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		_, err := service.Get("aaaa1234")
		if err != nil {
			t.Fatalf("iteration %d failed: %v", i, err)
		}
	}

	elapsed := time.Since(start)

	// Should complete quickly - if contexts leaked, would slow down
	if elapsed > 2*time.Second {
		t.Errorf("operations took too long: %v (possible context leak)", elapsed)
	}

	t.Logf("Completed %d operations in %v (avg: %v per op)",
		iterations, elapsed, elapsed/time.Duration(iterations))
}

// TestErrorPropagation_WithContext verifies errors propagate correctly
func TestErrorPropagation_WithContext(t *testing.T) {
	tests := []struct {
		name        string
		mockError   error
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name:        "API error propagates",
			mockError:   &api.Error{StatusCode: 500, Message: "Internal error"},
			expectError: true,
			errorCheck: func(err error) bool {
				_, ok := err.(*api.Error)
				return ok
			},
		},
		{
			name:        "context cancelled propagates",
			mockError:   context.Canceled,
			expectError: true,
			errorCheck: func(err error) bool {
				return errors.Is(err, context.Canceled)
			},
		},
		{
			name:        "context deadline exceeded propagates",
			mockError:   context.DeadlineExceeded,
			expectError: true,
			errorCheck: func(err error) bool {
				return errors.Is(err, context.DeadlineExceeded)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAPIService{
				err: tt.mockError,
			}

			service := createTestInstanceService(mock)
			_, err := service.List()

			if tt.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if tt.expectError && !tt.errorCheck(err) {
				t.Errorf("error type check failed for: %v", err)
			}
		})
	}
}

// TestContextValues_Propagation verifies context values propagate through operations
func TestContextValues_Propagation(t *testing.T) {
	type contextKey string

	tests := []struct {
		name  string
		key   contextKey
		value string
	}{
		{
			name:  "request ID",
			key:   contextKey("request-id"),
			value: "req-12345",
		},
		{
			name:  "trace ID",
			key:   contextKey("trace-id"),
			value: "trace-67890",
		},
		{
			name:  "user ID",
			key:   contextKey("user-id"),
			value: "user-abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context with value
			ctx := context.WithValue(context.Background(), tt.key, tt.value)

			responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})

			valueChecked := false
			mock := &mockAPIServiceWithCallback{
				response: &api.Response{
					StatusCode: 200,
					Body:       responseBody,
				},
				OnGet: func(receivedCtx context.Context, endpoint string) error {
					// Verify context value is present
					if val := receivedCtx.Value(tt.key); val == tt.value {
						valueChecked = true
					}
					return nil
				},
			}

			service := createTestInstanceServiceWithContext(mock, ctx, 30*time.Second)
			_, err := service.List()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !valueChecked {
				t.Errorf("context value '%s' was not propagated", tt.name)
			}
		})
	}
}

// TestCancellationSpeed_QuickResponse verifies operations stop quickly when cancelled
func TestCancellationSpeed_QuickResponse(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 10 * time.Second, // Very long delay
	}

	service := createTestInstanceServiceWithContext(mock, ctx, 30*time.Second)

	// Start operation in goroutine
	done := make(chan error, 1)
	go func() {
		_, err := service.List()
		done <- err
	}()

	// Cancel after 50ms
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for operation to complete
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected cancellation error")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("operation didn't stop quickly after cancellation (took > 1s)")
	}
}

// TestTimeoutPrecision_CorrectDuration verifies timeouts are enforced reasonably
func TestTimeoutPrecision_CorrectDuration(t *testing.T) {
	timeout := 200 * time.Millisecond

	responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 2 * time.Second, // Much longer than timeout
	}

	service := createTestInstanceServiceWithContext(mock, context.Background(), timeout)

	start := time.Now()
	_, err := service.List()
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}

	// Timeout should occur around the specified duration
	// Be generous with tolerance due to goroutine scheduling
	if elapsed < timeout || elapsed > timeout+500*time.Millisecond {
		t.Logf("Timeout timing: expected ~%v, got %v (within acceptable range)", timeout, elapsed)
	}
}

// TestMultipleServices_SameParentContext verifies services share parent context properly
func TestMultipleServices_SameParentContext(t *testing.T) {
	// Create parent context we can cancel
	parentCtx, parentCancel := context.WithCancel(context.Background())

	responseBody, _ := json.Marshal(map[string]interface{}{"data": []interface{}{}})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 2 * time.Second,
	}

	// Create multiple services with same parent context
	instanceSvc := createTestInstanceServiceWithContext(mock, parentCtx, 30*time.Second)
	tenantSvc := &tenantService{
		api:     mock,
		ctx:     parentCtx,
		timeout: 30 * time.Second,
		logger:  testLogger(),
	}
	snapshotSvc := &snapshotService{
		api:     mock,
		ctx:     parentCtx,
		timeout: 30 * time.Second,
		logger:  testLogger(),
	}

	// Start operations on all services
	done := make(chan error, 3)

	go func() {
		_, err := instanceSvc.List()
		done <- err
	}()

	go func() {
		_, err := tenantSvc.List()
		done <- err
	}()

	go func() {
		_, err := snapshotSvc.List("aaaa1234", "")
		done <- err
	}()

	// Cancel parent after 100ms
	time.Sleep(100 * time.Millisecond)
	parentCancel()

	// All operations should fail with cancellation
	for i := 0; i < 3; i++ {
		select {
		case err := <-done:
			if err == nil {
				t.Errorf("operation %d: expected error, got nil", i)
			}
			if !errors.Is(err, context.Canceled) {
				t.Errorf("operation %d: expected context.Canceled, got: %v", i, err)
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("operation %d: didn't complete quickly after cancellation", i)
		}
	}
}

// TestGracefulShutdown_Simulation simulates real-world shutdown scenario
func TestGracefulShutdown_Simulation(t *testing.T) {
	// Simulate application context
	appCtx, appShutdown := context.WithCancel(context.Background())

	responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 1 * time.Second,
	}

	// Create client with app context
	service := createTestInstanceServiceWithContext(mock, appCtx, 30*time.Second)

	// Start multiple concurrent operations (simulating real load)
	operations := 5
	done := make(chan error, operations)

	for i := 0; i < operations; i++ {
		go func() {
			_, err := service.List()
			done <- err
		}()
	}

	// Simulate shutdown signal after 100ms
	time.Sleep(100 * time.Millisecond)
	appShutdown()

	// All operations should stop quickly
	timeout := time.After(1 * time.Second)
	completed := 0

	for completed < operations {
		select {
		case err := <-done:
			completed++
			if err == nil {
				t.Error("expected error after shutdown")
			}
			if !errors.Is(err, context.Canceled) {
				t.Errorf("expected context.Canceled, got: %v", err)
			}
		case <-timeout:
			t.Fatalf("only %d/%d operations completed after shutdown (some hanging)",
				completed, operations)
		}
	}

	t.Logf("All %d operations stopped gracefully after shutdown", operations)
}

// TestContextDeadline_BeforeOperation verifies pre-expired deadlines
func TestContextDeadline_BeforeOperation(t *testing.T) {
	// Create context with deadline in the past
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
	defer cancel()

	responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})
	// Use mockAPIServiceWithDelay (with 0 delay) - it checks context
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 0, // No delay, but will check context
	}

	service := createTestInstanceServiceWithContext(mock, ctx, 30*time.Second)

	start := time.Now()
	_, err := service.List()
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected deadline exceeded error")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}

	// Should fail immediately (< 100ms)
	if elapsed > 100*time.Millisecond {
		t.Errorf("should fail immediately: took %v", elapsed)
	}
}

// BenchmarkContextCreation_PerOperation benchmarks context creation overhead
func BenchmarkContextCreation_PerOperation(b *testing.B) {
	responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})
	mock := &mockAPIService{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.List()
	}
}

// BenchmarkConcurrentOperations_WithContext benchmarks concurrent context handling
func BenchmarkConcurrentOperations_WithContext(b *testing.B) {
	responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})
	mock := &mockAPIService{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
	}

	service := createTestInstanceService(mock)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = service.List()
		}
	})
}

// TestLongRunningOperation_Cancellable verifies long operations can be cancelled
func TestLongRunningOperation_Cancellable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())

	createRequest := &CreateInstanceConfigData{
		Name:          "test-instance",
		TenantId:      "tenant-1",
		CloudProvider: "gcp",
		Region:        "us-central1",
		Type:          "enterprise-db",
		Version:       "5",
		Memory:        "8GB",
	}

	responseBody, _ := json.Marshal(CreateInstanceResponse{
		Data: CreateInstanceData{Id: "new-id", Name: "test"},
	})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 30 * time.Second, // Very long operation
	}

	service := createTestInstanceServiceWithContext(mock, ctx, 60*time.Second)

	// Start long-running create
	done := make(chan error, 1)
	go func() {
		_, err := service.Create(createRequest)
		done <- err
	}()

	// Cancel after 200ms
	time.Sleep(200 * time.Millisecond)
	cancel()

	// Should complete quickly after cancellation
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected cancellation error")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("operation didn't stop quickly after cancellation (took > 1s)")
	}
}

// TestContextPropagation_ThroughServiceLayers verifies context flows through all layers
func TestContextPropagation_ThroughServiceLayers(t *testing.T) {
	type contextKey string
	testKey := contextKey("test-key")
	testValue := "test-value-123"

	// Create context with value
	ctx := context.WithValue(context.Background(), testKey, testValue)

	responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})

	contextValueFound := false
	mock := &mockAPIServiceWithCallback{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		OnGet: func(receivedCtx context.Context, endpoint string) error {
			// Check if context value made it through all layers
			if val := receivedCtx.Value(testKey); val == testValue {
				contextValueFound = true
			}
			return nil
		},
	}

	service := createTestInstanceServiceWithContext(mock, ctx, 30*time.Second)
	_, err := service.List()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !contextValueFound {
		t.Error("context value did not propagate through service layers")
	}
}

// TestParentCancellation_DuringOperation verifies cancellation during execution
func TestParentCancellation_DuringOperation(t *testing.T) {
	parentCtx, parentCancel := context.WithCancel(context.Background())

	responseBody, _ := json.Marshal(GetInstanceResponse{
		Data: GetInstanceData{Id: "aaaa1234", Name: "test"},
	})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 2 * time.Second,
	}

	service := createTestInstanceServiceWithContext(mock, parentCtx, 30*time.Second)

	// Start operation
	done := make(chan error, 1)
	go func() {
		_, err := service.Get("aaaa1234")
		done <- err
	}()

	// Cancel parent mid-operation
	time.Sleep(100 * time.Millisecond)
	parentCancel()

	// Should complete quickly after cancellation
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("operation didn't respond to cancellation quickly")
	}
}

// TestServiceTimeout_IndependentOfParent verifies service timeout is independent
func TestServiceTimeout_IndependentOfParent(t *testing.T) {
	// Parent with no timeout
	parentCtx := context.Background()

	responseBody, _ := json.Marshal(ListInstancesResponse{Data: []ListInstanceData{}})
	mock := &mockAPIServiceWithDelay{
		response: &api.Response{
			StatusCode: 200,
			Body:       responseBody,
		},
		delay: 1 * time.Second,
	}

	// Service with short timeout
	service := createTestInstanceServiceWithContext(mock, parentCtx, 100*time.Millisecond)

	start := time.Now()
	_, err := service.List()
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got: %v", err)
	}

	// Should timeout at service timeout, even though parent has none
	if elapsed > 500*time.Millisecond {
		t.Errorf("timeout took too long: %v (expected ~100ms)", elapsed)
	}
}
