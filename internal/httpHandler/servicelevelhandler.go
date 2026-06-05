package httpHandler

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/internal/openapi"

	"github.com/gin-gonic/gin"
)

func (u *GINServer) HandleServiceLevelHandlers() {
	// Add health check endpoint
	u.engine.GET(fmt.Sprintf("/%s/health", configmanager.GetInstance().ServiceLBName), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   configmanager.GetInstance().ClassName,
			"version":   os.Getenv("OTEL_SERVICE_VERSION"),
		})
	})

	if configmanager.GetInstance().DeploymentEnv != "PROD" {
		// OpenAPI Swagger UI routes - only for current service
		swaggerHandler := openapi.NewSwaggerUIHandler()
		u.engine.GET(fmt.Sprintf("/%s/swagger", configmanager.GetInstance().ServiceLBName), swaggerHandler.ServeSwaggerUI())
		u.engine.GET(fmt.Sprintf("/%s/swagger/openapi.json", configmanager.GetInstance().ServiceLBName), swaggerHandler.ServeOpenAPISpec())
	}

}
