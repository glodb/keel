package middlewareregistry

import (
	"sync"

	"github.com/glodb/keel/middlewares/basemiddlewares"
)

// MiddlewareMap holds middleware chains keyed by tier name (e.g. "open", "auth").
type MiddlewareMap struct {
	middlewares map[string][]basemiddlewares.Middleware
}

// Registry stores global and per-service middleware tiers.
type Registry struct {
	middlewares        MiddlewareMap
	serviceMiddlewares map[string]MiddlewareMap
}

// SetupFunc wires application-specific middleware into the registry.
type SetupFunc func(*Registry)

var (
	instance *Registry
	once     sync.Once
	setupFn  SetupFunc
	setupMu  sync.RWMutex
)

// SetSetup registers the function that populates middleware tiers.
// Call from your application's init(), before GetInstance() is first used.
func SetSetup(fn SetupFunc) {
	setupMu.Lock()
	setupFn = fn
	setupMu.Unlock()
}

// GetInstance returns the singleton middleware registry.
func GetInstance() *Registry {
	once.Do(func() {
		instance = &Registry{
			middlewares: MiddlewareMap{
				middlewares: make(map[string][]basemiddlewares.Middleware),
			},
			serviceMiddlewares: make(map[string]MiddlewareMap),
		}
		setupMu.RLock()
		fn := setupFn
		setupMu.RUnlock()
		if fn != nil {
			fn(instance)
		}
	})
	return instance
}

// SetTier replaces the middleware chain for a tier.
func (m *Registry) SetTier(tier string, middlewares []basemiddlewares.Middleware) {
	m.middlewares.middlewares[tier] = middlewares
}

// RegisterMiddleware appends a middleware to a tier.
func (m *Registry) RegisterMiddleware(tier string, middleware basemiddlewares.Middleware) {
	m.middlewares.middlewares[tier] = append(m.middlewares.middlewares[tier], middleware)
}

// GetMiddlewares returns middleware tiers for the given service, or the global default.
func (m *Registry) GetMiddlewares(serviceName string) map[string][]basemiddlewares.Middleware {
	if serviceName == "" {
		return m.middlewares.middlewares
	}

	if _, ok := m.serviceMiddlewares[serviceName]; ok {
		return m.serviceMiddlewares[serviceName].middlewares
	}

	return m.middlewares.middlewares
}
