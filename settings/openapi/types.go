package openapi

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// OpenAPISpec represents the complete OpenAPI specification
type OpenAPISpec struct {
	OpenAPI    string                `json:"openapi"`
	Info       *Info                 `json:"info"`
	Servers    []*Server             `json:"servers"`
	Paths      map[string]*PathItem  `json:"paths"`
	Components *Components           `json:"components"`
	Security   []map[string][]string `json:"security"`
}

// Info represents the API information
type Info struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Contact     *Contact `json:"contact,omitempty"`
	License     *License `json:"license,omitempty"`
}

// Contact represents contact information
type Contact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// License represents license information
type License struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Server represents a server configuration
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

// PathItem represents a path item in the API
type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Options *Operation `json:"options,omitempty"`
	Head    *Operation `json:"head,omitempty"`
	Trace   *Operation `json:"trace,omitempty"`
}

// Operation represents an API operation
type Operation struct {
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	Tags        []string              `json:"tags,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]*Response  `json:"responses"`
	Parameters  []*Parameter          `json:"parameters,omitempty"`
	Security    []map[string][]string `json:"security,omitempty"`
}

// RequestBody represents a request body
type RequestBody struct {
	Required bool                  `json:"required"`
	Content  map[string]*MediaType `json:"content"`
}

// Response represents a response
type Response struct {
	Description string                `json:"description"`
	Content     map[string]*MediaType `json:"content,omitempty"`
	Headers     map[string]*Header    `json:"headers,omitempty"`
}

// MediaType represents a media type (e.g., application/json)
type MediaType struct {
	Schema *Schema `json:"schema"`
}

// Parameter represents a parameter
type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"`
	Required    bool    `json:"required"`
	Description string  `json:"description,omitempty"`
	Schema      *Schema `json:"schema"`
}

// Header represents a header
type Header struct {
	Description string  `json:"description"`
	Schema      *Schema `json:"schema"`
}

// Components represents the components section
type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty"`
	Parameters      map[string]*Parameter      `json:"parameters,omitempty"`
	Responses       map[string]*Response       `json:"responses,omitempty"`
	RequestBodies   map[string]*RequestBody    `json:"requestBodies,omitempty"`
}

// SecurityScheme represents a security scheme
type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	Description  string `json:"description,omitempty"`
}

