package socketresponses

import (
	"sync"

	"github.com/glodb/keel/app/models/genericmodels"
	"github.com/glodb/keel/settings/logger"
	socketio "github.com/glodb/keel/settings/socket.io"
)

type HttpCode int

const (
	AUTH_KEY_NOT_FOUND   = 1
	SESSION_NOT_FOUND    = 2
	EVENT_SUCCESS        = 3
	BAD_REQUEST          = 4
	FAILED_TIMESTAMP_OLD = 5
)

type SocketResponses struct {
	responses map[int]string
}

var getInstance = sync.OnceValue(func() *SocketResponses {
	instance := &SocketResponses{}
	instance.initResponses()
	return instance
})

// Singleton. Returns a single object of Factory
func GetInstance() *SocketResponses {
	return getInstance()
}

// InitResponses function just initialise the response headers to be sent
func (u *SocketResponses) initResponses() {
	u.responses = make(map[int]string)
	u.responses[AUTH_KEY_NOT_FOUND] = "Auth not found"
	u.responses[SESSION_NOT_FOUND] = "User not verified"
	u.responses[EVENT_SUCCESS] = "Event success"
	u.responses[BAD_REQUEST] = "Unable to parse data"
	u.responses[FAILED_TIMESTAMP_OLD] = "You can't add previous day records, once you have added new one"

}

// GetResponse returns the message for the particular response code
func (u *SocketResponses) getResponse(code int) map[string]interface{} {
	if code == 0 {
		return map[string]interface{}{}
	}
	message := make(map[string]interface{})
	message["code"] = code
	message["message"] = u.responses[code]

	return message
}

func (u *SocketResponses) WriteResponse(httpCode int, code int, action string, socketSession *socketio.SocketSession, err error, data interface{}) map[string]interface{} {
	returnMap := u.getResponse(code)

	auditTrial := genericmodels.SocketAuditTrial{}

	auditTrial.Action = action
	auditTrial.IP = socketSession.Conn().RemoteAddr().String()
	returnMap["httpCode"] = httpCode

	if err != nil {
		returnMap["error"] = err.Error()
		auditTrial.Error = err.Error()
	}
	if data != nil {
		returnMap["data"] = data
	}

	if socketSession != nil {
		if socketSession.UserId != "" {
			returnMap["userId"] = socketSession.UserId
			auditTrial.UserID = socketSession.UserId
			auditTrial.Phone = socketSession.Phone
			auditTrial.Email = socketSession.Username
			auditTrial.Session = socketSession.SessionId
		}
	}

	socketSession.SendEvent(action, returnMap)

	logger.Log().Debug("Socket Audit Trial", logger.AnyField("auditTrial", auditTrial))
	return returnMap
}
