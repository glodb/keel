package baseconnections

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/settings/errors"
	"github.com/glodb/keel/settings/logger"
)

// ConnectionPool manages database connections with proper lifecycle management
type ConnectionPool struct {
	connections      map[basetypes.DbType]ConnectionInterface
	activeQueries    map[basetypes.DbType]*int64 // Track number of active queries per connection (atomic)
	connectionParams map[basetypes.DbType]ConnectionParams
	mutex            sync.RWMutex
	logger           *logger.Logger
	healthCheck      *time.Ticker
	stopChan         chan struct{}
}

var (
	connectionPoolInstance *ConnectionPool
	connectionPoolOnce     sync.Once
)

// GetConnectionPool returns the singleton connection pool instance
func GetConnectionPool() *ConnectionPool {
	connectionPoolOnce.Do(func() {
		connectionPoolInstance = &ConnectionPool{
			connections:      make(map[basetypes.DbType]ConnectionInterface),
			activeQueries:    make(map[basetypes.DbType]*int64),
			connectionParams: make(map[basetypes.DbType]ConnectionParams),
			logger:           logger.Log(),
			stopChan:         make(chan struct{}),
		}
		connectionPoolInstance.startHealthChecker()
	})
	return connectionPoolInstance
}

func NewConnectionPool() *ConnectionPool {
	instance := &ConnectionPool{
		connections:      make(map[basetypes.DbType]ConnectionInterface),
		activeQueries:    make(map[basetypes.DbType]*int64),
		connectionParams: make(map[basetypes.DbType]ConnectionParams),
		logger:           logger.Log(),
		stopChan:         make(chan struct{}),
	}
	instance.startHealthChecker()

	return instance
}

// RegisterConnection stores connection parameters for the given database type so they
// are used instead of (or as an override of) the configmanager / env-var values the
// next time a connection of that type is dialed. It returns the pool itself so calls
// can be chained. Must be called before the first GetConnection for that dbType.
func (cp *ConnectionPool) RegisterConnection(dbType basetypes.DbType, params ConnectionParams) *ConnectionPool {
	cp.mutex.Lock()
	cp.connectionParams[dbType] = params
	cp.mutex.Unlock()
	return cp
}

// GetConnection retrieves or creates a database connection for the specified type
// Returns proper error handling instead of nil
func (cp *ConnectionPool) GetConnection(ctx context.Context, dbType basetypes.DbType) (ConnectionInterface, error) {
	cp.mutex.RLock()
	if connection, exists := cp.connections[dbType]; exists {
		cp.mutex.RUnlock()

		var err error
		// Verify connection is still healthy
		if healthy, err := connection.IsHealthy(ctx); healthy && err == nil {
			return connection, nil
		}

		// Connection is unhealthy, remove it and create a new one
		cp.logger.Warn("Unhealthy connection detected, recreating",
			logger.StringField("database_type", dbType.String()),
			logger.ErrorField("health_check_error", err))

		cp.mutex.Lock()
		delete(cp.connections, dbType)
		delete(cp.activeQueries, dbType)
		cp.mutex.Unlock()
	} else {
		cp.mutex.RUnlock()
	}

	// Create new connection
	return cp.createConnection(ctx, dbType)
}

