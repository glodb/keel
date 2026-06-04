package openapi

import (
	"github.com/glodb/keel/settings/logger"
)

// ControllerHelper provides helper functions for controllers to register their APIs
type ControllerHelper struct {
	generator *OpenAPIGenerator
}

// NewControllerHelper creates a new ControllerHelper instance
func NewControllerHelper() *ControllerHelper {
	return &ControllerHelper{
		generator: GetInstance(),
	}
}

// RegisterAPI registers an API endpoint with the OpenAPI documentation
func (ch *ControllerHelper) RegisterAPI(path, method, summary, description string, tags []string, requestBody *RequestBody, responses map[string]*Response) error {
	logger.Log().Debug("Registering API with documentation",
		logger.StringField("path", path),
		logger.StringField("method", method),
		logger.StringField("summary", summary))

	// Create operation
	operation := &Operation{
		Summary:     summary,
		Description: description,
		Tags:        tags,
		Responses:   responses,
	}

	// Add request body if provided
	if requestBody != nil {
		operation.RequestBody = requestBody
	}

	// Create path item
	pathItem := &PathItem{}

	// Set the appropriate method
	switch method {
	case "get":
		pathItem.Get = operation
	case "post":
		pathItem.Post = operation
	case "put":
		pathItem.Put = operation
	case "delete":
		pathItem.Delete = operation
	case "patch":
		pathItem.Patch = operation
	case "options":
		pathItem.Options = operation
	case "head":
		pathItem.Head = operation
	case "trace":
		pathItem.Trace = operation
	}

	err := ch.generator.AddPathToCurrentService(path, pathItem)
	if err != nil {
		logger.Log().Error("Failed to register API with documentation",
			logger.StringField("path", path),
			logger.StringField("method", method),
			logger.ErrorField("error", err))
		return err
	}

	logger.Log().Debug("API registered with documentation successfully",
		logger.StringField("path", path),
		logger.StringField("method", method),
		logger.StringField("summary", summary))

	return nil
}

// CreateRequestBody creates a request body specification
func (ch *ControllerHelper) CreateRequestBody(required bool, schema *Schema) *RequestBody {
	return &RequestBody{
		Required: required,
		Content: map[string]*MediaType{
			"application/json": {
				Schema: schema,
			},
		},
	}
}

// CreateResponse creates a response specification
func (ch *ControllerHelper) CreateResponse(description string, schema *Schema) *Response {
	response := &Response{
		Description: description,
	}

	if schema != nil {
		response.Content = map[string]*MediaType{
			"application/json": {
				Schema: schema,
			},
		}
	}

	return response
}

// CreateObjectSchema creates an object schema
func (ch *ControllerHelper) CreateObjectSchema(properties map[string]*Schema, required []string) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: properties,
	}

	if len(required) > 0 {
		schema.Required = required
	}

	return schema
}

// CreateStringSchema creates a string schema
func (ch *ControllerHelper) CreateStringSchema(description, example string) *Schema {
	schema := &Schema{
		Type: "string",
	}

	if description != "" {
		schema.Description = description
	}

	if example != "" {
		schema.Example = example
	}

	return schema
}

// CreateBooleanSchema creates a boolean schema
func (ch *ControllerHelper) CreateBooleanSchema(example bool) *Schema {
	return &Schema{
		Type:    "boolean",
		Example: example,
	}
}

// CreateIntegerSchema creates an integer schema
func (ch *ControllerHelper) CreateIntegerSchema(description string, example int) *Schema {
	schema := &Schema{
		Type:    "integer",
		Example: example,
	}

	if description != "" {
		schema.Description = description
	}

	return schema
}

// CreateArraySchema creates an array schema
func (ch *ControllerHelper) CreateArraySchema(items *Schema) *Schema {
	return &Schema{
		Type:  "array",
		Items: items,
	}
}

// RegisterSimpleAPI registers a simple API endpoint with minimal configuration
func (ch *ControllerHelper) RegisterSimpleAPI(path, method, summary, description string, tags []string) error {
	responses := map[string]*Response{
		"200": ch.CreateResponse("Success", nil),
	}

	return ch.RegisterAPI(path, method, summary, description, tags, nil, responses)
}

// RegisterAPIWithBody registers an API endpoint with request body
func (ch *ControllerHelper) RegisterAPIWithBody(path, method, summary, description string, tags []string, requestSchema *Schema) error {
	requestBody := ch.CreateRequestBody(true, requestSchema)
	responses := map[string]*Response{
		"200": ch.CreateResponse("Success", nil),
		"400": ch.CreateResponse("Bad request", nil),
	}

	return ch.RegisterAPI(path, method, summary, description, tags, requestBody, responses)
}

// RegisterAPIWithResponse registers an API endpoint with response schema
func (ch *ControllerHelper) RegisterAPIWithResponse(path, method, summary, description string, tags []string, responseSchema *Schema) error {
	responses := map[string]*Response{
		"200": ch.CreateResponse("Success", responseSchema),
	}

	return ch.RegisterAPI(path, method, summary, description, tags, nil, responses)
}

// RegisterFullAPI registers a complete API endpoint with request and response schemas
func (ch *ControllerHelper) RegisterFullAPI(path, method, summary, description string, tags []string, requestSchema, responseSchema *Schema) error {
	requestBody := ch.CreateRequestBody(true, requestSchema)
	responses := map[string]*Response{
		"200": ch.CreateResponse("Success", responseSchema),
		"400": ch.CreateResponse("Bad request", nil),
	}

	return ch.RegisterAPI(path, method, summary, description, tags, requestBody, responses)
}

// RegisterAPIFromStruct registers an API endpoint using Go structs for request/response schemas
func (ch *ControllerHelper) RegisterAPIFromStruct(path, method, summary, description string, tags []string, requestStruct, responseStruct interface{}) error {
	var requestSchema, responseSchema *Schema

	if requestStruct != nil {
		requestSchema = SchemaFromStruct(requestStruct)
	}

	if responseStruct != nil {
		responseSchema = SchemaFromStruct(responseStruct)
	}

	return ch.RegisterFullAPI(path, method, summary, description, tags, requestSchema, responseSchema)
}

// RegisterAPIFromStructWithBody registers an API endpoint with request struct and response schema
func (ch *ControllerHelper) RegisterAPIFromStructWithBody(path, method, summary, description string, tags []string, requestStruct interface{}, responseSchema *Schema) error {
	var requestSchema *Schema

	if requestStruct != nil {
		requestSchema = SchemaFromStruct(requestStruct)
	}

	return ch.RegisterFullAPI(path, method, summary, description, tags, requestSchema, responseSchema)
}

// RegisterAPIFromStructWithResponse registers an API endpoint with request schema and response struct
func (ch *ControllerHelper) RegisterAPIFromStructWithResponse(path, method, summary, description string, tags []string, requestSchema *Schema, responseStruct interface{}) error {
	var responseSchema *Schema

	if responseStruct != nil {
		responseSchema = SchemaFromStruct(responseStruct)
	}

	return ch.RegisterFullAPI(path, method, summary, description, tags, requestSchema, responseSchema)
}
