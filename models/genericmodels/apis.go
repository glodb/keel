package genericmodels

import (
	"github.com/glodb/keel/settings/customtypes"

	"github.com/gin-gonic/gin"
)

// DocParam describes a single query parameter or path parameter.
type DocParam struct {
	Name        string      // parameter name as it appears in the URL / query string
	Description string      // human-readable description shown in Swagger UI
	Required    bool        // true for path params, optional for query params
	Type        string      // OpenAPI primitive: "string" | "integer" | "number" | "boolean"
	Example     interface{} // example value shown in Swagger UI
}

// DocBodyField describes one field inside a JSON request body.
type DocBodyField struct {
	Name        string      // JSON key name
	Description string      // human-readable description
	Required    bool        // whether the field is required
	Type        string      // OpenAPI primitive: "string" | "integer" | "number" | "boolean" | "array" | "object"
	Example     interface{} // example value
}

// DocResponse extends a status-code entry with an optional response body shape.
type DocResponse struct {
	Description string         // e.g. "User registered successfully"
	Body        []DocBodyField // optional — fields returned in the response body
}

// ApiDoc holds the OpenAPI documentation metadata for a single route.
// All fields are optional — a nil Doc means the route is not documented.
type ApiDoc struct {
	Summary     string            // short one-line operation summary
	Description string            // longer description shown in Swagger UI
	Tags        []string          // groups the operation in the sidebar, e.g. ["Users"]
	Deprecated  bool              // marks the operation as deprecated

	// QueryParams documents GET / query-string parameters (e.g. ?q=, ?page=).
	QueryParams []DocParam

	// PathParams documents URL path segments (e.g. /users/:id).
	PathParams []DocParam

	// Body documents the JSON request body fields.
	Body         []DocBodyField
	BodyRequired bool // whether the request body as a whole is required (default true when Body is set)

	// Responses maps HTTP status codes to descriptions (fast path, no body schema).
	// Use ResponseBodies when you also want to document response fields.
	Responses map[string]string

	// ResponseBodies maps HTTP status codes to rich descriptions with body schemas.
	// When both Responses and ResponseBodies have the same status code, ResponseBodies wins.
	ResponseBodies map[string]DocResponse

	// Security lists the security scheme names that protect this route.
	// e.g. []string{"BearerAuth"}
	// If nil, no security annotation is added (inherits global default).
	Security []string
}

type Apis struct {
	ApiName   string               `json:"apiName"`
	ApiMethod customtypes.ApiTypes `json:"apiMethod"`
	Method    gin.HandlerFunc      `json:"method"`
	Doc       *ApiDoc              // nil = route is registered but not documented
}