// createConnection creates a new database connection with proper error handling
func (cp *ConnectionPool) createConnection(ctx context.Context, dbType basetypes.DbType) (ConnectionInterface, error) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	// Double-check pattern to avoid race conditions
	if connection, exists := cp.connections[dbType]; exists {
		return connection, nil
	}

	// Retrieve registered params (if any) while still holding the write lock.
	registeredParams, hasParams := cp.connectionParams[dbType]

	var connection ConnectionInterface
	var err error

	switch dbType {
	case basetypes.MYSQL:
		mysqlConn := &MysqlConnection{}
		if hasParams {
			mysqlConn.params = &registeredParams
		}
		connection, err = mysqlConn.CreateConnection()
		if err != nil {
			return nil, errors.NewDatabaseError("mysql connection creation", err).
				WithMetadata("database_type", "mysql")
		}

	case basetypes.PSQL:
		psqlConn := &PsqlConnection{}
		if hasParams {
			psqlConn.params = &registeredParams
		}
		connection, err = psqlConn.CreateConnection()
		if err != nil {
			return nil, errors.NewDatabaseError("postgresql connection creation", err).
				WithMetadata("database_type", "postgresql")
		}

	case basetypes.MONGO:
		mongoConn := &MongoConnection{}
		if hasParams {
			mongoConn.params = &registeredParams
		}
		connection, err = mongoConn.CreateConnection()
		if err != nil {
			return nil, errors.NewDatabaseError("mongodb connection creation", err).
				WithMetadata("database_type", "mongodb")
		}

	default:
		return nil, errors.NewAppError(
			errors.ErrCodeConfigurationError,
			fmt.Sprintf("Unsupported database type: %v", dbType),
			nil,
		).WithMetadata("database_type", dbType)
	}

	// Verify the connection before storing it
	if err := connection.Ping(ctx); err != nil {
		// Attempt to close the connection if ping fails
		if closeErr := connection.Close(); closeErr != nil {
			cp.logger.Error("Failed to close failed connection",
				logger.StringField("database_type", dbType.String()),
				logger.ErrorField("close_error", closeErr))
		}
		return nil, errors.NewDatabaseError("connection ping verification", err).
			WithMetadata("database_type", dbType.String())
	}

	// Store the healthy connection and initialize query counter
	cp.connections[dbType] = connection
	counter := int64(0)
	cp.activeQueries[dbType] = &counter

	cp.logger.Info("Database connection created successfully",
		logger.StringField("database_type", dbType.String()),
		logger.AnyField("connection_info", connection.GetConnectionInfo()))

	return connection, nil
}

// GetAllConnections returns information about all active connections
func (cp *ConnectionPool) GetAllConnections() map[basetypes.DbType]ConnectionInfo {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	result := make(map[basetypes.DbType]ConnectionInfo)
	for dbType, connection := range cp.connections {
		result[dbType] = connection.GetConnectionInfo()
	}
	return result
}

// CloseConnection closes a specific database connection
func (cp *ConnectionPool) CloseConnection(dbType basetypes.DbType) error {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	if connection, exists := cp.connections[dbType]; exists {
		delete(cp.connections, dbType)
		delete(cp.activeQueries, dbType)
		if err := connection.Close(); err != nil {
			cp.logger.Error("Failed to close database connection",
				logger.StringField("database_type", dbType.String()),
				logger.ErrorField("error", err))
			return errors.NewDatabaseError("connection close", err).
				WithMetadata("database_type", dbType.String())
		}
		cp.logger.Info("Database connection closed",
			logger.StringField("database_type", dbType.String()))
	}
	return nil
}

// CloseAllConnections gracefully closes all database connections
func (cp *ConnectionPool) CloseAllConnections() error {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	// Stop health checker
	close(cp.stopChan)
	if cp.healthCheck != nil {
		cp.healthCheck.Stop()
	}

	var lastError error
	for dbType, connection := range cp.connections {
		if err := connection.Close(); err != nil {
			cp.logger.Error("Failed to close database connection",
				logger.StringField("database_type", dbType.String()),
				logger.ErrorField("error", err))
			lastError = err
		} else {
			cp.logger.Info("Database connection closed",
				logger.StringField("database_type", dbType.String()))
		}
	}

	// Clear the connections and counters maps
	cp.connections = make(map[basetypes.DbType]ConnectionInterface)
	cp.activeQueries = make(map[basetypes.DbType]*int64)

	if lastError != nil {
		return errors.NewDatabaseError("bulk connection close", lastError)
	}
	return nil
}

// startHealthChecker starts a background goroutine to check connection health
func (cp *ConnectionPool) startHealthChecker() {
	cp.healthCheck = time.NewTicker(30 * time.Second) // Check every 30 seconds

	go func() {
		for {
			select {
			case <-cp.healthCheck.C:
				cp.performHealthChecks()
			case <-cp.stopChan:
				return
			}
		}
	}()
}

