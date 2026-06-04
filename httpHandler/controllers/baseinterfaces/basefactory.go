package baseinterfaces

import "github.com/glodb/keel/database/basetypes"

type BaseControllerFactory interface {
	GetController(dbType basetypes.DbType, controllerType string) (Controller, error)
}
