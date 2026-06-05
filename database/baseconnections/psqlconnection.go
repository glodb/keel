package baseconnections

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
)

// Keeping it open for multiple db or own db connections in microservices
type PsqlConnection struct {
	db     *sql.DB
	params *ConnectionParams
}

func (u *PsqlConnection) CreateConnection() (ConnectionInterface, error) {
	var host, port, username, password, dbName string
	if u.params != nil {
		host = u.params.Host
		port = u.params.Port
		username = u.params.Username
		password = u.params.Password
		dbName = u.params.DBName
	} else {
		cfg := configmanager.GetInstance().PSql
		host = cfg.Host
		port = cfg.Port
		username = cfg.Username
		password = cfg.Password
		dbName = cfg.DBName
	}

	dsn := "postgres://" + username + ":" + password + "@" + host + ":" + port + "?sslmode=disable"

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	defer db.Close()
	checkDBQuery := fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname='%s'", dbName)
	var exists int
	err = db.QueryRow(checkDBQuery).Scan(&exists)

	if err != nil && err != sql.ErrNoRows {
		logger.Log().Error("Error checking database", logger.ErrorField("error", err))
		return nil, err
	}

	// If the database doesn't exist, create it
	if exists != 1 {
		createDBQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)
		_, err := db.Exec(createDBQuery)
		if err != nil {
			logger.Log().Error("Error creating database", logger.ErrorField("error", err))
			return nil, err
		}
		logger.Log().Info("Database created successfully", logger.StringField("database_name", dbName))
	} else {
		logger.Log().Info("Database already exists", logger.StringField("database_name", dbName))
	}

	dsn = "postgres://" + username + ":" + password + "@" + host + ":" + port + "/" + dbName + "?sslmode=disable"

	newdb, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	u.db = newdb
	return u, nil
}

func (u *PsqlConnection) GetDB(dbType basetypes.DbType) interface{} {
	return u.db
}

func (u *PsqlConnection) Close() error {
	if u.db != nil {
		return u.db.Close()
	}
	return nil
}

func (u *PsqlConnection) Ping(ctx context.Context) error {
	if u.db != nil {
		return u.db.PingContext(ctx)
	}
	return errors.New("database is nil")
}

func (u *PsqlConnection) GetConnectionInfo() ConnectionInfo {
	var host, port, dbname string
	if u.params != nil {
		host = u.params.Host
		port = u.params.Port
		dbname = u.params.DBName
	} else {
		cfg := configmanager.GetInstance().PSql
		host = cfg.Host
		port = cfg.Port
		dbname = cfg.DBName
	}
	return ConnectionInfo{
		DatabaseType: basetypes.PSQL,
		Host:         host,
		Port:         port,
		DatabaseName: dbname,
		ConnectionState: func() string {
			if u.db != nil {
				return "connected"
			}
			return "disconnected"
		}(),
	}
}

func (u *PsqlConnection) IsHealthy(ctx context.Context) (bool, error) {
	if u.db == nil {
		return false, errors.New("database is nil")
	}

	err := u.db.PingContext(ctx)
	return err == nil, err
}
