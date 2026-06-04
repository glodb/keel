package controllersdocs

import (
	docinterface "github.com/glodb/keel/app/docs/docinterface"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/openapi"
)

type UserDocs struct {
	docinterface.DocInterface
}

func (u *UserDocs) RegisterDocs(helper *openapi.ControllerHelper) {
	logger.Log().Debug("Registering APIs for User with documentation system")

	// Register Send OTP API
	sendOtpRequestSchema := openapi.ObjectSchema().
		Property("phone", openapi.StringSchema().
			Description("User's phone number").
			Example("+1234567890").
			Build()).
		Property("currentNationalId", openapi.StringSchema().
			Description("User's national ID").
			Example("123456789").
			Build()).
		Required("phone").
		Build()

	sendOtpResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1046).Build()).
		Property("message", openapi.StringSchema().Example("OTP sent successfully").Build()).
		Property("data", openapi.ObjectSchema().
			Property("resendAfter", openapi.IntegerSchema().Example(1640995200).Build()).
			Build()).
		Build()

	// Create request body and responses for Send OTP
	sendOtpRequestBody := helper.CreateRequestBody(true, sendOtpRequestSchema)
	sendOtpResponses := map[string]*openapi.Response{
		"200": helper.CreateResponse("OTP sent successfully", sendOtpResponseSchema),
		"400": helper.CreateResponse("Bad request - validation error or OTP limit exceeded", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(400).Build()).
			Property("message", openapi.StringSchema().Example("Bad request").Build()).
			Property("error", openapi.StringSchema().Example("OTP limit exceeded").Build()).
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
	}

	err := helper.RegisterAPI(
		"/api/sendOtp",
		"post",
		"Send OTP",
		"Send OTP to user's phone number for authentication. Requires session context.",
		[]string{"Authentication"},
		sendOtpRequestBody,
		sendOtpResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Send OTP API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Send OTP API")
	}

	// Register Verify OTP API
	verifyOtpRequestSchema := openapi.ObjectSchema().
		Property("otp", openapi.StringSchema().
			Description("OTP code received via SMS").
			Example("123456").
			Build()).
		Required("otp").
		Build()

	verifyOtpResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1020).Build()).
		Property("message", openapi.StringSchema().Example("Login successful").Build()).
		Property("data", openapi.ObjectSchema().
			Property("jwtToken", openapi.StringSchema().Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...").Build()).
			Build()).
		Build()

	// Create request body and responses for Verify OTP
	verifyOtpRequestBody := helper.CreateRequestBody(true, verifyOtpRequestSchema)
	verifyOtpResponses := map[string]*openapi.Response{
		"200": helper.CreateResponse("Login successful / Phone verified", verifyOtpResponseSchema),
		"400": helper.CreateResponse("Bad request - invalid OTP or validation error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(400).Build()).
			Property("message", openapi.StringSchema().Example("Bad request").Build()).
			Property("error", openapi.StringSchema().Example("invalid OTP").Build()).
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
	}

	err = helper.RegisterAPI(
		"/api/verifyOtp",
		"post",
		"Verify OTP",
		"Verify OTP and authenticate user. Returns JWT token on successful verification.",
		[]string{"Authentication"},
		verifyOtpRequestBody,
		verifyOtpResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Verify OTP API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Verify OTP API")
	}

	// Register Resend OTP API
	resendOtpResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1020).Build()).
		Property("message", openapi.StringSchema().Example("Login successful").Build()).
		Property("data", openapi.ObjectSchema().
			Property("otp", openapi.StringSchema().Example("123456").Build()).
			Property("expiringAt", openapi.IntegerSchema().Example(1640995200).Build()).
			Build()).
		Build()

	// Create responses for Resend OTP (no request body)
	resendOtpResponses := map[string]*openapi.Response{
		"200": helper.CreateResponse("OTP resent successfully", resendOtpResponseSchema),
		"400": helper.CreateResponse("Bad request - validation error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(400).Build()).
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
	}

	err = helper.RegisterAPI(
		"/api/resendOtp",
		"post",
		"Resend OTP",
		"Resend OTP to user for authentication. Requires existing session without verification.",
		[]string{"Authentication"},
		nil, // No request body
		resendOtpResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Resend OTP API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Resend OTP API")
	}

	// Register Check Username API
	checkUsernameRequestSchema := openapi.ObjectSchema().
		Property("userName", openapi.StringSchema().
			Description("Username to check availability").
			Example("johndoe").
			Build()).
		Required("userName").
		Build()

	checkUsernameResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1027).Build()).
		Property("message", openapi.StringSchema().Example("User not found").Build()).
		Build()

	// Create request body and responses for Check Username
	checkUsernameRequestBody := helper.CreateRequestBody(true, checkUsernameRequestSchema)
	checkUsernameResponses := map[string]*openapi.Response{
		"200": helper.CreateResponse("Username not found (available for registration)", checkUsernameResponseSchema),
		"204": helper.CreateResponse("Username found (already taken)", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(1031).Build()).
			Property("message", openapi.StringSchema().Example("Username found").Build()).
			Property("error", openapi.StringSchema().Example("username already exists").Build()).
			Build()),
		"400": helper.CreateResponse("Bad request - validation error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(400).Build()).
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
	}

	err = helper.RegisterAPI(
		"/api/checkUserName",
		"post",
		"Check Username",
		"Check if username is available for registration",
		[]string{"User Management"},
		checkUsernameRequestBody,
		checkUsernameResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Check Username API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Check Username API")
	}

	// Register Check Phone API with all possible status codes
	checkPhoneRequestSchema := openapi.ObjectSchema().
		Property("phone", openapi.StringSchema().
			Description("Phone number to check registration status").
			Example("+1234567890").
			Build()).
		Required("phone").
		Build()

	// Success response - Phone not found (available)
	checkPhoneSuccessResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1051).Build()).
		Property("message", openapi.StringSchema().Example("Phone not found").Build()).
		Build()

	// Phone found response
	checkPhoneFoundResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1031).Build()).
		Property("message", openapi.StringSchema().Example("Phone found").Build()).
		Property("error", openapi.StringSchema().Example("phone already exists").Build()).
		Build()

	// Create request body
	requestBody := helper.CreateRequestBody(true, checkPhoneRequestSchema)

	// Create comprehensive responses map
	responses := map[string]*openapi.Response{
		"200": helper.CreateResponse("Phone not found (available for registration)", checkPhoneSuccessResponseSchema),
		"204": helper.CreateResponse("Phone found (already registered)", checkPhoneFoundResponseSchema),
		"400": helper.CreateResponse("Bad request - validation error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(400).Build()).
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
	}

	err = helper.RegisterAPI(
		"/api/checkPhone",
		"post",
		"Check Phone",
		"Check if phone number is already registered in the system",
		[]string{"User Management"},
		requestBody,
		responses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Check Phone API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Check Phone API")
	}

	// Register Update Profile Photo API (Authenticated)
	updateProfilePhotoRequestSchema := openapi.ObjectSchema().
		Property("avatarUrl", openapi.StringSchema().
			Description("URL of the new avatar image").
			Example("https://example.com/avatar.jpg").
			Build()).
		Required("avatarUrl").
		Build()

	updateProfilePhotoResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1024).Build()).
		Property("message", openapi.StringSchema().Example("Data success").Build()).
		Property("data", openapi.ObjectSchema().
			Property("userId", openapi.StringSchema().Example("user123").Build()).
			Property("userName", openapi.StringSchema().Example("johndoe").Build()).
			Property("phone", openapi.StringSchema().Example("+1234567890").Build()).
			Property("avatarUrl", openapi.StringSchema().Example("https://example.com/avatar.jpg").Build()).
			Property("updatedAt", openapi.IntegerSchema().Example(1640995200).Build()).
			Build()).
		Build()

	// Create request body and responses for Update Profile Photo
	updateProfilePhotoRequestBody := helper.CreateRequestBody(true, updateProfilePhotoRequestSchema)
	updateProfilePhotoResponses := map[string]*openapi.Response{
		"200": helper.CreateResponse("Profile photo updated successfully", updateProfilePhotoResponseSchema),
		"400": helper.CreateResponse("Bad request - validation error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(400).Build()).
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
	}

	err = helper.RegisterAPI(
		"/api/updateProfilePhoto",
		"put",
		"Update Profile Photo",
		"Update user's profile photo. Requires authentication.",
		[]string{"User Profile"},
		updateProfilePhotoRequestBody,
		updateProfilePhotoResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Update Profile Photo API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Update Profile Photo API")
	}

	// Register Logout API (Authenticated)
	logoutResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1029).Build()).
		Property("message", openapi.StringSchema().Example("Logout successful").Build()).
		Property("data", openapi.ObjectSchema().
			Property("message", openapi.StringSchema().Example("Session deleted successfully").Build()).
			Build()).
		Build()

	// Create responses for Logout (no request body)
	logoutResponses := map[string]*openapi.Response{
		"200": helper.CreateResponse("Logout successful", logoutResponseSchema),
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
	}

	err = helper.RegisterAPI(
		"/api/logout",
		"post",
		"Logout",
		"Logout user and delete session. Requires authentication.",
		[]string{"Authentication"},
		nil, // No request body
		logoutResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Logout API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Logout API")
	}

	// Register Search Users API (Authenticated)
	searchUsersResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1024).Build()).
		Property("message", openapi.StringSchema().Example("Data success").Build()).
		Property("data", openapi.ArraySchema().
			Items(openapi.ObjectSchema().
				Property("userId", openapi.StringSchema().Example("user123").Build()).
				Property("userName", openapi.StringSchema().Example("johndoe").Build()).
				Property("phone", openapi.StringSchema().Example("+1234567890").Build()).
				Property("avatarUrl", openapi.StringSchema().Example("https://example.com/avatar.jpg").Build()).
				Build()).
			Build()).
		Build()

	// Create responses for Search Users (no request body for GET)
	searchUsersResponses := map[string]*openapi.Response{
		"200": helper.CreateResponse("Users found successfully", searchUsersResponseSchema),
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
	}

	err = helper.RegisterAPI(
		"/api/searchUsers",
		"get",
		"Search Users",
		"Search for users in the system. Requires authentication.",
		[]string{"User Management"},
		nil, // No request body for GET
		searchUsersResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Search Users API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Search Users API")
	}

	// Register Logout User API (Authenticated)
	logoutUserRequestSchema := openapi.ObjectSchema().
		Property("phoneNumbers", openapi.ArraySchema().
			Items(openapi.StringSchema().Example("+1234567890").Build()).
			Description("Array of phone numbers to logout").
			Build()).
		Required("phoneNumbers").
		Build()

	logoutUserResponseSchema := openapi.ObjectSchema().
		Property("message", openapi.StringSchema().Example("Success logging out user").Build()).
		Build()

	// Create request body and responses for Logout User
	logoutUserRequestBody := helper.CreateRequestBody(true, logoutUserRequestSchema)
	logoutUserResponses := map[string]*openapi.Response{
		"200": helper.CreateResponse("Users logged out successfully", logoutUserResponseSchema),
		"400": helper.CreateResponse("Bad request - validation error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(400).Build()).
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
	}

	err = helper.RegisterAPI(
		"/api/logoutUser",
		"post",
		"Logout User",
		"Logout users by phone numbers. Requires authentication.",
		[]string{"User Management"},
		logoutUserRequestBody,
		logoutUserResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Logout User API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Logout User API")
	}

	// Register Get User IDs API (Authenticated)
	getUserIdsRequestSchema := openapi.ObjectSchema().
		Property("phoneNumbers", openapi.ArraySchema().
			Items(openapi.StringSchema().Example("+1234567890").Build()).
			Description("Array of phone numbers to get user IDs for").
			Build()).
		Required("phoneNumbers").
		Build()

	getUserIdsResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1024).Build()).
		Property("message", openapi.StringSchema().Example("Data success").Build()).
		Property("data", openapi.ObjectSchema().
			Property("+1234567890", openapi.StringSchema().Example("user123").Build()).
			Property("+9876543210", openapi.StringSchema().Example("user456").Build()).
			Build()).
		Build()

	// Create request body and responses for Get User IDs
	getUserIdsRequestBody := helper.CreateRequestBody(true, getUserIdsRequestSchema)
	getUserIdsResponses := map[string]*openapi.Response{
		"200": helper.CreateResponse("User IDs retrieved successfully", getUserIdsResponseSchema),
		"400": helper.CreateResponse("Bad request - validation error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(400).Build()).
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
	}

	err = helper.RegisterAPI(
		"/api/getUserIds",
		"post",
		"Get User IDs",
		"Get user IDs by phone numbers. Requires authentication.",
		[]string{"User Management"},
		getUserIdsRequestBody,
		getUserIdsResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Get User IDs API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Get User IDs API")
	}

	// Register Get User Count API (Authenticated)
	getUserCountRequestSchema := openapi.ObjectSchema().
		Property("startTime", openapi.IntegerSchema().
			Description("Start timestamp for user count (Unix timestamp)").
			Example(1640995200).
			Build()).
		Property("endTime", openapi.IntegerSchema().
			Description("End timestamp for user count (Unix timestamp)").
			Example(1640998800).
			Build()).
		Required("startTime", "endTime").
		Build()

	getUserCountResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1024).Build()).
		Property("message", openapi.StringSchema().Example("Data success").Build()).
		Property("data", openapi.ObjectSchema().
			Property("2021-12-31", openapi.ObjectSchema().
				Property("totalUsers", openapi.IntegerSchema().Example(100).Build()).
				Property("usersWithFiftyPlusReadings", openapi.IntegerSchema().Example(25).Build()).
				Property("usersWithBelowFiftyReadings", openapi.IntegerSchema().Example(50).Build()).
				Property("usersWithZeroReadings", openapi.IntegerSchema().Example(25).Build()).
				Build()).
			Build()).
		Build()

	// Create request body and responses for Get User Count
	getUserCountRequestBody := helper.CreateRequestBody(true, getUserCountRequestSchema)
	getUserCountResponses := map[string]*openapi.Response{
		"200": helper.CreateResponse("User count statistics retrieved successfully", getUserCountResponseSchema),
		"400": helper.CreateResponse("Bad request - validation error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(400).Build()).
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
	}

	err = helper.RegisterAPI(
		"/api/getUserCount",
		"post",
		"Get User Count",
		"Get user count statistics by date range. Requires authentication.",
		[]string{"Analytics"},
		getUserCountRequestBody,
		getUserCountResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Get User Count API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Get User Count API")
	}

	logger.Log().Debug("Successfully registered all User APIs with documentation")

	// Register Register User API
	registerUserRequestSchema := openapi.ObjectSchema().
		Property("email", openapi.StringSchema().
			Description("User's email address").
			Example("user@example.com").
			Pattern(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).
			Build()).
		Property("password", openapi.StringSchema().
			Description("User's password (8-32 characters, must contain special characters)").
			Example("SecurePass123!").
			MinLength(8).
			MaxLength(32).
			Build()).
		Required("email", "password").
		Build()

	registerUserResponseSchema := openapi.ObjectSchema().
		Property("code", openapi.IntegerSchema().Example(1008).Build()).
		Property("message", openapi.StringSchema().Example("User registered successfully").Build()).
		Property("data", openapi.ObjectSchema().
			Property("userId", openapi.StringSchema().Example("user123456789").Build()).
			Property("email", openapi.StringSchema().Example("user@example.com").Build()).
			Property("message", openapi.StringSchema().Example("User registered successfully").Build()).
			Build()).
		Build()

	// Create request body and responses for Register User
	registerUserRequestBody := helper.CreateRequestBody(true, registerUserRequestSchema)

	registerUserResponses := map[string]*openapi.Response{
		"201": helper.CreateResponse("User registered successfully", registerUserResponseSchema),
		"400": helper.CreateResponse("Bad request - validation error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(400).Build()).
			Property("message", openapi.StringSchema().Example("Bad request").Build()).
			Property("error", openapi.StringSchema().Example("validation error").Build()).
			Build()),
		"409": helper.CreateResponse("Conflict - email already registered", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(409).Build()).
			Property("message", openapi.StringSchema().Example("Email already registered").Build()).
			Property("error", openapi.StringSchema().Example("Email already registered").Build()).
			Build()),
		"500": helper.CreateResponse("Internal server error", openapi.ObjectSchema().
			Property("code", openapi.IntegerSchema().Example(500).Build()).
			Property("message", openapi.StringSchema().Example("Server error").Build()).
			Property("error", openapi.StringSchema().Example("Failed to create user").Build()).
			Build()),
	}

	err = helper.RegisterAPI(
		"/api/registerUser",
		"post",
		"Register User",
		"Register a new user with email and password. Password is hashed before storage.",
		[]string{"Authentication"},
		registerUserRequestBody,
		registerUserResponses,
	)

	if err != nil {
		logger.Log().Error("Failed to register Register User API", logger.ErrorField("error", err))
	} else {
		logger.Log().Debug("Successfully registered Register User API")
	}
}
