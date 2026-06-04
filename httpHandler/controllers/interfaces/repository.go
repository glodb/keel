package interfaces

import (
	"context"

	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/settings/errors"
)

// Repository defines a generic repository interface that abstracts database operations
// This reduces tight coupling between controllers and specific database implementations
type Repository interface {
	// Basic CRUD operations
	Create(ctx context.Context, data interface{}) error
	FindOne(ctx context.Context, filter interface{}, result interface{}) error
	FindMany(ctx context.Context, filter interface{}, options FindOptions) ([]interface{}, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}) error
	UpdateMany(ctx context.Context, filter interface{}, update interface{}) (int64, error)
	DeleteOne(ctx context.Context, filter interface{}) error
	DeleteMany(ctx context.Context, filter interface{}) (int64, error)

	// Advanced operations
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, result interface{}) error
	Upsert(ctx context.Context, filter interface{}, update interface{}) error
	Count(ctx context.Context, filter interface{}) (int64, error)
	Aggregate(ctx context.Context, pipeline interface{}) ([]interface{}, error)

	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error

	// Health and metadata
	Ping(ctx context.Context) error
	GetCollectionInfo() CollectionInfo
}

// FindOptions provides options for find operations
type FindOptions struct {
	Limit      int                    `json:"limit,omitempty"`
	Skip       int                    `json:"skip,omitempty"`
	Sort       map[string]int         `json:"sort,omitempty"`
	Projection map[string]interface{} `json:"projection,omitempty"`
}

// CollectionInfo provides metadata about the collection/table
type CollectionInfo struct {
	Name          string           `json:"name"`
	DatabaseType  basetypes.DbType `json:"database_type"`
	IndexCount    int              `json:"index_count"`
	DocumentCount int64            `json:"document_count"`
}

// RepositoryFactory creates repositories for different database types
type RepositoryFactory interface {
	CreateRepository(ctx context.Context, dbType basetypes.DbType, collectionName basetypes.CollectionName) (Repository, error)
	GetSupportedTypes() []basetypes.DbType
}

// ControllerInterface defines a cleaner controller interface with reduced coupling
type ControllerInterface interface {
	// Lifecycle methods
	Initialize(ctx context.Context, repo Repository) error
	Shutdown(ctx context.Context) error

	// Health and validation
	HealthCheck(ctx context.Context) error
	ValidateConfiguration() error

	// API registration
	RegisterAPIs() error
	RegisterDocumentation() error

	// Metadata
	GetControllerInfo() ControllerInfo
}

// ControllerInfo provides metadata about the controller
type ControllerInfo struct {
	Name             string                 `json:"name"`
	Version          string                 `json:"version"`
	SupportedDBTypes []basetypes.DbType     `json:"supported_db_types"`
	APIEndpoints     []APIEndpoint          `json:"api_endpoints"`
	Dependencies     []string               `json:"dependencies"`
	Configuration    map[string]interface{} `json:"configuration"`
}

// APIEndpoint represents an API endpoint
type APIEndpoint struct {
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	Description string   `json:"description"`
	Protected   bool     `json:"protected"`
	Deprecated  bool     `json:"deprecated"`
	Version     string   `json:"version"`
	Tags        []string `json:"tags"`
}

// BaseController provides common functionality for all controllers
type BaseController struct {
	repository     Repository
	logger         interface{} // Will be properly typed based on your logger
	metrics        interface{} // Will be properly typed based on your metrics
	config         interface{} // Will be properly typed based on your config
	initialized    bool
	controllerInfo ControllerInfo
}

// Initialize sets up the base controller with its dependencies
func (bc *BaseController) Initialize(ctx context.Context, repo Repository) error {
	bc.repository = repo
	bc.initialized = true
	return nil
}

// GetRepository returns the repository instance
func (bc *BaseController) GetRepository() Repository {
	return bc.repository
}

// IsInitialized returns whether the controller has been initialized
func (bc *BaseController) IsInitialized() bool {
	return bc.initialized
}

// Shutdown performs cleanup
func (bc *BaseController) Shutdown(ctx context.Context) error {
	bc.initialized = false
	return nil
}

// ValidateConfiguration validates the controller configuration
func (bc *BaseController) ValidateConfiguration() error {
	if !bc.initialized {
		return errors.NewValidationError("controller", "controller not initialized", nil)
	}
	return nil
}

// GetControllerInfo returns controller metadata
func (bc *BaseController) GetControllerInfo() ControllerInfo {
	return bc.controllerInfo
}

// HealthCheck performs a health check
func (bc *BaseController) HealthCheck(ctx context.Context) error {
	if !bc.initialized {
		return errors.NewValidationError("controller", "controller not initialized", nil)
	}

	if bc.repository != nil {
		return bc.repository.Ping(ctx)
	}

	return nil
}
