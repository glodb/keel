package keelerrors

import (
	"fmt"
	"strings"
	"sync"

	"github.com/glodb/keel/app/models/rpcreplymodels/rpcinterface"
	"github.com/glodb/keel/settings/configmanager"

	"go.mongodb.org/mongo-driver/mongo"
)

type ErrorTypes string

const (
	RECORD_DOESNOT_EXISTS               ErrorTypes = "Record does not exist"
	DUPLICATION_EXISTS_ON_FIELDS        ErrorTypes = "Duplicated data on fields"
	APP_DOES_NOT_EXISTS                 ErrorTypes = "App doesn't exist"
	ROLE_DOES_NOT_EXISTS                ErrorTypes = "Role doesn't exist"
	MEMBER_ROLE_NOT_FOUND               ErrorTypes = "No member found with this id in organization/clinic specified"
	COUNT_MEMBER_FAILED                 ErrorTypes = "Failed to count members"
	DELETE_MEMBER_FAILED                ErrorTypes = "Failed to delete the member"
	DELETE_PROTECTED_MEMBER             ErrorTypes = "Member has protected roles and cannot be deleted"
	DB_FAILED                           ErrorTypes = "Failed on DB connection"
	STATUS_UNAUTHORIZED                 ErrorTypes = "http.StatusUnauthorized"
	ACCESS_MIDDLEWARE_ERROR             ErrorTypes = "ACCESS_MIDDLEWARE_ERROR, Please login to perfoem the operation"
	AUTH_MIDDLEWARE_ERROR_BEARER        ErrorTypes = "AUTH_MIDDLEWARE_ERROR, Bearer not provided or length is not 2"
	AUTH_MIDDLEWARE_ERROR_TOKEN_EXPIRED ErrorTypes = "AUTH_MIDDLEWARE_ERROR, Token expired"
	AUTH_MIDDLEWARE_ERROR_VALIDATION    ErrorTypes = "AUTH_MIDDLEWARE_ERROR, Validation failed"
	NO_RECORD_UPDATED                   ErrorTypes = "NO_RECORD_UPDATED"
)

type keelErrors struct {
	message string
}

func (e *keelErrors) Error() string {
	return fmt.Sprintf("keel error in service %s: %s", configmanager.GetInstance().ClassName, e.message)
}

func (e *keelErrors) GenerateError(message string) error {
	return fmt.Errorf("keel error in service %s: %s", configmanager.GetInstance().ClassName, message)
}

func (e *keelErrors) GenerateErrorWithCode(message ErrorTypes, errorType ErrorTypes) error {
	return fmt.Errorf("keelerror in %s, with code %s", message, errorType)
}

var getInstance = sync.OnceValue(func() *keelErrors {
	instance := &keelErrors{}
	return instance
})

// Singleton. Returns a single object of Errors
func GetInstance() *keelErrors {
	return getInstance()
}
func (e *keelErrors) GetRpcError(err error, code, httpCode int, reply rpcinterface.RPCReplyInterface) bool {
	if err != nil {
		reply.SetErrorReply(err, code, httpCode)
		return true
	}
	return false
}

func (e *keelErrors) GetError(errorConst ErrorTypes) *keelErrors {
	return &keelErrors{message: string(errorConst)}
}

func (e *keelErrors) ErrNoDocuments(err error) bool {
	return err == mongo.ErrNoDocuments
}

func (e *keelErrors) CheckDuplicationError(err error) bool {
	if writeErr, ok := err.(mongo.WriteException); ok {
		for _, writeError := range writeErr.WriteErrors {
			if writeError.Code == 11000 {
				return true
			}
		}
	} else if commandError, ok := err.(mongo.CommandError); ok {
		if commandError.Code == 11000 {
			return true
		}
	}
	return false
}

func (e *keelErrors) GetDuplicationErrors(err error) *keelErrors {
	if writeErr, ok := err.(mongo.WriteException); ok {
		for _, writeError := range writeErr.WriteErrors {
			if writeError.Code == 11000 {
				return e.GetFieldsError(writeError.Message)
			}
		}
	} else if commandError, ok := err.(mongo.CommandError); ok {
		if commandError.Code == 11000 {
			return e.GetFieldsError(commandError.Message)
		}
	}
	return nil
}

func (e *keelErrors) GetFieldsError(errorMessage string) *keelErrors {
	// Example error message format: "E11000 duplicate key error collection: test.example index: compound_index dup key: { field1: \"value1\", field2: \"value2\" }"
	// Split the error message by "dup key: { "
	parts := strings.Split(errorMessage, "dup key: { ")
	// Extract the second part, which contains the fields and values causing the violation
	fieldsPart := parts[len(parts)-1]
	// Remove the closing bracket '}'
	fieldsPart = strings.TrimSuffix(fieldsPart, " }")
	// Split the fields part by ", "
	fieldPairs := strings.Split(fieldsPart, ", ")

	// Iterate over the field pairs and extract field names and values
	message := string(DUPLICATION_EXISTS_ON_FIELDS)
	for id, pair := range fieldPairs {
		// Split each pair by ": "
		parts := strings.Split(pair, ": ")
		// Extract the field name and value
		field := strings.Trim(parts[0], "\"")
		value := strings.Trim(parts[1], "\"")
		// Add the field and value to the map

		if id != 0 {
			message += " and"
		}
		message += " " + field + " with value " + value
	}
	return &keelErrors{message: string(message)}
}
