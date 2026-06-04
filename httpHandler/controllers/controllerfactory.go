package controllers

import (
	"errors"
	"sync"

	"github.com/glodb/keel/database/basefunctions"

	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/httpHandler/controllers/baseinterfaces"

	"github.com/glodb/keel/settings/openapi"
)

var getInstance = sync.OnceValue(func() *controllersObject {
	instance := &controllersObject{}
	instance.controllers = make(map[basetypes.DbType]map[string]baseinterfaces.Controller)
	instance.controllers[basetypes.MONGO] = make(map[string]baseinterfaces.Controller)
	instance.controllers[basetypes.MYSQL] = make(map[string]baseinterfaces.Controller)
	instance.controllers[basetypes.PSQL] = make(map[string]baseinterfaces.Controller)
	instance.mutex = sync.Mutex{}
	instance.helper = openapi.NewControllerHelper()

	return instance
})

func Controllers() *controllersObject {
	return getInstance()
}

func Register(dbType basetypes.DbType, name string, factory func() baseinterfaces.Controller) {
	Controllers().Register(dbType, name, factory())
}

// Controllers struct
type controllersObject struct {
	controllers map[basetypes.DbType]map[string]baseinterfaces.Controller
	mutex       sync.Mutex
	helper      *openapi.ControllerHelper
}

// GetController createController is a factory to return the appropriate controller
func (c *controllersObject) GetController(dbType basetypes.DbType, controllerType string) (baseinterfaces.Controller, error) {
	if _, ok := c.controllers[dbType][controllerType]; ok {
		return c.controllers[dbType][controllerType], nil
	} else {
		return nil, errors.New("controller not found")
	}
}
func (c *controllersObject) Register(dbType basetypes.DbType, controllerName string, object baseinterfaces.Controller) {
	if _, ok := c.controllers[dbType][controllerName]; ok {
		return
	}
	c.controllers[dbType][controllerName] = object
	funcs, _ := basefunctions.BaseFunctions().GetFunctions(dbType, object.GetDBName())
	object.SetDependencies(*funcs)
	object.Initialize()

	if object.GetApisMap() != nil {
		object.RegisterApis(object.GetApisMap())
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.controllers[dbType][controllerName] = object
}
