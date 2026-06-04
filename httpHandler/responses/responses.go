package responses

import (
	"path/filepath"
	"runtime"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	WELCOME_TO_keel                             = 1000
	API_NOT_AVAILABLE                           = 1001
	OPTIONS_NOT_ALLOWED                         = 1002
	CREATE_SESSION_SUCCESS                      = 1003
	GENERATING_UUID_FAILED                      = 1004
	NOT_FOUND                                   = 1005
	SESSION_ID_NOT_PRESENT                      = 1006
	SESSION_NOT_VALID                           = 1007
	REGISTER_USER_SUCCESS                       = 1008
	MALFORMED_JSON                              = 1010
	SERVER_ERROR                                = 1011
	FORBIDDEN                                   = 1012
	BAD_REQUEST                                 = 1013
	SESSION_NOT_FOUND                           = 1015
	SESSION_NOT_PROVIDED                        = 1016
	BEARER_AUTH_FAILED                          = 1017
	TOKEN_EXPIRED                               = 1018
	METHOD_NOT_AVAILABLE                        = 1019
	LOGIN_SUCCESS                               = 1020
	VALIDATION_FAILED                           = 1021
	DATA_SUCCESS                                = 1024
	DATA_FAIL                                   = 1025
	USER_FOUND                                  = 1026
	USER_NOT_FOUND                              = 1027
	INVALID_CREDENTIAL                          = 1028
	LOGOUT_SUCCESS                              = 1029
	EMAIL_FOUND                                 = 1030
	PHONE_FOUND                                 = 1031
	INVALID_PHONE                               = 1032
	INVALID_APP                                 = 1033
	PASSWORD_NOT_VALID                          = 1034
	EMAIL_SEND_FAILED                           = 1036
	UNAUTHORIZED                                = 1037
	MIN_VERSION_NOT_SATISFIED                   = 1038
	PLEASE_LOGIN_TO_PERFORM_THIS_OPERATION      = 1039
	INVALID_OTP                                 = 1040
	UPDATED_SUCCESSFULLY                        = 1041
	PLEASE_VALIDATEE_PHONE_NUMBER_FIRST         = 1042
	OTP_LIMIT                                   = 1043
	OTP_RESEND_LIMIT                            = 1044
	LOGOUT_FIRST                                = 1045
	SENT_OTP_SUCCESS                            = 1046
	PHONE_SUCCESSFULLY_VERIFIED                 = 1047
	PHONE_ALREADY_REGISTERED                    = 1048
	USER_NAME_ALREADY_REGISTERED                = 1049
	FAILED                                      = 1050
	PHONE_NOT_FOUND                             = 1051
	INVALID_VERSION                             = 1052
	RATE_LIMIT_EXCEEDED                         = 1053
	ROLE_NOT_ALLOWED_FOR_THIS_OPERATION         = 1054
	EMAIL_NOT_VERIFIED                          = 1055
	WAIT_FOR_OTP_RESEND                         = 1056
	USER_ALREADY_VERIFIED                       = 1057
	USER_DOES_NOT_EXIST                         = 1058
	USER_NOT_VERIFIED                           = 1059
	USER_ID_MISMATCH                            = 1060
	LOGOUT_FIRST_TO_LOGIN                       = 1061
	USER_NOT_LOGGED_IN                          = 1062
	PASSWORD_CHANGED_SUCCESSFULLY               = 1063
	OLD_PASSWORD_INCORRECT                      = 1064
	FORGET_PASSWORD_OTP_SENT                    = 1065
	PASSWORD_RESET_SUCCESSFUL                   = 1066
	EMAIL_ALREADY_REGISTERED                    = 1067
	EMAIL_NOT_REGISTERED                        = 1068
	LOCATION_NOT_ADDED                          = 1069
	SPORTS_NOT_ADDED                            = 1070
	SCHEDULE_NOT_ADDED                          = 1071
	SPORTS_ALREADY_EXISTS                       = 1072
	PAL_NOT_FOUND                               = 1073
	CANT_REMOVE_LAST_SPORT                      = 1074
	CANNOT_DELETE_VENUE_WITH_FACILITIES         = 1075
	SENDS_A_MESSAGE                             = 1076
	NEW_MESSAGE_RECEIVED                        = 1077
	FACILITY_SCHEDULE_NOT_WITHIN_VENUE_SCHEDULE = 1078
	VENUE_SCHEDULE_NOT_SET                      = 1079
	VENUE_NOT_FOUND                             = 1080
	SUCCESS                                     = 1081
)

type Responses struct {
	responses map[int]string
}

var getInstance = sync.OnceValue(func() *Responses {
	instance := &Responses{}
	instance.InitResponses()
	return instance
})

// Singleton. Returns a single object of Factory
func GetInstance() *Responses {
	return getInstance()
}

