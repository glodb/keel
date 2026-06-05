package httpHandler

import (
	"net/http"
	"sync"

	"github.com/glodb/keel/httpHandler/baserouter"
	"github.com/glodb/keel/middlewares/basemiddlewares"
	"github.com/glodb/keel/settings/configmanager"

	"github.com/gin-gonic/gin"
)

// GetRouterRegistry returns this server's router registry.
// Pass it to controllers.InitializeControllers() so controllers
// register their routes on this server, not on a global singleton.
func (u *GINServer) GetRouterRegistry() *baserouter.BaseRouterHandler {
	return u.router
}

// GINServer struct represents a server instance using the Gin framework.
// Each instance owns its own router registry so multiple services can run
// in a single binary without sharing route state.
type GINServer struct {
	engine *gin.Engine
	router *baserouter.BaseRouterHandler
}

var getInstance = sync.OnceValue(func() *GINServer {
	return newGINServer()
})

func newGINServer() *GINServer {
	if configmanager.GetInstance().IsProduction || configmanager.GetInstance().Production {
		gin.SetMode(gin.ReleaseMode)
	}
	return &GINServer{engine: gin.New()}
}

// NewGINServer creates a fresh GINServer for use in tests or multi-service binaries.
func NewGINServer() *GINServer {
	return newGINServer()
}

func Server() *GINServer {
	return getInstance()
}

// GetEngine returns the Gin engine instance.
func (u *GINServer) GetEngine() *gin.Engine {
	return u.engine // Return the Gin engine
}

// HandleBlank handles the root route and returns a JSON response.
func (u *GINServer) HandleBlank() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{"msg": "Welcome to keel default route"}) // Return a welcome message
	}
}

// Setup initialises the Gin engine and builds one router group ("tier") per
// entry in middlewares.  Each tier is a distinct middleware stack; controllers
// declare which tier their routes belong to via GetApisMap().
//
// # Naming tiers
//
// The string keys of the middlewares map become the tier names that controllers
// reference with customtypes.RouterNames.  Choose names that reflect the
// authentication / rate-limit contract, not URL structure.  Convention:
//
//	"open"  — public, no auth (rate-limit middleware only)
//	"auth"  — requires a valid JWT / session token
//	"base"  — internal health, metrics, or admin endpoints
//
// A controller that returns an unrecognised tier in GetApisMap() will have
// those routes skipped and an error logged at startup.
//
// # URL assembly
//
// Routes are ultimately registered as:
//
//	config.ServiceLBName + config.ApiPrefix + "/" + Apis.ApiName
//
// The tier group is an all-paths group ("/"), so the full URL lives entirely
// in the three config fields above plus each Apis.ApiName.
func (u *GINServer) Setup(middlewares map[string][]basemiddlewares.Middleware) {
	u.router = baserouter.New()

	u.HandleServiceLevelHandlers()

	u.engine.GET("/", u.HandleBlank())

	for tier, tierMiddlewares := range middlewares {
		group := u.engine.Group("/")
		for _, m := range tierMiddlewares {
			group.Use(m.GetHandlerFunc())
		}
		u.router.SetRouter(tier, group)
	}

	// Publish this server's router so apiregistration.RegisterApis() can resolve
	// tier groups without reaching for a global singleton.
	baserouter.SetActive(u.router)
}
