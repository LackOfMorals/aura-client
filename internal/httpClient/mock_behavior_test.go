// Package httpClient_test contains tests of the testutil.MockHTTPService that
// are kept here rather than in the httpClient package itself to avoid an
// import cycle:
//
//   httpClient (white-box test) → testutil → httpClient   ← cycle
//   httpClient_test (external)  → testutil → httpClient   ← no cycle
package httpClient_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/LackOfMorals/aura-client/internal/testutil"
)

func TestMockHTTPService_Get(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(200, `{"status":"ok"}`)

	resp, err := mock.Get(context.Background(), "/test", map[string]string{"X-Test": "value"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if mock.LastMethod != "GET" {
		t.Errorf("expected method GET, got %s", mock.LastMethod)
	}
	if mock.LastURL != "/test" {
		t.Errorf("expected URL /test, got %s", mock.LastURL)
	}
	if mock.CallCount != 1 {
		t.Errorf("expected call count 1, got %d", mock.CallCount)
	}
}

func TestMockHTTPService_Post(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithPostResponse(201, `{"id":"123"}`)

	resp, err := mock.Post(context.Background(), "/test", nil, `{"name":"test"}`)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
	if mock.LastBody != `{"name":"test"}` {
		t.Errorf("expected body, got '%s'", mock.LastBody)
	}
}

func TestMockHTTPService_Error(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	expected := fmt.Errorf("network error")
	mock.WithError(expected)

	_, err := mock.Get(context.Background(), "/test", nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != expected {
		t.Errorf("expected %v, got %v", expected, err)
	}
}

func TestMockHTTPService_Reset(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(200, "test")
	mock.Get(context.Background(), "/test", nil) //nolint:errcheck

	mock.Reset()

	if mock.CallCount != 0 {
		t.Errorf("expected call count 0 after reset, got %d", mock.CallCount)
	}
	if mock.Response != nil {
		t.Error("expected nil response after reset")
	}
	if len(mock.CallHistory) != 0 {
		t.Errorf("expected empty call history after reset, got %d", len(mock.CallHistory))
	}
}

func TestMockHTTPService_CallHistory(t *testing.T) {
	mock := testutil.NewMockHTTPService()
	mock.WithResponse(200, "ok")

	ctx := context.Background()
	mock.Get(ctx, "/test1", map[string]string{"X-Test": "1"})    //nolint:errcheck
	mock.Post(ctx, "/test2", map[string]string{"X-Test": "2"}, "body") //nolint:errcheck
	mock.Delete(ctx, "/test3", nil)                              //nolint:errcheck

	if len(mock.CallHistory) != 3 {
		t.Fatalf("expected 3 calls in history, got %d", len(mock.CallHistory))
	}
	if mock.CallHistory[0].Method != "GET" {
		t.Errorf("expected first call GET, got %s", mock.CallHistory[0].Method)
	}
	if mock.CallHistory[1].Method != "POST" {
		t.Errorf("expected second call POST, got %s", mock.CallHistory[1].Method)
	}
	if mock.CallHistory[1].Body != "body" {
		t.Errorf("expected second call body 'body', got '%s'", mock.CallHistory[1].Body)
	}
	if mock.CallHistory[2].Method != "DELETE" {
		t.Errorf("expected third call DELETE, got %s", mock.CallHistory[2].Method)
	}
}