// InitResponses function just initialise the response headers to be sent
func (u *Responses) InitResponses() {
	u.responses = make(map[int]string)
	u.responses[WELCOME_TO_keel] = "Welcome to keel"
	u.responses[API_NOT_AVAILABLE] = "API not available"
	u.responses[OPTIONS_NOT_ALLOWED] = "Options not allowed"
	u.responses[CREATE_SESSION_SUCCESS] = "Session created successfully"
	u.responses[GENERATING_UUID_FAILED] = "Failed to generate UUID"
	u.responses[SESSION_ID_NOT_PRESENT] = "Session ID not present"
	u.responses[SESSION_NOT_VALID] = "Session not valid"
	u.responses[REGISTER_USER_SUCCESS] = "User registered successfully"
	u.responses[MALFORMED_JSON] = "Malformed JSON"
	u.responses[SERVER_ERROR] = "Internal server error"
	u.responses[FORBIDDEN] = "Forbidden"
	u.responses[BAD_REQUEST] = "Bad request"
	u.responses[SESSION_NOT_FOUND] = "Session not found"
	u.responses[SESSION_NOT_PROVIDED] = "Session not provided"
	u.responses[BEARER_AUTH_FAILED] = "Bearer authentication failed"
	u.responses[TOKEN_EXPIRED] = "Token expired"
	u.responses[METHOD_NOT_AVAILABLE] = "Method not available"
	u.responses[LOGIN_SUCCESS] = "Login successful"
	u.responses[VALIDATION_FAILED] = "Validation failed"
	u.responses[DATA_SUCCESS] = "Data fetched successfully"
	u.responses[DATA_FAIL] = "Data fetch failed"
	u.responses[USER_FOUND] = "User found"
	u.responses[USER_NOT_FOUND] = "User not found"
	u.responses[INVALID_CREDENTIAL] = "Invalid credentials"
	u.responses[LOGOUT_SUCCESS] = "Logout successful"
	u.responses[EMAIL_FOUND] = "Email found"
	u.responses[PHONE_FOUND] = "Phone found"
	u.responses[INVALID_PHONE] = "Invalid phone"
	u.responses[INVALID_APP] = "Invalid app"
	u.responses[PASSWORD_NOT_VALID] = "Password not valid"
	u.responses[EMAIL_SEND_FAILED] = "Failed to send email"
	u.responses[UNAUTHORIZED] = "Unauthorized"
	u.responses[MIN_VERSION_NOT_SATISFIED] = "Minimum version not satisfied"
	u.responses[PLEASE_LOGIN_TO_PERFORM_THIS_OPERATION] = "Please login to perform this operation"
	u.responses[INVALID_OTP] = "Invalid OTP"
	u.responses[UPDATED_SUCCESSFULLY] = "Updated successfully"
	u.responses[PLEASE_VALIDATEE_PHONE_NUMBER_FIRST] = "Please validate phone number first"
	u.responses[OTP_LIMIT] = "OTP limit reached"
	u.responses[OTP_RESEND_LIMIT] = "OTP resend limit reached"
	u.responses[LOGOUT_FIRST] = "Please logout first"
	u.responses[SENT_OTP_SUCCESS] = "OTP sent successfully"
	u.responses[PHONE_SUCCESSFULLY_VERIFIED] = "Phone successfully verified"
	u.responses[PHONE_ALREADY_REGISTERED] = "Phone already registered"
	u.responses[USER_NAME_ALREADY_REGISTERED] = "Username already registered"
	u.responses[FAILED] = "Failed"
	u.responses[PHONE_NOT_FOUND] = "Phone not found"
	u.responses[INVALID_VERSION] = "Invalid version"
	u.responses[RATE_LIMIT_EXCEEDED] = "Rate limit exceeded"
	u.responses[ROLE_NOT_ALLOWED_FOR_THIS_OPERATION] = "Role not allowed for this operation"
	u.responses[EMAIL_NOT_VERIFIED] = "Email not verified"
	u.responses[USER_ALREADY_VERIFIED] = "User already verified"
	u.responses[WAIT_FOR_OTP_RESEND] = "Please wait before resending OTP"
	u.responses[USER_DOES_NOT_EXIST] = "User does not exist"
	u.responses[USER_NOT_VERIFIED] = "User not verified"
	u.responses[USER_ID_MISMATCH] = "User ID mismatch"
	u.responses[LOGOUT_FIRST_TO_LOGIN] = "Please logout first to login"
	u.responses[USER_NOT_LOGGED_IN] = "User not logged in"
	u.responses[PASSWORD_CHANGED_SUCCESSFULLY] = "Password changed successfully"
	u.responses[OLD_PASSWORD_INCORRECT] = "Old password is incorrect"
	u.responses[FORGET_PASSWORD_OTP_SENT] = "Password reset OTP sent to email"
	u.responses[PASSWORD_RESET_SUCCESSFUL] = "Password reset successful"
	u.responses[EMAIL_ALREADY_REGISTERED] = "Email already registered"
	u.responses[EMAIL_NOT_REGISTERED] = "Email not registered"
	u.responses[LOCATION_NOT_ADDED] = "Location not added"
	u.responses[SPORTS_NOT_ADDED] = "Sports not added"
	u.responses[SCHEDULE_NOT_ADDED] = "Schedule not added"
	u.responses[SPORTS_ALREADY_EXISTS] = "Sports already exists"
	u.responses[PAL_NOT_FOUND] = "Event not found"
	u.responses[CANT_REMOVE_LAST_SPORT] = "Cannot remove the last sport"
	u.responses[CANNOT_DELETE_VENUE_WITH_FACILITIES] = "Cannot delete venue with existing facilities. Please delete all facilities first."
	u.responses[NEW_MESSAGE_RECEIVED] = "New message"
}

// GetResponse returns the message for the particular response code
func (u *Responses) GetResponse(code int) map[string]interface{} {
	message := make(map[string]interface{})
	message["code"] = code
	message["message"] = u.responses[code]

	return message
}

func SetResponse(c *gin.Context, httpcode uint, code int, err interface{}, data interface{}) {
	c.Set("code", code)
	c.Set("message", getInstance().responses[code])
	c.Set("data", data)
	c.Set("error", err)
	c.Set("httpcode", httpcode)

	if _, file, line, ok := runtime.Caller(1); ok {
		c.Set("callerFile", filepath.Base(file))
		c.Set("callerLine", line)
		// if fn := runtime.FuncForPC(pc); fn != nil {
		// 	c.Set("callerFunc", fn.Name())
		// }
	}
}
