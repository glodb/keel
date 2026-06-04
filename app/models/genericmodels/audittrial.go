package genericmodels

import (
	"strconv"
	"time"
)

type AuditTrial struct {
	Url         string `json:"url"`
	Code        int    `json:"code"`
	IP          string `json:"ip"`
	SessionId   string `json:"sessionId"`
	Method      string `json:"method"`
	Session     string `json:"session"`
	UserID      string `json:"userID"`
	UserName    string `json:"userName"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Role        int    `json:"role"`
	QueryParams string `json:"queryParams"`
	Body        string `json:"body"`
	Response    string `json:"response"`
	Error       string `json:"error"`
	Message     string `json:"message"`
	Platform    string `json:"platform"`
	Version     string `json:"version"`
}

func (u AuditTrial) String() string {
	str := "{\"time\":\"" + time.Now().String() + "\",\"url\":\"" + u.Url + "\",\"method\":\"" + u.Method + "\""

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

	if u.QueryParams != "" {
		str += ",\"queryParams\":" + u.QueryParams
	}

	if u.Body != "" {
		str += ",\"body\":" + u.Body
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
