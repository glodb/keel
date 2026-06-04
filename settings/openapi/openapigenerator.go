package openapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
)

// OpenAPIGenerator handles dynamic OpenAPI specification generation
type OpenAPIGenerator struct {
	specs map[string]*OpenAPISpec
}

var instance *OpenAPIGenerator

// GetInstance returns the singleton instance of OpenAPIGenerator
func GetInstance() *OpenAPIGenerator {
	if instance == nil {
		instance = &OpenAPIGenerator{
			specs: make(map[string]*OpenAPISpec),
		}
		instance.initializeSpecs()
	}
	return instance
}

// initializeSpecs initializes OpenAPI specifications for each service
func (oag *OpenAPIGenerator) initializeSpecs() {
	// Initialize specs for each service
	services := []string{"sso", "otp", "notification", "sessionmanager", "notificationsender"}

	for _, service := range services {
		oag.specs[service] = oag.createBaseSpec(service)
	}
}

// createBaseSpec creates a base OpenAPI specification for a service
func (oag *OpenAPIGenerator) createBaseSpec(serviceName string) *OpenAPISpec {
	title := fmt.Sprintf("%s API", strings.ToUpper(serviceName))
	description := fmt.Sprintf("API documentation for %s service", serviceName)

	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: &Info{
			Title:       title,
			Description: description,
			Version:     configmanager.GetInstance().ServiceVersion.String(),
			Contact: &Contact{
				Name:  "keel Team",
				Email: "support@keel.com",
			},
			License: &License{
				Name: "MIT",
				URL:  "https://opensource.org/licenses/MIT",
			},
		},
		Servers: []*Server{
			{
				URL:         "http://localhost:8080",
				Description: "Development server",
			},
		},
		Paths: make(map[string]*PathItem),
		Components: &Components{
			Schemas:         make(map[string]*Schema),
			SecuritySchemes: make(map[string]*SecurityScheme),
		},
		Security: []map[string][]string{
			{
				"BearerAuth": {},
			},
		},
	}

	// Add BearerAuth security scheme
	spec.Components.SecuritySchemes["BearerAuth"] = &SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT Authorization header using the Bearer scheme",
	}

	// Add service-specific paths
	oag.addServicePaths(spec, serviceName)

	return spec
}

// addServicePaths adds service-specific API paths to the specification
func (oag *OpenAPIGenerator) addServicePaths(spec *OpenAPISpec, serviceName string) {
	switch serviceName {
	case "sso":
		oag.addSSOPaths(spec)
	case "otp":
		oag.addOTPPaths(spec)
	case "notification":
		oag.addNotificationPaths(spec)
	case "sessionmanager":
		oag.addSessionManagerPaths(spec)
	case "notificationsender":
		oag.addNotificationSenderPaths(spec)
	}
}

// addSSOPaths adds SSO service API paths
func (oag *OpenAPIGenerator) addSSOPaths(spec *OpenAPISpec) {
	// Send OTP endpoint
	spec.Paths["/api/sendOtp"] = &PathItem{
		Post: &Operation{
			Summary:     "Send OTP",
			Description: "Send OTP to user's phone number for authentication",
			Tags:        []string{"Authentication"},
			RequestBody: &RequestBody{
				Required: true,
				Content: map[string]*MediaType{
					"application/json": {
						Schema: ObjectSchema().
							Property("phone", StringSchema().
								Description("User's phone number").
								Example("+1234567890").
								Build()).
							Required("phone").
							Build(),
					},
				},
			},
			Responses: map[string]*Response{
				"200": {
					Description: "OTP sent successfully",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: ObjectSchema().
								Property("success", BooleanSchema().Example(true).Build()).
								Property("message", StringSchema().Example("OTP sent successfully").Build()).
								Build(),
						},
					},
				},
				"400": {
					Description: "Bad request",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: ObjectSchema().
								Property("success", BooleanSchema().Example(false).Build()).
								Property("error", StringSchema().Example("Invalid phone number").Build()).
								Build(),
						},
					},
				},
			},
		},
	}

	// Verify OTP endpoint
	spec.Paths["/api/verifyOtp"] = &PathItem{
		Post: &Operation{
			Summary:     "Verify OTP",
			Description: "Verify OTP and authenticate user",
			Tags:        []string{"Authentication"},
			RequestBody: &RequestBody{
				Required: true,
				Content: map[string]*MediaType{
					"application/json": {
						Schema: ObjectSchema().
							Property("phone", StringSchema().
								Description("User's phone number").
								Example("+1234567890").
								Build()).
							Property("otp", StringSchema().
								Description("OTP code").
								Example("123456").
								Build()).
							Required("phone", "otp").
							Build(),
					},
				},
			},
			Responses: map[string]*Response{
				"200": {
					Description: "OTP verified successfully",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: ObjectSchema().
								Property("success", BooleanSchema().Example(true).Build()).
								Property("token", StringSchema().Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...").Build()).
								Property("user", ObjectSchema().
									Property("id", StringSchema().Example("user123").Build()).
									Property("phone", StringSchema().Example("+1234567890").Build()).
									Build()).
								Build(),
						},
					},
				},
				"400": {
					Description: "Bad request",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: ObjectSchema().
								Property("success", BooleanSchema().Example(false).Build()).
								Property("error", StringSchema().Example("Invalid OTP").Build()).
								Build(),
						},
					},
				},
			},
		},
	}
}

