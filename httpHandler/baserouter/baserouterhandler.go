package baserouter

import (
	"errors"
	"sync"

	"github.com/glodb/keel/settings/customtypes"

	"github.com/gin-gonic/gin"
)

// BaseRouterHandler manages the tier router groups for a single GINServer instance.
// Each GINServer owns one; the active one is published via SetActive so that
// apiregistration can resolve routes without a global singleton.
type BaseRouterHandler struct {
	router map[string]*gin.RouterGroup
}

// New creates a fresh, empty BaseRouterHandler.
func New() *BaseRouterHandler {
	return &BaseRouterHandler{router: make(map[string]*gin.RouterGroup)}
}

// active holds the BaseRouterHandler that is currently being set up.
// Set by GINServer.Setup() before InitializeControllers() is called.
var (
	activeMu sync.RWMutex
	active   *BaseRouterHandler
)

// SetActive registers r as the router that apiregistration should use during
// the current controller initialization pass. Call this right after Setup().
func SetActive(r *BaseRouterHandler) {
	activeMu.Lock()
	active = r
	activeMu.Unlock()
}

// GetActive returns the currently active BaseRouterHandler.
func GetActive() *BaseRouterHandler {
	activeMu.RLock()
	defer activeMu.RUnlock()
	return active
}

// SetRouter registers a Gin router group under the given tier name.
func (u *BaseRouterHandler) SetRouter(name string, router *gin.RouterGroup) {
	u.router[name] = router
}

// GetRouter returns the router group for the given tier, or an error if not found.
func (u *BaseRouterHandler) GetRouter(name customtypes.RouterNames) (*gin.RouterGroup, error) {
	if r, ok := u.router[string(name)]; ok {
		return r, nil
	}
	return nil, errors.New("router not found")
}
