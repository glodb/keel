package baseconnections

import (
	"context"
	"time"

	"github.com/glodb/keel/database/basetypes"
)

// ConnectionInterface defines the contract for database connections
// All database implementations must satisfy this interface
type ConnectionInterface interface {
	// CreateConnection establishes a new database connection with proper configuration
	CreateConnection() (ConnectionInterface, error)

	// GetDB returns the underlying database connection for the specified database type
	GetDB(basetypes.DbType) interface{}

	// Ping verifies the database connection is alive and responsive
	Ping(ctx context.Context) error

	// Close gracefully closes the database connection and releases resources
	Close() error

	// GetConnectionInfo returns connection metadata for monitoring
	GetConnectionInfo() ConnectionInfo

	// IsHealthy performs a comprehensive health check on the connection
	IsHealthy(ctx context.Context) (bool, error)
}

// ConnectionParams holds the credentials and dial options for a single database connection.
// Pass to GetConnectionPool().RegisterConnection() to supply parameters programmatically
// from outside the library, instead of (or overriding) the JSON config / env vars.
// Fields that are left at their zero value are ignored and the configmanager value is
// used instead, so you can override only the fields you care about.
type ConnectionParams struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string

	// MongoDB-specific options
	Atlas               bool
	SecureMongo         bool
	CertFile            string
	MongoMaxConnections int
	AppName             string
}

// ConnectionInfo provides metadata about database connections for observability
type ConnectionInfo struct {
	DatabaseType     basetypes.DbType `json:"database_type"`
	Host             string           `json:"host"`
	Port             string           `json:"port"`
	DatabaseName     string           `json:"database_name"`
	MaxConnections   int              `json:"max_connections"`
	ActiveConnection int              `json:"active_connections"`
	IdleConnections  int              `json:"idle_connections"`
	ConnectionState  string           `json:"connection_state"`
	LastPingTime     time.Time        `json:"last_ping_time"`
	LastPingLatency  time.Duration    `json:"last_ping_latency"`
}
