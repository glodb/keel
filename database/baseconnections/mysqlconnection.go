package baseconnections

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql"

	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/settings/configmanager"
)

// Keeping it open for multiple db or own db connections in microservices
type MysqlConnection struct {
	db     *sql.DB
	params *ConnectionParams
}

func (u *MysqlConnection) CreateConnection() (ConnectionInterface, error) {
	var host, port, username, password, dbname string
	if u.params != nil {
		host = u.params.Host
		port = u.params.Port
		username = u.params.Username
		password = u.params.Password
		dbname = u.params.DBName
	} else {
		cfg := configmanager.GetInstance().MySql
		host = cfg.Host
		port = cfg.Port
		username = cfg.Username
		password = cfg.Password
		dbname = cfg.DBName
	}

	dsn := username + ":" + password + "@tcp(" + host + ":" + port + ")/" + dbname
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	u.db = db
	return u, nil
}

func (u *MysqlConnection) GetDB(dbType basetypes.DbType) interface{} {
	return u.db
}

func (u *MysqlConnection) Close() error {
	if u.db != nil {
		return u.db.Close()
	}
	return nil
}

func (u *MysqlConnection) Ping(ctx context.Context) error {
	if u.db != nil {
		return u.db.PingContext(ctx)
	}
	return errors.New("database is nil")
}

func (u *MysqlConnection) GetConnectionInfo() ConnectionInfo {
	var host, port, dbname string
	if u.params != nil {
		host = u.params.Host
		port = u.params.Port
		dbname = u.params.DBName
	} else {
		cfg := configmanager.GetInstance().MySql
		host = cfg.Host
		port = cfg.Port
		dbname = cfg.DBName
	}
	return ConnectionInfo{
		DatabaseType: basetypes.MYSQL,
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

func (u *MysqlConnection) IsHealthy(ctx context.Context) (bool, error) {
	if u.db == nil {
		return false, errors.New("database is nil")
	}

	err := u.db.PingContext(ctx)
	return err == nil, err
}
