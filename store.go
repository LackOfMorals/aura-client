package aura

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// StoreError represents an error from the store service
type StoreError struct {
	Op      string // Operation that failed
	Label   string // Configuration label (if applicable)
	Message string // Error message
	Err     error  // Underlying error
}

func (e *StoreError) Error() string {
	if e.Label != "" {
		return fmt.Sprintf("store %s failed for label '%s': %s", e.Op, e.Label, e.Message)
	}
	return fmt.Sprintf("store %s failed: %s", e.Op, e.Message)
}

func (e *StoreError) Unwrap() error {
	return e.Err
}

// Common store errors
var (
	ErrConfigNotFound      = errors.New("configuration not found")
	ErrConfigAlreadyExists = errors.New("configuration already exists")
	ErrInvalidLabel        = errors.New("label cannot be empty")
	ErrInvalidConfig       = errors.New("configuration cannot be nil")
)

// storeService handles instance configuration storage operations
type storeService struct {
	db     *sql.DB
	ctx    context.Context
	logger *slog.Logger
}

// newStoreService creates a new store service with SQLite backend
func newStoreService(ctx context.Context, dbPath string, logger *slog.Logger) (*storeService, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, &StoreError{
			Op:      "initialize",
			Message: "failed to create database directory",
			Err:     err,
		}
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, &StoreError{
			Op:      "initialize",
			Message: "failed to open database",
			Err:     err,
		}
	}

	// Create table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS instance_configs (
		label TEXT PRIMARY KEY,
		config TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
		db.Close()
		return nil, &StoreError{
			Op:      "initialize",
			Message: "failed to create table",
			Err:     err,
		}
	}

	logger.InfoContext(ctx, "store service initialized", slog.String("dbPath", dbPath))

	return &storeService{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (s *storeService) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Create stores a new instance configuration with the given label
func (s *storeService) Create(label string, config *CreateInstanceConfigData) error {
	if label == "" {
		return &StoreError{
			Op:      "create",
			Message: "label cannot be empty",
			Err:     ErrInvalidLabel,
		}
	}

	if config == nil {
		return &StoreError{
			Op:      "create",
			Label:   label,
			Message: "configuration cannot be nil",
			Err:     ErrInvalidConfig,
		}
	}

	s.logger.DebugContext(s.ctx, "creating configuration", slog.String("label", label))

	// Serialize config to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return &StoreError{
			Op:      "create",
			Label:   label,
			Message: "failed to serialize configuration",
			Err:     err,
		}
	}

	// Insert into database
	insertSQL := `INSERT INTO instance_configs (label, config) VALUES (?, ?)`
	_, err = s.db.ExecContext(s.ctx, insertSQL, label, string(configJSON))
	if err != nil {
		// Check if it's a duplicate key error
		if err.Error() == "UNIQUE constraint failed: instance_configs.label" {
			return &StoreError{
				Op:      "create",
				Label:   label,
				Message: "configuration with this label already exists",
				Err:     ErrConfigAlreadyExists,
			}
		}
		return &StoreError{
			Op:      "create",
			Label:   label,
			Message: "failed to insert configuration",
			Err:     err,
		}
	}

	s.logger.InfoContext(s.ctx, "configuration created", slog.String("label", label))
	return nil
}

// Read retrieves an instance configuration by label
func (s *storeService) Read(label string) (*CreateInstanceConfigData, error) {
	if label == "" {
		return nil, &StoreError{
			Op:      "read",
			Message: "label cannot be empty",
			Err:     ErrInvalidLabel,
		}
	}

	s.logger.DebugContext(s.ctx, "reading configuration", slog.String("label", label))

	// Query database
	selectSQL := `SELECT config FROM instance_configs WHERE label = ?`
	var configJSON string
	err := s.db.QueryRowContext(s.ctx, selectSQL, label).Scan(&configJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &StoreError{
				Op:      "read",
				Label:   label,
				Message: "configuration not found",
				Err:     ErrConfigNotFound,
			}
		}
		return nil, &StoreError{
			Op:      "read",
			Label:   label,
			Message: "failed to query configuration",
			Err:     err,
		}
	}

	// Deserialize JSON
	var config CreateInstanceConfigData
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, &StoreError{
			Op:      "read",
			Label:   label,
			Message: "failed to deserialize configuration",
			Err:     err,
		}
	}

	s.logger.DebugContext(s.ctx, "configuration read", slog.String("label", label))
	return &config, nil
}

