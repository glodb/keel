package baserouter

import (
	"errors"
	"sync"

	"github.com/glodb/keel/settings/customtypes"

	"github.com/gin-gonic/gin"
)

// baseRouterHandler struct manages different router groups.
type baseRouterHandler struct {
	router map[string]*gin.RouterGroup // Map to store router groups by name
}

var getInstance = sync.OnceValue(func() *baseRouterHandler {
	instance := &baseRouterHandler{}
	instance.router = make(map[string]*gin.RouterGroup)
	return instance
})

// GetInstance returns a single object of baseRouterHandler (singleton pattern).
func GetInstance() *baseRouterHandler {
	return getInstance()
}

// SetRouter sets a router group with a given name.
func (u *baseRouterHandler) SetRouter(name string, router *gin.RouterGroup) {
	u.router[name] = router // Set the router group in the map with the given name
}

func (u *baseRouterHandler) GetRouter(name customtypes.RouterNames) (*gin.RouterGroup, error) {
	if _, ok := u.router[string(name)]; ok {
		return u.router[string(name)], nil
	}
	return nil, errors.New("router not found")
}
