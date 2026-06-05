package basefunctions

import (
	"errors"
	"sync"

	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/httpHandler/controllers/baseinterfaces"
)

type baseFunctions struct {
	dbfunctions map[basetypes.DbType]*baseinterfaces.BaseFunctionsInterface
}

var getBaseFunctions = sync.OnceValue(func() *baseFunctions {
	instance := &baseFunctions{}
	instance.dbfunctions = make(map[basetypes.DbType]*baseinterfaces.BaseFunctionsInterface)
	return instance
})

func BaseFunctions() *baseFunctions {
	return getBaseFunctions()
}

func (u *baseFunctions) GetFunctions(dbType basetypes.DbType, dbName basetypes.DBName) (*baseinterfaces.BaseFunctionsInterface, error) {
	if connection, ok := u.dbfunctions[dbType]; ok {
		return connection, nil
	}
	switch dbType {
	case basetypes.MYSQL:
		{
			connection := MySqlFunctions{}
			functionsInterface := connection.GetFunctions()

			u.dbfunctions[dbType] = &functionsInterface
			return u.dbfunctions[dbType], nil
		}
	case basetypes.PSQL:
		{ //Adding this because ken wants to use framework for IOT
			connection := PSqlFunctions{}
			functionsInterface := connection.GetFunctions()

			u.dbfunctions[dbType] = &functionsInterface
			return u.dbfunctions[dbType], nil
		}
	case basetypes.MONGO:
		{ //Adding this because ken wants to use framework for IOT
			connection := MongoDBFunctions{}
			functionsInterface := connection.GetFunctions()

			u.dbfunctions[dbType] = &functionsInterface
			return u.dbfunctions[dbType], nil
		}
	}

	return nil, errors.New("not configured for this db")
}
