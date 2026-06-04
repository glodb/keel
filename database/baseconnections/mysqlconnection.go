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
	db *sql.DB
}

func (u *MysqlConnection) CreateConnection() (ConnectionInterface, error) {
	dsn := configmanager.GetInstance().MySql.Username + ":" + configmanager.GetInstance().MySql.Password + "@tcp(" + configmanager.GetInstance().MySql.Host + ":" + configmanager.GetInstance().MySql.Port + ")/" + configmanager.GetInstance().MySql.DBName
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
	config := configmanager.GetInstance().MySql
	return ConnectionInfo{
		DatabaseType: basetypes.MYSQL,
		Host:         config.Host,
		Port:         config.Port,
		DatabaseName: config.DBName,
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
