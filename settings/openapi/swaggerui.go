package openapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"

	"github.com/gin-gonic/gin"
)

// SwaggerUIHandler handles serving Swagger UI for dynamic OpenAPI specifications
type SwaggerUIHandler struct {
	openAPIGenerator *OpenAPIGenerator
}

// NewSwaggerUIHandler creates a new SwaggerUIHandler instance
func NewSwaggerUIHandler() *SwaggerUIHandler {
	return &SwaggerUIHandler{
		openAPIGenerator: GetInstance(),
	}
}

// ServeSwaggerUI serves the Swagger UI for the current service
func (sh *SwaggerUIHandler) ServeSwaggerUI() gin.HandlerFunc {
	return func(c *gin.Context) {
		config := configmanager.GetInstance()
		serviceName := strings.ToUpper(config.ClassName)

		// Get OpenAPI spec for the current service
		specJSON, err := sh.openAPIGenerator.GetCurrentServiceSpecJSON()
		if err != nil {
			logger.Log().Error("Failed to get OpenAPI spec",
				logger.StringField("service", config.ClassName),
				logger.ErrorField("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate API spec"})
			return
		}

		// Create Swagger UI HTML with the spec
		html := sh.createSwaggerUIHTML(serviceName, string(specJSON))

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	}
}

// ServeOpenAPISpec serves the raw OpenAPI specification JSON for the current service
func (sh *SwaggerUIHandler) ServeOpenAPISpec() gin.HandlerFunc {
	return func(c *gin.Context) {
		specJSON, err := sh.openAPIGenerator.GetCurrentServiceSpecJSON()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate API spec"})
			return
		}

		c.Header("Content-Type", "application/json")
		c.Data(http.StatusOK, "application/json", specJSON)
	}
}

// createSwaggerUIHTML creates the HTML for Swagger UI with embedded OpenAPI spec
func (sh *SwaggerUIHandler) createSwaggerUIHTML(serviceName, specJSON string) string {
	config := configmanager.GetInstance()

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
        .swagger-ui .topbar {
            background-color: #667eea;
        }
        .swagger-ui .topbar .download-url-wrapper .select-label {
            color: white;
        }
        .swagger-ui .topbar .download-url-wrapper input[type=text] {
            border: 2px solid #764ba2;
        }
        .service-info {
            position: fixed;
            top: 20px;
            left: 20px;
            z-index: 1000;
            background: #667eea;
            color: white;
            padding: 15px 20px;
            border-radius: 8px;
            font-family: Arial, sans-serif;
            font-size: 14px;
            box-shadow: 0 4px 15px rgba(0,0,0,0.2);
            max-width: 300px;
        }
        .service-info h3 {
            margin: 0 0 10px 0;
            font-size: 18px;
        }
        .service-info p {
            margin: 5px 0;
            opacity: 0.9;
        }
        .environment-badge {
            display: inline-block;
            background: %s;
            color: white;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: bold;
            margin-top: 10px;
        }
    </style>
</head>
<body>
    <div class="service-info">
        <h3>%s Service</h3>
        <p><strong>Environment:</strong> %s</p>
        <p><strong>Deployment:</strong> %s</p>
        <div class="environment-badge">%s</div>
    </div>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                spec: %s,
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                onComplete: function() {
                    console.log('Swagger UI loaded successfully for %s service');
                }
            });
        };
    </script>
</body>
</html>`,
		serviceName,
		getEnvironmentColor(config.DeploymentEnv),
		serviceName,
		config.DeploymentEnv,
		config.ClassName,
		getEnvironmentBadge(config.DeploymentEnv),
		specJSON,
		serviceName)
}

// getEnvironmentColor returns the color for the environment badge
func getEnvironmentColor(env string) string {
	switch strings.ToUpper(env) {
	case "PROD":
		return "#dc3545" // Red
	case "UAT":
		return "#fd7e14" // Orange
	case "TEST":
		return "#ffc107" // Yellow
	case "DEV":
		return "#28a745" // Green
	default:
		return "#6c757d" // Gray
	}
}

// getEnvironmentBadge returns the environment badge text
func getEnvironmentBadge(env string) string {
	return strings.ToUpper(env)
}
