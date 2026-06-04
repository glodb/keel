package genericmodels

import (
	"strconv"
	"time"
)

type SocketAuditTrial struct {
	Code     int    `json:"code,omitempty"`
	IP       string `json:"ip,omitempty"`
	Session  string `json:"session,omitempty"`
	UserID   string `json:"userID,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     int    `json:"role,omitempty"`
	Action   string `json:"action,omitempty"`
	Response string `json:"response,omitempty"`
	Error    string `json:"error,omitempty"`
	Message  string `json:"message,omitempty"`
	Platform string `json:"platform,omitempty"`
	Version  string `json:"version,omitempty"`
}

func (u SocketAuditTrial) String() string {
	str := "{\"time\":\"" + time.Now().String() + "\",\"action\":\"" + u.Action + "\""

	if u.IP != "" {
		str += ",\"ip\":\"" + u.IP + "\""
	}

	// role := config.GetMapKeyString("mapAcl", strconv.Itoa(u.Role))

	// if role != "" {
	// 	str += ",\"role\":\"" + role + "\""
	// }

	if u.UserID != "" {
		str += ",\"userID\":\"" + u.UserID + "\""
	}

	if u.Phone != "" {
		str += ",\"phone\":\"" + u.Phone + "\""
	}

	if u.Email != "" {
		str += ",\"email\":\"" + u.Email + "\""
	}

	if u.Response != "" {
		str += ",\"response\":" + u.Response
	}

	if u.Platform != "" {
		str += ",\"platform\":\"" + u.Platform + "\""
	}

	if u.Version != "" {
		str += ",\"version\":\"" + u.Version + "\""
	}

	if u.Error != "" {
		str += ",\"error\":\"" + u.Error + "\""
	}

	if u.Message != "" {
		str += ",\"message\":\"" + u.Message + "\""
	}
	str += ",\"code\":\"" + strconv.Itoa(u.Code) + "\"}"
	return str
}
