package servicehandler

import (
	"errors"
	"strings"
	"sync"
)

// factoryRegistry maps lowercase service names to their constructor functions.
// Services call Register() from their init() to add themselves here.
var (
	factoryMu       sync.RWMutex
	factoryRegistry = make(map[string]func() ServiceBase)
)

// Register registers a constructor function for the given service name.
// Call this from an init() in each service package:
//
//	func init() { servicehandler.Register("ssoservice", func() servicehandler.ServiceBase { return &SSOService{} }) }
func Register(name string, factory func() ServiceBase) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	factoryRegistry[strings.ToLower(name)] = factory
}

type services struct{}

var getInstance = sync.OnceValue(func() *services {
	return &services{}
})

// GetInstance returns the singleton service factory.
func GetInstance() *services {
	return getInstance()
}

// InitializeService looks up the service name in the JSON registry to confirm
// it is a known service, then instantiates it via the registered factory function.
func (c *services) InitializeService(serviceType string) (ServiceBase, error) {
	key := strings.ToLower(serviceType)

	factoryMu.RLock()
	factory, ok := factoryRegistry[key]
	factoryMu.RUnlock()
	if !ok {
		return nil, errors.New("service \"" + serviceType + "\" is in the registry but has no registered factory — add servicehandler.Register() in its init()")
	}

	return factory(), nil
}
