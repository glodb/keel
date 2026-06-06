package openapi

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/glodb/keel/models/genericmodels"
	"github.com/glodb/keel/settings/configmanager"
)

// OpenAPIGenerator builds the OpenAPI 3.0 spec at runtime from route registrations.
// Endpoints are added via RegisterEndpoint(), called automatically by apiregistration
// when a route has a non-nil Doc field — no manual wiring required.
type OpenAPIGenerator struct {
	mu          sync.RWMutex
	currentSpec *OpenAPISpec
}

var (
	oagInstance *OpenAPIGenerator
	oagOnce     sync.Once
)

// GetInstance returns the singleton OpenAPIGenerator.
func GetInstance() *OpenAPIGenerator {
	oagOnce.Do(func() {
		oagInstance = &OpenAPIGenerator{}
		oagInstance.currentSpec = oagInstance.buildBaseSpec()
	})
	return oagInstance
}

// buildBaseSpec creates the skeleton spec from the current configmanager values.
func (oag *OpenAPIGenerator) buildBaseSpec() *OpenAPISpec {
	cfg := configmanager.GetInstance()
	serviceName := cfg.ServiceLBName
	if serviceName == "" {
		serviceName = cfg.ClassName
	}

	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:   strings.ToUpper(serviceName) + " API",
			Version: cfg.ServiceVersion.String(),
		},
		Servers: []*Server{
			{URL: "http://localhost:8080", Description: "Development server"},
		},
		Paths: make(map[string]*PathItem),
		Components: &Components{
			Schemas:         make(map[string]*Schema),
			SecuritySchemes: make(map[string]*SecurityScheme),
		},
	}

	spec.Components.SecuritySchemes["BearerAuth"] = &SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT Authorization header using the Bearer scheme",
	}

	return spec
}

// RegisterEndpoint adds a route to the OpenAPI spec.
// It is called automatically by apiregistration.RegisterApis() for every Apis
// entry whose Doc field is non-nil. Controllers never need to call this directly.
func (oag *OpenAPIGenerator) RegisterEndpoint(httpMethod, path string, doc *genericmodels.ApiDoc) {
	if doc == nil {
		return
	}

	op := &Operation{
		Summary:     doc.Summary,
		Description: doc.Description,
		Tags:        doc.Tags,
		Deprecated:  doc.Deprecated,
		Responses:   make(map[string]*Response),
	}

	// --- security ---
	if len(doc.Security) > 0 {
		secEntry := make(map[string][]string)
		for _, s := range doc.Security {
			secEntry[s] = []string{}
		}
		op.Security = []map[string][]string{secEntry}
	}

	// --- path parameters ---
	for _, p := range doc.PathParams {
		op.Parameters = append(op.Parameters, &Parameter{
			Name:        p.Name,
			In:          "path",
			Required:    true,
			Description: p.Description,
			Schema:      docParamSchema(p),
		})
	}

	// --- query parameters ---
	for _, p := range doc.QueryParams {
		op.Parameters = append(op.Parameters, &Parameter{
			Name:        p.Name,
			In:          "query",
			Required:    p.Required,
			Description: p.Description,
			Schema:      docParamSchema(p),
		})
	}

	// --- request body ---
	if len(doc.Body) > 0 {
		bodySchema := &Schema{
			Type:       "object",
			Properties: make(map[string]*Schema),
		}
		for _, f := range doc.Body {
			fieldSchema := &Schema{
				Type:        docFieldType(f.Type),
				Description: f.Description,
				Example:     f.Example,
			}
			bodySchema.Properties[f.Name] = fieldSchema
			if f.Required {
				bodySchema.Required = append(bodySchema.Required, f.Name)
			}
		}
		required := doc.BodyRequired || len(doc.Body) > 0
		op.RequestBody = &RequestBody{
			Required: required,
			Content: map[string]*MediaType{
				"application/json": {Schema: bodySchema},
			},
		}
	}

	// --- responses (simple) ---
	for code, description := range doc.Responses {
		if _, overridden := doc.ResponseBodies[code]; !overridden {
			op.Responses[code] = &Response{Description: description}
		}
	}

	// --- responses (rich with body schema) ---
	for code, rb := range doc.ResponseBodies {
		resp := &Response{Description: rb.Description}
		if len(rb.Body) > 0 {
			bodySchema := &Schema{
				Type:       "object",
				Properties: make(map[string]*Schema),
			}
			for _, f := range rb.Body {
				bodySchema.Properties[f.Name] = &Schema{
					Type:        docFieldType(f.Type),
					Description: f.Description,
					Example:     f.Example,
				}
			}
			resp.Content = map[string]*MediaType{
				"application/json": {Schema: bodySchema},
			}
		}
		op.Responses[code] = resp
	}

	if len(op.Responses) == 0 {
		op.Responses["200"] = &Response{Description: "OK"}
	}

	oag.mu.Lock()
	defer oag.mu.Unlock()

	if _, exists := oag.currentSpec.Paths[path]; !exists {
		oag.currentSpec.Paths[path] = &PathItem{}
	}

	switch strings.ToUpper(httpMethod) {
	case "GET":
		oag.currentSpec.Paths[path].Get = op
	case "POST":
		oag.currentSpec.Paths[path].Post = op
	case "PUT":
		oag.currentSpec.Paths[path].Put = op
	case "DELETE":
		oag.currentSpec.Paths[path].Delete = op
	case "PATCH":
		oag.currentSpec.Paths[path].Patch = op
	}
}

// docParamSchema converts a DocParam type string to an OpenAPI Schema.
func docParamSchema(p genericmodels.DocParam) *Schema {
	s := &Schema{
		Type:    docFieldType(p.Type),
		Example: p.Example,
	}
	if s.Type == "" {
		s.Type = "string"
	}
	return s
}

// docFieldType normalises a free-text type name to an OpenAPI primitive.
func docFieldType(t string) string {
	switch strings.ToLower(t) {
	case "int", "int32", "int64", "integer":
		return "integer"
	case "float", "float32", "float64", "number":
		return "number"
	case "bool", "boolean":
		return "boolean"
	case "array":
		return "array"
	case "object":
		return "object"
	default:
		return "string"
	}
}

// GetCurrentServiceSpecJSON returns the live OpenAPI spec as indented JSON.
// Used by the Swagger UI handler.
func (oag *OpenAPIGenerator) GetCurrentServiceSpecJSON() ([]byte, error) {
	oag.mu.RLock()
	defer oag.mu.RUnlock()
	return json.MarshalIndent(oag.currentSpec, "", "  ")
}
