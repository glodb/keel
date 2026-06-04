package controllers

import (
	"errors"
	"sync"

	"github.com/glodb/keel/database/basefunctions"

	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/httpHandler/controllers/baseinterfaces"

	"github.com/glodb/keel/settings/openapi"
)

type factoryEntry struct {
	dbType  basetypes.DbType
	factory func() baseinterfaces.Controller
}

var (
	pendingFactories []factoryEntry
	factoryMu        sync.Mutex
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

// Register stores the factory — initialization is deferred until InitializeControllers().
func Register(dbType basetypes.DbType, name string, factory func() baseinterfaces.Controller) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	pendingFactories = append(pendingFactories, factoryEntry{dbType: dbType, factory: factory})
}

// InitializeControllers initializes all registered controllers. Call after Boot().
func InitializeControllers() {
	factoryMu.Lock()
	factories := pendingFactories
	pendingFactories = nil
	factoryMu.Unlock()

	for _, e := range factories {
		Controllers().initialize(e.dbType, e.factory())
	}
}

// Controllers struct
type controllersObject struct {
	controllers map[basetypes.DbType]map[string]baseinterfaces.Controller
	mutex       sync.Mutex
	helper      *openapi.ControllerHelper
}

// GetController returns an already-initialized controller.
func (c *controllersObject) GetController(dbType basetypes.DbType, controllerType string) (baseinterfaces.Controller, error) {
	if _, ok := c.controllers[dbType][controllerType]; ok {
		return c.controllers[dbType][controllerType], nil
	}
	return nil, errors.New("controller not found")
}

func (c *controllersObject) initialize(dbType basetypes.DbType, object baseinterfaces.Controller) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	funcs, _ := basefunctions.BaseFunctions().GetFunctions(dbType, object.GetDBName())
	object.SetDependencies(*funcs)
	object.Initialize()

	if object.GetApisMap() != nil {
		object.RegisterApis(object.GetApisMap())
	}

	c.controllers[dbType][string(object.GetDBName())] = object
}