// addOTPPaths adds OTP service API paths
func (oag *OpenAPIGenerator) addOTPPaths(spec *OpenAPISpec) {
	spec.Paths["/api/otp/status"] = &PathItem{
		Get: &Operation{
			Summary:     "Get OTP Status",
			Description: "Get the status of OTP service",
			Tags:        []string{"OTP"},
			Responses: map[string]*Response{
				"200": {
					Description: "OTP service status",
					Content: map[string]*MediaType{
						"application/json": {
							Schema: ObjectSchema().
								Property("status", StringSchema().Example("active").Build()).
								Build(),
						},
					},
				},
			},
		},
	}
}

// addNotificationPaths adds Notification service API paths
func (oag *OpenAPIGenerator) addNotificationPaths(spec *OpenAPISpec) {
	spec.Paths["/api/notification/send"] = &PathItem{
		Post: &Operation{
			Summary:     "Send Notification",
			Description: "Send a notification to a user",
			Tags:        []string{"Notification"},
			RequestBody: &RequestBody{
				Required: true,
				Content: map[string]*MediaType{
					"application/json": {
						Schema: ObjectSchema().
							Property("userId", StringSchema().Description("User ID").Build()).
							Property("message", StringSchema().Description("Notification message").Build()).
							Required("userId", "message").
							Build(),
					},
				},
			},
			Responses: map[string]*Response{
				"200": {
					Description: "Notification sent successfully",
				},
			},
		},
	}
}

// addSessionManagerPaths adds Session Manager service API paths
func (oag *OpenAPIGenerator) addSessionManagerPaths(spec *OpenAPISpec) {
	spec.Paths["/api/session/validate"] = &PathItem{
		Post: &Operation{
			Summary:     "Validate Session",
			Description: "Validate a user session",
			Tags:        []string{"Session"},
			RequestBody: &RequestBody{
				Required: true,
				Content: map[string]*MediaType{
					"application/json": {
						Schema: ObjectSchema().
							Property("sessionId", StringSchema().Description("Session ID").Build()).
							Required("sessionId").
							Build(),
					},
				},
			},
			Responses: map[string]*Response{
				"200": {
					Description: "Session validated successfully",
				},
			},
		},
	}
}

// addNotificationSenderPaths adds Notification Sender service API paths
func (oag *OpenAPIGenerator) addNotificationSenderPaths(spec *OpenAPISpec) {
	spec.Paths["/api/notification-sender/status"] = &PathItem{
		Get: &Operation{
			Summary:     "Get Notification Sender Status",
			Description: "Get the status of notification sender service",
			Tags:        []string{"Notification Sender"},
			Responses: map[string]*Response{
				"200": {
					Description: "Notification sender service status",
				},
			},
		},
	}
}

