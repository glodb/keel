package genericmodels

import (
	"github.com/glodb/keel/settings/customtypes"

	"github.com/gin-gonic/gin"
)

// ApiDoc holds the OpenAPI documentation metadata for a single route.
// It is the Go equivalent of Java's @Operation / @Tag / @ApiResponse annotations.
// All fields are optional — a nil Doc means the route is not documented.
type ApiDoc struct {
	Summary     string            // @Operation(summary=...)
	Description string            // @Operation(description=...)
	Tags        []string          // @Tag(name=...)
	Deprecated  bool              // marks the endpoint as deprecated in the spec
	Responses   map[string]string // HTTP status code → description, e.g. "200" → "OK"
}

type Apis struct {
	ApiName   string               `json:"apiName"`
	ApiMethod customtypes.ApiTypes `json:"apiMethod"`
	Method    gin.HandlerFunc      `json:"method"`
	Doc       *ApiDoc              // nil = route is registered but not documented
}
