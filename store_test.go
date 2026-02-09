package aura

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to create a temporary test database
func setupTestStore(t *testing.T) (*storeService, string) {
	t.Helper()

	// Create temp directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test-store.db")

	// Create logger
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	// Create store service
	store, err := newStoreService(context.Background(), dbPath, logger)
	if err != nil {
		t.Fatalf("failed to create store service: %v", err)
	}

	return store, dbPath
}

// TestStoreService_Create verifies creating new configurations
func TestStoreService_Create(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	config := &CreateInstanceConfigData{
		Name:          "test-instance",
		TenantId:      "tenant-123",
		CloudProvider: "gcp",
		Region:        "us-central1",
		Type:          "enterprise-db",
		Version:       "5",
		Memory:        "8GB",
	}

	err := store.Create("test-label", config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it was stored
	retrieved, err := store.Read("test-label")
	if err != nil {
		t.Fatalf("expected to read config, got error: %v", err)
	}

	if retrieved.Name != config.Name {
		t.Errorf("expected name '%s', got '%s'", config.Name, retrieved.Name)
	}
	if retrieved.TenantId != config.TenantId {
		t.Errorf("expected tenant ID '%s', got '%s'", config.TenantId, retrieved.TenantId)
	}
}

// TestStoreService_Create_DuplicateLabel verifies error on duplicate labels
func TestStoreService_Create_DuplicateLabel(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	config := &CreateInstanceConfigData{
		Name:          "test-instance",
		TenantId:      "tenant-123",
		CloudProvider: "gcp",
		Region:        "us-central1",
		Type:          "enterprise-db",
		Version:       "5",
		Memory:        "8GB",
	}

	// Create first time
	err := store.Create("duplicate", config)
	if err != nil {
		t.Fatalf("expected no error on first create, got %v", err)
	}

	// Try to create again with same label
	err = store.Create("duplicate", config)
	if err == nil {
		t.Fatal("expected error on duplicate label, got nil")
	}

	var storeErr *StoreError
	if !errors.As(err, &storeErr) {
		t.Fatalf("expected StoreError, got %T", err)
	}

	if !errors.Is(err, ErrConfigAlreadyExists) {
		t.Errorf("expected ErrConfigAlreadyExists, got %v", err)
	}
}

// TestStoreService_Create_EmptyLabel verifies error on empty label
func TestStoreService_Create_EmptyLabel(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	config := &CreateInstanceConfigData{
		Name: "test",
	}

	err := store.Create("", config)
	if err == nil {
		t.Fatal("expected error for empty label, got nil")
	}

	if !errors.Is(err, ErrInvalidLabel) {
		t.Errorf("expected ErrInvalidLabel, got %v", err)
	}
}

// TestStoreService_Create_NilConfig verifies error on nil config
func TestStoreService_Create_NilConfig(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	err := store.Create("test-label", nil)
	if err == nil {
		t.Fatal("expected error for nil config, got nil")
	}

	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig, got %v", err)
	}
}

// TestStoreService_Read verifies reading configurations
func TestStoreService_Read(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	config := &CreateInstanceConfigData{
		Name:          "read-test",
		TenantId:      "tenant-456",
		CloudProvider: "aws",
		Region:        "us-east-1",
		Type:          "professional-db",
		Version:       "5",
		Memory:        "4GB",
	}

	// Create config
	if err := store.Create("read-label", config); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Read it back
	retrieved, err := store.Read("read-label")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.Name != "read-test" {
		t.Errorf("expected name 'read-test', got '%s'", retrieved.Name)
	}
	if retrieved.Region != "us-east-1" {
		t.Errorf("expected region 'us-east-1', got '%s'", retrieved.Region)
	}
}

// TestStoreService_Read_NotFound verifies error on missing config
func TestStoreService_Read_NotFound(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	_, err := store.Read("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent config, got nil")
	}

	if !errors.Is(err, ErrConfigNotFound) {
		t.Errorf("expected ErrConfigNotFound, got %v", err)
	}
}

// TestStoreService_Read_EmptyLabel verifies error on empty label
func TestStoreService_Read_EmptyLabel(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	_, err := store.Read("")
	if err == nil {
		t.Fatal("expected error for empty label, got nil")
	}

	if !errors.Is(err, ErrInvalidLabel) {
		t.Errorf("expected ErrInvalidLabel, got %v", err)
	}
}

// TestStoreService_Update verifies updating configurations
func TestStoreService_Update(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	// Create initial config
	initialConfig := &CreateInstanceConfigData{
		Name:          "initial-name",
		TenantId:      "tenant-789",
		CloudProvider: "gcp",
		Region:        "us-west1",
		Type:          "enterprise-db",
		Version:       "5",
		Memory:        "8GB",
	}

	if err := store.Create("update-label", initialConfig); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Update the config
	updatedConfig := &CreateInstanceConfigData{
		Name:          "updated-name",
		TenantId:      "tenant-789",
		CloudProvider: "aws",
		Region:        "eu-west-1",
		Type:          "professional-db",
		Version:       "5",
		Memory:        "16GB",
	}

	err := store.Update("update-label", updatedConfig)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it was updated
	retrieved, err := store.Read("update-label")
	if err != nil {
		t.Fatalf("failed to read updated config: %v", err)
	}

	if retrieved.Name != "updated-name" {
		t.Errorf("expected name 'updated-name', got '%s'", retrieved.Name)
	}
	if retrieved.Region != "eu-west-1" {
		t.Errorf("expected region 'eu-west-1', got '%s'", retrieved.Region)
	}
	if retrieved.Memory != "16GB" {
		t.Errorf("expected memory '16GB', got '%s'", retrieved.Memory)
	}
}