// GetCurrentServiceSpec returns the OpenAPI specification for the current service
func (oag *OpenAPIGenerator) GetCurrentServiceSpec() (*OpenAPISpec, error) {
	config := configmanager.GetInstance()
	serviceName := strings.ToLower(config.ClassName)

	logger.Log().Debug("Getting current service spec",
		logger.StringField("className", config.ClassName),
		logger.StringField("serviceName", serviceName))

	// Map service class names to our service names
	serviceMap := map[string]string{
		"ssoservice":            "sso",
		"otpservice":            "otp",
		"notificationservice":   "notification",
		"sessionmanagerservice": "sessionmanager",
		"notificationsender":    "notificationsender",
	}

	if mappedService, exists := serviceMap[serviceName]; exists {
		logger.Log().Debug("Service mapped successfully",
			logger.StringField("originalService", serviceName),
			logger.StringField("mappedService", mappedService))
		return oag.GetSpec(mappedService)
	}

	// If service not found, return a generic spec
	logger.Log().Warn("Service not found in mapping, creating generic spec",
		logger.StringField("serviceName", serviceName))
	return oag.createBaseSpec(serviceName), nil
}

// GetSpec returns the OpenAPI specification for a specific service
func (oag *OpenAPIGenerator) GetSpec(serviceName string) (*OpenAPISpec, error) {
	if spec, exists := oag.specs[serviceName]; exists {
		return spec, nil
	}
	return nil, fmt.Errorf("specification not found for service: %s", serviceName)
}

// GetSpecJSON returns the OpenAPI specification as JSON for a specific service
func (oag *OpenAPIGenerator) GetSpecJSON(serviceName string) ([]byte, error) {
	spec, err := oag.GetSpec(serviceName)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(spec, "", "  ")
}

// GetCurrentServiceSpecJSON returns the OpenAPI specification as JSON for the current service
func (oag *OpenAPIGenerator) GetCurrentServiceSpecJSON() ([]byte, error) {
	spec, err := oag.GetCurrentServiceSpec()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(spec, "", "  ")
}

// GetSpecYAML returns the OpenAPI specification as YAML for a specific service
func (oag *OpenAPIGenerator) GetSpecYAML(serviceName string) ([]byte, error) {
	spec, err := oag.GetSpec(serviceName)
	if err != nil {
		return nil, err
	}

	// For now, return JSON as YAML conversion would require additional dependencies
	// You can add yaml.Marshal if needed
	return json.MarshalIndent(spec, "", "  ")
}

// AddPath adds a new path to a service specification
func (oag *OpenAPIGenerator) AddPath(serviceName, path string, pathItem *PathItem) error {
	spec, err := oag.GetSpec(serviceName)
	if err != nil {
		return err
	}

	spec.Paths[path] = pathItem
	return nil
}

// AddPathToCurrentService adds a new path to the current service specification
func (oag *OpenAPIGenerator) AddPathToCurrentService(path string, pathItem *PathItem) error {
	config := configmanager.GetInstance()
	serviceName := strings.ToLower(config.ClassName)

	logger.Log().Debug("Adding path to current service",
		logger.StringField("path", path),
		logger.StringField("serviceName", serviceName))

	// Map service class names to our service names
	serviceMap := map[string]string{
		"ssoservice":            "sso",
		"otpservice":            "otp",
		"notificationservice":   "notification",
		"sessionmanagerservice": "sessionmanager",
		"notificationsender":    "notificationsender",
	}

	if mappedService, exists := serviceMap[serviceName]; exists {
		logger.Log().Debug("Adding path to mapped service",
			logger.StringField("path", path),
			logger.StringField("mappedService", mappedService))
		return oag.AddPath(mappedService, path, pathItem)
	}

	// If service not found, log and return error
	logger.Log().Warn("Service not found for API documentation",
		logger.StringField("service", serviceName))
	return fmt.Errorf("service not found: %s", serviceName)
}

// AddSchema adds a new schema to a service specification
func (oag *OpenAPIGenerator) AddSchema(serviceName, schemaName string, schema *Schema) error {
	spec, err := oag.GetSpec(serviceName)
	if err != nil {
		return err
	}

	if spec.Components.Schemas == nil {
		spec.Components.Schemas = make(map[string]*Schema)
	}
	spec.Components.Schemas[schemaName] = schema
	return nil
}
