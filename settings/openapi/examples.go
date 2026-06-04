package openapi

import (
	"github.com/glodb/keel/settings/logger"
)

// ExampleRequest represents an example request struct
type ExampleRequest struct {
	UserID   string `json:"userId"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

// ExampleResponse represents an example response struct
type ExampleResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		ID      string `json:"id"`
		Created string `json:"created"`
	} `json:"data"`
}

// Examples demonstrates different ways to register APIs with the new type system
func Examples() {
	helper := NewControllerHelper()

	// Example 1: Using SchemaBuilder (Fluent API)
	logger.Log().Info("Registering API using SchemaBuilder")

	requestSchema := ObjectSchema().
		Property("userId", StringSchema().
			Description("User ID").
			Example("user123").
			Build()).
		Property("message", StringSchema().
			Description("Message content").
			Example("Hello World").
			Build()).
		Property("priority", IntegerSchema().
			Description("Message priority").
			Example(1).
			Minimum(1).
			Maximum(10).
			Build()).
		Required("userId", "message").
		Build()

	responseSchema := ObjectSchema().
		Property("success", BooleanSchema().Example(true).Build()).
		Property("message", StringSchema().Example("Message sent successfully").Build()).
		Property("data", ObjectSchema().
			Property("id", StringSchema().Example("msg_123").Build()).
			Property("created", StringSchema().Example("2024-01-01T00:00:00Z").Build()).
			Build()).
		Build()

	helper.RegisterFullAPI(
		"/api/sendMessage",
		"post",
		"Send Message",
		"Send a message to a user",
		[]string{"Messaging"},
		requestSchema,
		responseSchema,
	)

	// Example 2: Using Go structs (Automatic schema generation)
	logger.Log().Info("Registering API using Go structs")

	helper.RegisterAPIFromStruct(
		"/api/sendMessageStruct",
		"post",
		"Send Message (Struct)",
		"Send a message using struct-based schema",
		[]string{"Messaging"},
		ExampleRequest{},
		ExampleResponse{},
	)

	// Example 3: Simple API without request/response schemas
	logger.Log().Info("Registering simple API")

	helper.RegisterSimpleAPI(
		"/api/health",
		"get",
		"Health Check",
		"Check service health",
		[]string{"Health"},
	)

	// Example 4: API with only request body
	logger.Log().Info("Registering API with request body only")

	helper.RegisterAPIWithBody(
		"/api/createUser",
		"post",
		"Create User",
		"Create a new user",
		[]string{"Users"},
		ObjectSchema().
			Property("name", StringSchema().Example("John Doe").Build()).
			Property("email", StringSchema().Example("john@example.com").Build()).
			Required("name", "email").
			Build(),
	)

	// Example 5: API with only response schema
	logger.Log().Info("Registering API with response schema only")

	helper.RegisterAPIWithResponse(
		"/api/getUser",
		"get",
		"Get User",
		"Get user information",
		[]string{"Users"},
		ObjectSchema().
			Property("id", StringSchema().Example("user123").Build()).
			Property("name", StringSchema().Example("John Doe").Build()).
			Property("email", StringSchema().Example("john@example.com").Build()).
			Property("createdAt", StringSchema().Example("2024-01-01T00:00:00Z").Build()).
			Build(),
	)

	// Example 6: API with complex nested schemas
	logger.Log().Info("Registering API with complex nested schemas")

	complexRequestSchema := ObjectSchema().
		Property("user", ObjectSchema().
			Property("id", StringSchema().Example("user123").Build()).
			Property("name", StringSchema().Example("John Doe").Build()).
			Build()).
		Property("items", ArraySchema().
			Items(ObjectSchema().
				Property("id", StringSchema().Example("item1").Build()).
				Property("quantity", IntegerSchema().Example(5).Build()).
				Property("price", NumberSchema().Example(10.99).Build()).
				Build()).
			Build()).
		Property("metadata", ObjectSchema().
			Property("tags", ArraySchema().
				Items(StringSchema().Example("electronics").Build()).
				Build()).
			Property("priority", StringSchema().
				Enum("low", "medium", "high").
				Example("medium").
				Build()).
			Build()).
		Required("user", "items").
		Build()

	complexResponseSchema := ObjectSchema().
		Property("success", BooleanSchema().Example(true).Build()).
		Property("orderId", StringSchema().Example("order_123").Build()).
		Property("total", NumberSchema().Example(54.95).Build()).
		Property("items", ArraySchema().
			Items(ObjectSchema().
				Property("id", StringSchema().Example("item1").Build()).
				Property("quantity", IntegerSchema().Example(5).Build()).
				Property("price", NumberSchema().Example(10.99).Build()).
				Property("subtotal", NumberSchema().Example(54.95).Build()).
				Build()).
			Build()).
		Property("createdAt", StringSchema().Example("2024-01-01T00:00:00Z").Build()).
		Build()

	helper.RegisterFullAPI(
		"/api/createOrder",
		"post",
		"Create Order",
		"Create a new order with items",
		[]string{"Orders"},
		complexRequestSchema,
		complexResponseSchema,
	)

	// Example 7: API with validation constraints
	logger.Log().Info("Registering API with validation constraints")

	validatedRequestSchema := ObjectSchema().
		Property("email", StringSchema().
			Description("User email address").
			Example("user@example.com").
			Pattern(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).
			MinLength(5).
			MaxLength(254).
			Build()).
		Property("age", IntegerSchema().
			Description("User age").
			Example(25).
			Minimum(18).
			Maximum(120).
			Build()).
		Property("password", StringSchema().
			Description("User password").
			Example("securePassword123").
			MinLength(8).
			MaxLength(128).
			Pattern(`^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d@$!%*?&]{8,}$`).
			Build()).
		Required("email", "age", "password").
		Build()

	helper.RegisterAPIWithBody(
		"/api/register",
		"post",
		"Register User",
		"Register a new user with validation",
		[]string{"Authentication"},
		validatedRequestSchema,
	)

	logger.Log().Info("All example APIs registered successfully")
}