// performHealthChecks checks the health of all connections
func (cp *ConnectionPool) performHealthChecks() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cp.mutex.RLock()
	connections := make(map[basetypes.DbType]ConnectionInterface)
	counters := make(map[basetypes.DbType]*int64)
	for dbType, conn := range cp.connections {
		connections[dbType] = conn
		if counter, exists := cp.activeQueries[dbType]; exists {
			counters[dbType] = counter
		}
	}
	cp.mutex.RUnlock()

	for dbType, connection := range connections {
		go func(dt basetypes.DbType, conn ConnectionInterface) {
			// Check if connection is currently in use
			if counter, exists := counters[dt]; exists && counter != nil {
				activeCount := atomic.LoadInt64(counter)
				if activeCount > 0 {
					cp.logger.Debug("Skipping health check - connection in use",
						logger.StringField("database_type", dt.String()),
						logger.Int64Field("active_queries", activeCount))
					return
				}
			}

			healthy, err := conn.IsHealthy(ctx)
			if !healthy || err != nil {
				// Double-check connection is not in use before closing
				if counter, exists := counters[dt]; exists && counter != nil {
					activeCount := atomic.LoadInt64(counter)
					if activeCount > 0 {
						cp.logger.Debug("Skipping close - connection became active",
							logger.StringField("database_type", dt.String()),
							logger.Int64Field("active_queries", activeCount))
						return
					}
				}

				// Remove the unhealthy connection
				cp.mutex.Lock()
				delete(cp.connections, dt)
				delete(cp.activeQueries, dt)
				cp.mutex.Unlock()

				// Attempt to close the connection
				if closeErr := conn.Close(); closeErr != nil {
					cp.logger.Error("Failed to close unhealthy connection",
						logger.StringField("database_type", dt.String()),
						logger.ErrorField("close_error", closeErr))
				}
			}
		}(dbType, connection)
	}
}

// GetConnectionCount returns the number of active connections
func (cp *ConnectionPool) GetConnectionCount() int {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()
	return len(cp.connections)
}

// IncrementActiveQueries increments the active query count for a connection type
func (cp *ConnectionPool) IncrementActiveQueries(dbType basetypes.DbType) {
	cp.mutex.RLock()
	counter, exists := cp.activeQueries[dbType]
	cp.mutex.RUnlock()

	if exists && counter != nil {
		atomic.AddInt64(counter, 1)
	}
}

// DecrementActiveQueries decrements the active query count for a connection type
func (cp *ConnectionPool) DecrementActiveQueries(dbType basetypes.DbType) {
	cp.mutex.RLock()
	counter, exists := cp.activeQueries[dbType]
	cp.mutex.RUnlock()

	if exists && counter != nil {
		atomic.AddInt64(counter, -1)
	}
}

// IsConnectionHealthy checks if a specific connection is healthy
func (cp *ConnectionPool) IsConnectionHealthy(ctx context.Context, dbType basetypes.DbType) (bool, error) {
	cp.mutex.RLock()
	connection, exists := cp.connections[dbType]
	cp.mutex.RUnlock()

	if !exists {
		return false, errors.NewAppError(
			errors.ErrCodeRecordNotFound,
			fmt.Sprintf("No connection found for database type: %v", dbType),
			nil,
		).WithMetadata("database_type", dbType.String())
	}

	return connection.IsHealthy(ctx)
}

// Legacy support - maintain backward compatibility
type dbConnections struct {
	pool *ConnectionPool
}

func DBConnection() *dbConnections {
	return &dbConnections{
		pool: GetConnectionPool(),
	}
}

// GetConnection provides backward compatibility with the old interface
// Deprecated: Use GetConnectionPool().GetConnection() instead
func (dc *dbConnections) GetConnection(dbType basetypes.DbType) ConnectionInterface {
	ctx := context.Background()
	connection, err := dc.pool.GetConnection(ctx, dbType)
	if err != nil {
		logger.Log().Error("Failed to get database connection",
			logger.StringField("database_type", dbType.String()),
			logger.ErrorField("error", err))
		return nil
	}

	return connection
}