// Schema represents a JSON schema
type Schema struct {
	Type                 string             `json:"type,omitempty"`
	Format               string             `json:"format,omitempty"`
	Description          string             `json:"description,omitempty"`
	Example              interface{}        `json:"example,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	Enum                 []interface{}      `json:"enum,omitempty"`
	Minimum              *float64           `json:"minimum,omitempty"`
	Maximum              *float64           `json:"maximum,omitempty"`
	MinLength            *int               `json:"minLength,omitempty"`
	MaxLength            *int               `json:"maxLength,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty"`
	Ref                  string             `json:"$ref,omitempty"`
	AllOf                []*Schema          `json:"allOf,omitempty"`
	AnyOf                []*Schema          `json:"anyOf,omitempty"`
	OneOf                []*Schema          `json:"oneOf,omitempty"`
	Not                  *Schema            `json:"not,omitempty"`
}

// SchemaBuilder provides a fluent interface for building schemas
type SchemaBuilder struct {
	schema *Schema
}

// NewSchema creates a new schema builder
func NewSchema() *SchemaBuilder {
	return &SchemaBuilder{
		schema: &Schema{},
	}
}

// Type sets the schema type
func (sb *SchemaBuilder) Type(typ string) *SchemaBuilder {
	sb.schema.Type = typ
	return sb
}

// Format sets the schema format
func (sb *SchemaBuilder) Format(format string) *SchemaBuilder {
	sb.schema.Format = format
	return sb
}

// Description sets the schema description
func (sb *SchemaBuilder) Description(desc string) *SchemaBuilder {
	sb.schema.Description = desc
	return sb
}

// Example sets the schema example
func (sb *SchemaBuilder) Example(example interface{}) *SchemaBuilder {
	sb.schema.Example = example
	return sb
}

// Required sets the required fields for object schemas
func (sb *SchemaBuilder) Required(fields ...string) *SchemaBuilder {
	sb.schema.Required = fields
	return sb
}

// Property adds a property to an object schema
func (sb *SchemaBuilder) Property(name string, schema *Schema) *SchemaBuilder {
	if sb.schema.Properties == nil {
		sb.schema.Properties = make(map[string]*Schema)
	}
	sb.schema.Properties[name] = schema
	return sb
}

// Items sets the items schema for array schemas
func (sb *SchemaBuilder) Items(schema *Schema) *SchemaBuilder {
	sb.schema.Items = schema
	return sb
}

// Enum sets the enum values
func (sb *SchemaBuilder) Enum(values ...interface{}) *SchemaBuilder {
	sb.schema.Enum = values
	return sb
}

// Minimum sets the minimum value
func (sb *SchemaBuilder) Minimum(min float64) *SchemaBuilder {
	sb.schema.Minimum = &min
	return sb
}

// Maximum sets the maximum value
func (sb *SchemaBuilder) Maximum(max float64) *SchemaBuilder {
	sb.schema.Maximum = &max
	return sb
}

// MinLength sets the minimum length
func (sb *SchemaBuilder) MinLength(min int) *SchemaBuilder {
	sb.schema.MinLength = &min
	return sb
}

// MaxLength sets the maximum length
func (sb *SchemaBuilder) MaxLength(max int) *SchemaBuilder {
	sb.schema.MaxLength = &max
	return sb
}

// Pattern sets the pattern
func (sb *SchemaBuilder) Pattern(pattern string) *SchemaBuilder {
	sb.schema.Pattern = pattern
	return sb
}

// Ref sets the reference
func (sb *SchemaBuilder) Ref(ref string) *SchemaBuilder {
	sb.schema.Ref = ref
	return sb
}

// Build returns the built schema
func (sb *SchemaBuilder) Build() *Schema {
	return sb.schema
}

// Helper functions for common schema types
func StringSchema() *SchemaBuilder {
	return NewSchema().Type("string")
}

func IntegerSchema() *SchemaBuilder {
	return NewSchema().Type("integer")
}

func NumberSchema() *SchemaBuilder {
	return NewSchema().Type("number")
}

func BooleanSchema() *SchemaBuilder {
	return NewSchema().Type("boolean")
}

func ArraySchema() *SchemaBuilder {
	return NewSchema().Type("array")
}

func ObjectSchema() *SchemaBuilder {
	return NewSchema().Type("object")
}

// SchemaFromStruct creates a schema from a Go struct
func SchemaFromStruct(v interface{}) *Schema {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	schema := ObjectSchema()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Extract field name from JSON tag
		fieldName := jsonTag
		if idx := len(jsonTag); idx > 0 && jsonTag[idx-1] == ',' {
			fieldName = jsonTag[:idx-1]
		}

		// Create schema for field type
		fieldSchema := schemaFromType(field.Type)
		if fieldSchema != nil {
			schema.Property(fieldName, fieldSchema)
		}
	}

	return schema.Build()
}

// schemaFromType creates a schema from a Go type
func schemaFromType(t reflect.Type) *Schema {
	switch t.Kind() {
	case reflect.String:
		return StringSchema().Build()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return IntegerSchema().Build()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return IntegerSchema().Build()
	case reflect.Float32, reflect.Float64:
		return NumberSchema().Build()
	case reflect.Bool:
		return BooleanSchema().Build()
	case reflect.Slice, reflect.Array:
		elemType := t.Elem()
		elemSchema := schemaFromType(elemType)
		return ArraySchema().Items(elemSchema).Build()
	case reflect.Struct:
		// For structs, we could recursively build the schema
		// For now, return a generic object schema
		return ObjectSchema().Build()
	case reflect.Ptr:
		return schemaFromType(t.Elem())
	default:
		return nil
	}
}

// ToMap converts the OpenAPI spec to a map[string]interface{}
func (spec *OpenAPISpec) ToMap() map[string]interface{} {
	data, _ := json.Marshal(spec)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// FromMap creates an OpenAPI spec from a map[string]interface{}
func FromMap(data map[string]interface{}) (*OpenAPISpec, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	var spec OpenAPISpec
	if err := json.Unmarshal(jsonData, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to spec: %w", err)
	}

	return &spec, nil
}
