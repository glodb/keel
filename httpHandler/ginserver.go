package httpHandler

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/glodb/keel/httpHandler/baserouter"
	"github.com/glodb/keel/middlewares/basemiddlewares"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/tracing"

	"github.com/gin-gonic/gin"
)

// GINServer struct represents a server instance using the Gin framework.
type GINServer struct {
	engine *gin.Engine // Gin HTTP engine
}

var getInstance = sync.OnceValue(func() *GINServer {
	instance := &GINServer{} // Initialize instance once

	// Always use custom engine with our panic recovery
	instance.engine = gin.New() // Create a new Gin engine

	// Configure Gin mode based on environment
	if configmanager.GetInstance().IsProduction || configmanager.GetInstance().Production {
		gin.SetMode(gin.ReleaseMode)
	}

	return instance
})

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
		c.JSON(http.StatusOK, map[string]interface{}{"msg": "Welcome to gpsina gin api"}) // Return a welcome message
	}
}

func (u *GINServer) Setup(middlewares map[string][]basemiddlewares.Middleware) {

	u.HandleServiceLevelHandlers()

	u.engine.GET("/", u.HandleBlank())

	for tier, tierMiddlewares := range middlewares {
		group := u.engine.Group("/")
		for _, m := range tierMiddlewares {
			group.Use(m.GetHandlerFunc())
		}
		baserouter.GetInstance().SetRouter(tier, group)
	}
	// // Logger middleware setup
	// loggerMiddleware := commonmiddlewares.LoggingMiddleware{}
	// u.engine.Use(loggerMiddleware.GetHandlerFunc())

	// // Rate limiting middleware setup
	// rateLimitMiddleware := commonmiddlewares.RateLimitMiddleware{}
	// u.engine.Use(rateLimitMiddleware.GetHandlerFunc())

	// // CORS middleware setup
	// corsMiddleware := commonmiddlewares.CORSMiddleware{}
	// u.engine.Use(corsMiddleware.GetHandlerFunc())

	// // Define handler for root route
	// u.engine.GET("/", u.HandleBlank())

	// // API middleware setup
	// apiMiddleware := commonmiddlewares.ApiMiddleware{}
	// u.engine.Use(apiMiddleware.GetHandlerFunc())

	// // Group for base routes
	// baseginrouter := u.engine.Group("/")
	// baserouter.GetInstance().SetRouter(string(customtypes.Base), baseginrouter)

	// // Session middleware setup for open routes
	// baseopenrouter := baseginrouter.Group("/")

	// sessionMiddleware := commonmiddlewares.SessionMiddleware{}
	// baseopenrouter.Use(sessionMiddleware.GetHandlerFunc())

	// // Check allowed version of the app
	// versionMiddleware := commonmiddlewares.VersionMiddleware{}
	// baseopenrouter.Use(versionMiddleware.GetHandlerFunc())

	// baserouter.GetInstance().SetRouter(string(customtypes.Open), baseopenrouter)

	// baseauthwriter := baseopenrouter.Group("/")

	// userMiddleware := commonmiddlewares.UserMiddleware{}
	// baseauthwriter.Use(userMiddleware.GetHandlerFunc())

	// // Access middleware setup for open routes
	// accessMiddleware := commonmiddlewares.AccessMiddleware{}
	// baseauthwriter.Use(accessMiddleware.GetHandlerFunc())

	// authMiddleware := commonmiddlewares.AuthMiddleware{}
	// baseauthwriter.Use(authMiddleware.GetHandlerFunc())

	// validationMiddleware := commonmiddlewares.ValidationMiddleware{}
	// baseauthwriter.Use(validationMiddleware.GetHandlerFunc())

	// rolesMiddleware := commonmiddlewares.RolesMiddleware{}
	// baseauthwriter.Use(rolesMiddleware.GetHandlerFunc())

	// baserouter.GetInstance().SetRouter(string(customtypes.Auth), baseauthwriter)

	// basebusinesswriter := baseauthwriter.Group("/")

	// onboardingMiddleware := businessmiddlewares.OnboardingMiddleware{}
	// basebusinesswriter.Use(onboardingMiddleware.GetHandlerFunc())

	// baserouter.GetInstance().SetRouter(string(customtypes.Business), basebusinesswriter)
}

// Start starts the Gin server with graceful shutdown
func (u *GINServer) Start() error {
	// Initialize observability
	tracer := tracing.GetInstance()
	defer func() {
		if err := tracer.Shutdown(context.Background()); err != nil {
			logger.Log().Error("Failed to shutdown tracer", logger.ErrorField("err", err))
		}
	}()

	// Start main server
	go func() {
		if err := u.engine.Run(":8080"); err != nil && err != http.ErrServerClosed {
			logger.Log().Fatal("Failed to start server", logger.ErrorField("err", err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log().Info("Shutting down server...")

	logger.Log().Debug("Server exited")
	return nil
}

