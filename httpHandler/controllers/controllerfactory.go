package controllers

import (
	"errors"
	"sync"

	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/httpHandler/controllers/baseinterfaces"
	"github.com/glodb/keel/internal/database/basefunctions"
	"github.com/glodb/keel/settings/logger"
)

type factoryEntry struct {
	dbType  basetypes.DbType
	name    string
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
	return instance
})

func Controllers() *controllersObject {
	return getInstance()
}

func NewControllers() *controllersObject {
	instance := &controllersObject{}
	instance.controllers = make(map[basetypes.DbType]map[string]baseinterfaces.Controller)
	instance.controllers[basetypes.MONGO] = make(map[string]baseinterfaces.Controller)
	instance.controllers[basetypes.MYSQL] = make(map[string]baseinterfaces.Controller)
	instance.controllers[basetypes.PSQL] = make(map[string]baseinterfaces.Controller)
	instance.mutex = sync.Mutex{}
	return instance
}

// WithFunctions replaces the DB-function resolver used during controller
// initialization. The signature matches basefunctions.BaseFunctions().GetFunctions.
// Use in tests to inject mock implementations without a live database:
//
//	ctrl := controllers.NewControllers().WithFunctions(func(dt basetypes.DbType, dn basetypes.DBName) (*baseinterfaces.BaseFunctionsInterface, error) {
//	    var fi baseinterfaces.BaseFunctionsInterface = &MyMockFunctions{}
//	    return &fi, nil
//	})
func (c *controllersObject) WithFunctions(fn func(dbType basetypes.DbType, dbName basetypes.DBName) (*baseinterfaces.BaseFunctionsInterface, error)) *controllersObject {
	c.getFunctions = fn
	return c
}

// Register stores the factory — initialization is deferred until InitializeControllers().
func Register(dbType basetypes.DbType, name string, factory func() baseinterfaces.Controller) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	pendingFactories = append(pendingFactories, factoryEntry{dbType: dbType, name: name, factory: factory})
}

// InitializeControllersInto drains the pending factory queue and initializes
// every registered controller into target. Use this with NewControllers() in
// tests to get an isolated controller graph that does not touch the singleton:
//
//	ctrl := controllers.NewControllers().WithFunctions(mockFn)
//	controllers.InitializeControllersInto(ctrl)
//	// ctrl now holds your initialized controllers; pass it around explicitly.
func InitializeControllersInto(target *controllersObject) {
	factoryMu.Lock()
	factories := pendingFactories
	pendingFactories = nil
	factoryMu.Unlock()

	for _, e := range factories {
		target.initialize(e.dbType, e.name, e.factory())
	}
}

// InitializeControllers initializes all registered controllers into the
// singleton. Call after Boot().
func InitializeControllers() {
	InitializeControllersInto(Controllers())
}

type controllersObject struct {
	controllers  map[basetypes.DbType]map[string]baseinterfaces.Controller
	mutex        sync.Mutex
	// getFunctions, when non-nil, is used instead of basefunctions.BaseFunctions()
	// to resolve DB function implementations during controller initialization.
	// Inject a mock via NewControllers().WithFunctions(fn) in tests.
	getFunctions func(dbType basetypes.DbType, dbName basetypes.DBName) (*baseinterfaces.BaseFunctionsInterface, error)
}

// GetController returns an already-initialized controller.
func (c *controllersObject) GetController(dbType basetypes.DbType, controllerType string) (baseinterfaces.Controller, error) {
	if _, ok := c.controllers[dbType][controllerType]; ok {
		return c.controllers[dbType][controllerType], nil
	}
	return nil, errors.New("controller not found")
}

func (c *controllersObject) initialize(dbType basetypes.DbType, name string, object baseinterfaces.Controller) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var funcs *baseinterfaces.BaseFunctionsInterface
	if c.getFunctions != nil {
		funcs, _ = c.getFunctions(dbType, object.GetDBName())
	} else {
		funcs, _ = basefunctions.BaseFunctions().GetFunctions(dbType, object.GetDBName())
	}
	object.SetDependencies(*funcs, c)
	object.Initialize()

	if object.GetApisMap() != nil {
		if routeErrs := object.RegisterApis(object.GetApisMap()); len(routeErrs) > 0 {
			for _, re := range routeErrs {
				logger.Log().Error("Controller route registration error",
					logger.StringField("controller", string(object.GetDBName())),
					logger.ErrorField("error", re),
				)
			}
		}
	}

	c.controllers[dbType][string(object.GetDBName())] = object
	if name != "" {
		c.controllers[dbType][name] = object
	}
}
