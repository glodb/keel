package controllersdocs

import (
	docinterface "github.com/glodb/keel/app/docs/docinterface"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/openapi"
)

type SessionDocs struct {
	docinterface.DocInterface
}

func (s *SessionDocs) RegisterDocs(helper *openapi.ControllerHelper) {
	logger.Log().Debug("Registering APIs for Session with documentation system")

	// Register SSO Default API (GET /)
	ssoDefaultResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1000).Build()).
		Property("message", openapi.StringSchema().Example("Welcome to keel").Build()).
		Build()

	// Create responses for SSO Default (no request body for GET)
	ssoDefaultResponses := map[string]*openapi.Response{
		"200": helper.CreateResponse("Welcome message", ssoDefaultResponseSchema),
		"405": helper.CreateResponse("Method not allowed", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(405).Build()).
			Property("message", openapi.StringSchema().Example("Method not allowed").Build()).
			Property("error", openapi.StringSchema().Example("invalid HTTP method").Build()).
			Build()),
		"412": helper.CreateResponse("Precondition failed - version mismatch", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(412).Build()).
			Property("message", openapi.StringSchema().Example("Precondition failed").Build()).
			Property("error", openapi.StringSchema().Example("version mismatch").Build()).
			Build()),
		"429": helper.CreateResponse("Too many requests - rate limit exceeded", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(429).Build()).
			Property("message", openapi.StringSchema().Example("Too many requests").Build()).
			Property("error", openapi.StringSchema().Example("rate limit exceeded").Build()).
			Build()),
	}

	err := helper.RegisterAPI(
		"/",
		"get",
		"SSO Default",
		"Welcome endpoint for keel SSO service",
		[]string{"Session Management"},
		nil, // No request body for GET
		ssoDefaultResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register SSO Default API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered SSO Default API")
	}

	// Register Create Session API (POST /api/getSession)
	createSessionRequestSchema := openapi.ObjectSchema().
		Property("app", openapi.StringSchema().
			Description("Application identifier").
			Example("keel-app").
			Build()).
		Property("fcmKey", openapi.StringSchema().
			Description("Firebase Cloud Messaging key").
			Example("fcm_key_123456").
			Build()).
		Property("version", openapi.StringSchema().
			Description("Application version").
			Example("1.0.0").
			Build()).
		Property("language", openapi.StringSchema().
			Description("User's preferred language").
			Example("ar").
			Build()).
		Property("country", openapi.StringSchema().
			Description("User's country").
			Example("Saudi Arabia").
			Build()).
		Property("city", openapi.StringSchema().
			Description("User's city").
			Example("Riyadh").
			Build()).
		Property("platform", openapi.StringSchema().
			Description("Device platform").
			Example("ios").
			Build()).
		Property("voipKey", openapi.StringSchema().
			Description("VoIP push notification key").
			Example("voip_key_123456").
			Build()).
		Property("osType", openapi.StringSchema().
			Description("Operating system type").
			Example("iOS").
			Build()).
		Property("osVersion", openapi.StringSchema().
			Description("Operating system version").
			Example("15.0").
			Build()).
		Property("deviceModel", openapi.StringSchema().
			Description("Device model").
			Example("iPhone 13").
			Build()).
		Property("deviceId", openapi.StringSchema().
			Description("Unique device identifier").
			Example("device_123456").
			Build()).
		Required("app", "version").
		Build()

	createSessionResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1003).Build()).
		Property("message", openapi.StringSchema().Example("Getting session successful").Build()).
		Property("data", openapi.ObjectSchema().
			Property("sessionId", openapi.StringSchema().Example("session_123456").Build()).
			Property("cookieValue", openapi.StringSchema().Example("cookie_123456").Build()).
			Property("language", openapi.StringSchema().Example("ar").Build()).
			Property("createdAt", openapi.IntegerSchema().Example(1640995200).Build()).
			Property("expiringAt", openapi.IntegerSchema().Example(1640998800).Build()).
			Property("app", openapi.StringSchema().Example("keel-app").Build()).
			Property("version", openapi.StringSchema().Example("1.0.0").Build()).
			Property("fcmKey", openapi.StringSchema().Example("fcm_key_123456").Build()).
			Property("platform", openapi.StringSchema().Example("ios").Build()).
			Property("voipKey", openapi.StringSchema().Example("voip_key_123456").Build()).
			Property("osType", openapi.StringSchema().Example("iOS").Build()).
			Property("osVersion", openapi.StringSchema().Example("15.0").Build()).
			Property("deviceModel", openapi.StringSchema().Example("iPhone 13").Build()).
			Property("country", openapi.StringSchema().Example("Saudi Arabia").Build()).
			Property("city", openapi.StringSchema().Example("Riyadh").Build()).
			Build()).
		Build()

	// Create request body and responses for Create Session
	createSessionRequestBody := helper.CreateRequestBody(true, createSessionRequestSchema)
	createSessionResponses := map[string]*openapi.Response{
		"200": helper.CreateResponse("Session created successfully", createSessionResponseSchema),
		"400": helper.CreateResponse("Bad request - validation error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(1013).Build()).
			Property("message", openapi.StringSchema().Example("Bad request").Build()).
			Property("error", openapi.StringSchema().Example("validation error").Build()).
			Build()),
		"401": helper.CreateResponse("Unauthorized - authentication required", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(401).Build()).
			Property("message", openapi.StringSchema().Example("Unauthorized").Build()).
			Property("error", openapi.StringSchema().Example("authentication required").Build()).
			Build()),
		"405": helper.CreateResponse("Method not allowed", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(405).Build()).
			Property("message", openapi.StringSchema().Example("Method not allowed").Build()).
			Property("error", openapi.StringSchema().Example("invalid HTTP method").Build()).
			Build()),
		"412": helper.CreateResponse("Precondition failed - version mismatch", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(412).Build()).
			Property("message", openapi.StringSchema().Example("Precondition failed").Build()).
			Property("error", openapi.StringSchema().Example("version mismatch").Build()).
			Build()),
		"429": helper.CreateResponse("Too many requests - rate limit exceeded", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(429).Build()).
			Property("message", openapi.StringSchema().Example("Too many requests").Build()).
			Property("error", openapi.StringSchema().Example("rate limit exceeded").Build()).
			Build()),
		"500": helper.CreateResponse("Internal server error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(1011).Build()).
			Property("message", openapi.StringSchema().Example("Server error").Build()).
			Property("error", openapi.StringSchema().Example("internal server error").Build()).
			Build()),
	}

	err = helper.RegisterAPI(
		"/api/getSession",
		"post",
		"Create Session",
		"Create a new user session with device information and preferences",
		[]string{"Session Management"},
		createSessionRequestBody,
		createSessionResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Create Session API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Create Session API")
	}

	// Register Update Session API (POST /api/updateSession)
	updateSessionRequestSchema := openapi.ObjectSchema().
		Property("app", openapi.StringSchema().
			Description("Application identifier").
			Example("keel-app").
			Build()).
		Property("fcmKey", openapi.StringSchema().
			Description("Firebase Cloud Messaging key").
			Example("fcm_key_123456").
			Build()).
		Property("version", openapi.StringSchema().
			Description("Application version").
			Example("1.0.0").
			Build()).
		Property("language", openapi.StringSchema().
			Description("User's preferred language").
			Example("ar").
			Build()).
		Property("country", openapi.StringSchema().
			Description("User's country").
			Example("Saudi Arabia").
			Build()).
		Property("city", openapi.StringSchema().
			Description("User's city").
			Example("Riyadh").
			Build()).
		Property("platform", openapi.StringSchema().
			Description("Device platform").
			Example("ios").
			Build()).
		Property("voipKey", openapi.StringSchema().
			Description("VoIP push notification key").
			Example("voip_key_123456").
			Build()).
		Property("osType", openapi.StringSchema().
			Description("Operating system type").
			Example("iOS").
			Build()).
		Property("osVersion", openapi.StringSchema().
			Description("Operating system version").
			Example("15.0").
			Build()).
		Property("deviceModel", openapi.StringSchema().
			Description("Device model").
			Example("iPhone 13").
			Build()).
		Property("deviceId", openapi.StringSchema().
			Description("Unique device identifier").
			Example("device_123456").
			Build()).
		Required("app", "version").
		Build()

	updateSessionResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1041).Build()).
		Property("message", openapi.StringSchema().Example("Updated successfully").Build()).
		Property("data", openapi.ObjectSchema().
			Property("sessionId", openapi.StringSchema().Example("session_123456").Build()).
			Property("cookieValue", openapi.StringSchema().Example("cookie_123456").Build()).
			Property("language", openapi.StringSchema().Example("ar").Build()).
			Property("createdAt", openapi.IntegerSchema().Example(1640995200).Build()).
			Property("expiringAt", openapi.IntegerSchema().Example(1640998800).Build()).
			Property("app", openapi.StringSchema().Example("keel-app").Build()).
			Property("version", openapi.StringSchema().Example("1.0.0").Build()).
			Property("fcmKey", openapi.StringSchema().Example("fcm_key_123456").Build()).
			Property("platform", openapi.StringSchema().Example("ios").Build()).
			Property("voipKey", openapi.StringSchema().Example("voip_key_123456").Build()).
			Property("osType", openapi.StringSchema().Example("iOS").Build()).
			Property("osVersion", openapi.StringSchema().Example("15.0").Build()).
			Property("deviceModel", openapi.StringSchema().Example("iPhone 13").Build()).
			Property("country", openapi.StringSchema().Example("Saudi Arabia").Build()).
			Property("city", openapi.StringSchema().Example("Riyadh").Build()).
			Build()).
		Build()

	// Create request body and responses for Update Session
	updateSessionRequestBody := helper.CreateRequestBody(true, updateSessionRequestSchema)
	updateSessionResponses := map[string]*openapi.Response{
		"200": helper.CreateResponse("Session updated successfully", updateSessionResponseSchema),
		"400": helper.CreateResponse("Bad request - validation error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(1013).Build()).
			Property("message", openapi.StringSchema().Example("Bad request").Build()).
			Property("error", openapi.StringSchema().Example("validation error").Build()).
			Build()),
		"401": helper.CreateResponse("Unauthorized - authentication required", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(401).Build()).
			Property("message", openapi.StringSchema().Example("Unauthorized").Build()).
			Property("error", openapi.StringSchema().Example("authentication required").Build()).
			Build()),
		"405": helper.CreateResponse("Method not allowed", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(405).Build()).
			Property("message", openapi.StringSchema().Example("Method not allowed").Build()).
			Property("error", openapi.StringSchema().Example("invalid HTTP method").Build()).
			Build()),
		"412": helper.CreateResponse("Precondition failed - version mismatch", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(412).Build()).
			Property("message", openapi.StringSchema().Example("Precondition failed").Build()).
			Property("error", openapi.StringSchema().Example("version mismatch").Build()).
			Build()),
		"429": helper.CreateResponse("Too many requests - rate limit exceeded", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(429).Build()).
			Property("message", openapi.StringSchema().Example("Too many requests").Build()).
			Property("error", openapi.StringSchema().Example("rate limit exceeded").Build()).
			Build()),
		"500": helper.CreateResponse("Internal server error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(1011).Build()).
			Property("message", openapi.StringSchema().Example("Server error").Build()).
			Property("error", openapi.StringSchema().Example("internal server error").Build()).
			Build()),
	}

	err = helper.RegisterAPI(
		"/api/updateSession",
		"post",
		"Update Session",
		"Update an existing user session with new device information and preferences. Requires authentication.",
		[]string{"Session Management"},
		updateSessionRequestBody,
		updateSessionResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Update Session API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Update Session API")
	}

	logger.Log().Debug("Successfully registered all Session APIs with documentation")
}