// Update modifies an existing instance configuration
func (s *storeService) Update(label string, config *CreateInstanceConfigData) error {
	if label == "" {
		return &StoreError{
			Op:      "update",
			Message: "label cannot be empty",
			Err:     ErrInvalidLabel,
		}
	}

	if config == nil {
		return &StoreError{
			Op:      "update",
			Label:   label,
			Message: "configuration cannot be nil",
			Err:     ErrInvalidConfig,
		}
	}

	s.logger.DebugContext(s.ctx, "updating configuration", slog.String("label", label))

	// Serialize config to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return &StoreError{
			Op:      "update",
			Label:   label,
			Message: "failed to serialize configuration",
			Err:     err,
		}
	}

	// Update database
	updateSQL := `UPDATE instance_configs SET config = ?, updated_at = CURRENT_TIMESTAMP WHERE label = ?`
	result, err := s.db.ExecContext(s.ctx, updateSQL, string(configJSON), label)
	if err != nil {
		return &StoreError{
			Op:      "update",
			Label:   label,
			Message: "failed to update configuration",
			Err:     err,
		}
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &StoreError{
			Op:      "update",
			Label:   label,
			Message: "failed to check update result",
			Err:     err,
		}
	}

	if rowsAffected == 0 {
		return &StoreError{
			Op:      "update",
			Label:   label,
			Message: "configuration not found",
			Err:     ErrConfigNotFound,
		}
	}

	s.logger.InfoContext(s.ctx, "configuration updated", slog.String("label", label))
	return nil
}

// Delete removes an instance configuration by label
func (s *storeService) Delete(label string) error {
	if label == "" {
		return &StoreError{
			Op:      "delete",
			Message: "label cannot be empty",
			Err:     ErrInvalidLabel,
		}
	}

	s.logger.DebugContext(s.ctx, "deleting configuration", slog.String("label", label))

	// Delete from database
	deleteSQL := `DELETE FROM instance_configs WHERE label = ?`
	result, err := s.db.ExecContext(s.ctx, deleteSQL, label)
	if err != nil {
		return &StoreError{
			Op:      "delete",
			Label:   label,
			Message: "failed to delete configuration",
			Err:     err,
		}
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &StoreError{
			Op:      "delete",
			Label:   label,
			Message: "failed to check delete result",
			Err:     err,
		}
	}

	if rowsAffected == 0 {
		return &StoreError{
			Op:      "delete",
			Label:   label,
			Message: "configuration not found",
			Err:     ErrConfigNotFound,
		}
	}

	s.logger.InfoContext(s.ctx, "configuration deleted", slog.String("label", label))
	return nil
}

// List returns all stored configuration labels
func (s *storeService) List() ([]string, error) {
	s.logger.DebugContext(s.ctx, "listing configurations")

	// Query database
	selectSQL := `SELECT label FROM instance_configs ORDER BY label`
	rows, err := s.db.QueryContext(s.ctx, selectSQL)
	if err != nil {
		return nil, &StoreError{
			Op:      "list",
			Message: "failed to query configurations",
			Err:     err,
		}
	}
	defer rows.Close()

	var labels []string
	for rows.Next() {
		var label string
		if err := rows.Scan(&label); err != nil {
			return nil, &StoreError{
				Op:      "list",
				Message: "failed to scan label",
				Err:     err,
			}
		}
		labels = append(labels, label)
	}

	if err := rows.Err(); err != nil {
		return nil, &StoreError{
			Op:      "list",
			Message: "error iterating rows",
			Err:     err,
		}
	}

	s.logger.DebugContext(s.ctx, "configurations listed", slog.Int("count", len(labels)))
	return labels, nil
}