// TestStoreService_Update_NotFound verifies error on updating nonexistent config
func TestStoreService_Update_NotFound(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	config := &CreateInstanceConfigData{
		Name: "test",
	}

	err := store.Update("nonexistent", config)
	if err == nil {
		t.Fatal("expected error for nonexistent config, got nil")
	}

	if !errors.Is(err, ErrConfigNotFound) {
		t.Errorf("expected ErrConfigNotFound, got %v", err)
	}
}

// TestStoreService_Update_EmptyLabel verifies error on empty label
func TestStoreService_Update_EmptyLabel(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	config := &CreateInstanceConfigData{
		Name: "test",
	}

	err := store.Update("", config)
	if err == nil {
		t.Fatal("expected error for empty label, got nil")
	}

	if !errors.Is(err, ErrInvalidLabel) {
		t.Errorf("expected ErrInvalidLabel, got %v", err)
	}
}

// TestStoreService_Update_NilConfig verifies error on nil config
func TestStoreService_Update_NilConfig(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	err := store.Update("test-label", nil)
	if err == nil {
		t.Fatal("expected error for nil config, got nil")
	}

	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig, got %v", err)
	}
}

// TestStoreService_Delete verifies deleting configurations
func TestStoreService_Delete(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	// Create a config
	config := &CreateInstanceConfigData{
		Name:          "delete-test",
		TenantId:      "tenant-delete",
		CloudProvider: "gcp",
		Region:        "us-central1",
		Type:          "enterprise-db",
		Version:       "5",
		Memory:        "8GB",
	}

	if err := store.Create("delete-label", config); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Delete it
	err := store.Delete("delete-label")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it's gone
	_, err = store.Read("delete-label")
	if err == nil {
		t.Fatal("expected error reading deleted config, got nil")
	}

	if !errors.Is(err, ErrConfigNotFound) {
		t.Errorf("expected ErrConfigNotFound, got %v", err)
	}
}

// TestStoreService_Delete_NotFound verifies error on deleting nonexistent config
func TestStoreService_Delete_NotFound(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	err := store.Delete("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent config, got nil")
	}

	if !errors.Is(err, ErrConfigNotFound) {
		t.Errorf("expected ErrConfigNotFound, got %v", err)
	}
}

// TestStoreService_Delete_EmptyLabel verifies error on empty label
func TestStoreService_Delete_EmptyLabel(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	err := store.Delete("")
	if err == nil {
		t.Fatal("expected error for empty label, got nil")
	}

	if !errors.Is(err, ErrInvalidLabel) {
		t.Errorf("expected ErrInvalidLabel, got %v", err)
	}
}

// TestStoreService_List verifies listing all configurations
func TestStoreService_List(t *testing.T) {
	store, _ := setupTestStore(t)
	defer store.Close()

	// Initially empty
	labels, err := store.List()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(labels) != 0 {
		t.Errorf("expected empty list, got %d items", len(labels))
	}

	// Add some configs
	configs := map[string]*CreateInstanceConfigData{
		"config-a": {Name: "a", TenantId: "t1", CloudProvider: "gcp", Region: "us", Type: "ent", Version: "5", Memory: "8GB"},
		"config-b": {Name: "b", TenantId: "t2", CloudProvider: "aws", Region: "eu", Type: "pro", Version: "5", Memory: "4GB"},
		"config-c": {Name: "c", TenantId: "t3", CloudProvider: "azure", Region: "us", Type: "ent", Version: "5", Memory: "16GB"},
	}

	for label, config := range configs {
		if err := store.Create(label, config); err != nil {
			t.Fatalf("failed to create config %s: %v", label, err)
		}
	}

	// List them
	labels, err = store.List()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(labels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(labels))
	}

	// Verify labels are in alphabetical order
	expectedLabels := []string{"config-a", "config-b", "config-c"}
	for i, expected := range expectedLabels {
		if i >= len(labels) {
			t.Errorf("missing label at index %d: expected '%s'", i, expected)
			continue
		}
		if labels[i] != expected {
			t.Errorf("expected label '%s' at index %d, got '%s'", expected, i, labels[i])
		}
	}
}

// TestStoreService_Persistence verifies data persists across store instances
func TestStoreService_Persistence(t *testing.T) {
	// Create temp directory that won't be cleaned up immediately
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "persist-test.db")

	// Create logger
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	// Create first store instance and add data
	store1, err := newStoreService(context.Background(), dbPath, logger)
	if err != nil {
		t.Fatalf("failed to create first store: %v", err)
	}

	config := &CreateInstanceConfigData{
		Name:          "persist-test",
		TenantId:      "tenant-persist",
		CloudProvider: "gcp",
		Region:        "us-central1",
		Type:          "enterprise-db",
		Version:       "5",
		Memory:        "8GB",
	}

	if err := store1.Create("persist-label", config); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Close first instance
	store1.Close()

	// Create second store instance with same database
	store2, err := newStoreService(context.Background(), dbPath, logger)
	if err != nil {
		t.Fatalf("failed to create second store: %v", err)
	}
	defer store2.Close()

	// Verify data persisted
	retrieved, err := store2.Read("persist-label")
	if err != nil {
		t.Fatalf("failed to read config from second store: %v", err)
	}

	if retrieved.Name != "persist-test" {
		t.Errorf("expected name 'persist-test', got '%s'", retrieved.Name)
	}
}

// TestStoreService_Close verifies database connection cleanup
func TestStoreService_Close(t *testing.T) {
	store, _ := setupTestStore(t)

	err := store.Close()
	if err != nil {
		t.Errorf("expected no error on close, got %v", err)
	}

	// Calling close again should be safe
	err = store.Close()
	if err != nil {
		t.Errorf("expected no error on second close, got %v", err)
	}
}
