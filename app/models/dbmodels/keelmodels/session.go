package keelmodels

import (
	"github.com/bytedance/sonic"
)

//This structure mainly handle the session
/*
registrationType variable can have following values
1- System Registration
2- Google Registration
*/

type Session struct {
	SessionId    string `json:"sessionId,omitempty"`
	UserId       string `json:"userId,omitempty"`
	Email        string `json:"email,omitempty"`
	Token        string `json:"token,omitempty"`
	LastActivity int64  `json:"lastActivity,omitempty"`
	LastOTPTime  int64  `json:"lastOTPTime,omitempty"`
	CreatedAt    int64  `json:"createdAt,omitempty"`
	ExpiringAt   int64  `json:"expiringAt,omitempty"`
	FcmKey       string `json:"fcmKey,omitempty"`
	Version      string `json:"version,omitempty"`
	Platform     string `json:"platform"`
	VoipKey      string `json:"voipKey"`
	OsType       string `json:"osType"`
	OsVersion    string `json:"osVersion"`
	DeviceModel  string `json:"deviceModel"`
	DeviceId     string `json:"deviceId"`
	Language     string `json:"language,omitempty"`
	IsLogin      bool   `json:"isLogin,omitempty"`

	// RegisterLogin     bool          `json:"registerLogin,omitempty"`
}

type SessionRequest struct {
	FcmKey      string `json:"fcmKey"`
	Version     string `json:"version" validate:"required" field:"version"`
	Language    string `json:"language" validate:"required" field:"language"`
	Country     string `json:"country" validate:"required" field:"country"`
	City        string `json:"city" validate:"required" field:"city"`
	Platform    string `json:"platform" validate:"required" field:"platform"`
	VoipKey     string `json:"voipKey"`
	OsType      string `json:"osType"`
	OsVersion   string `json:"osVersion"`
	DeviceModel string `json:"deviceModel" validate:"required" field:"deviceModel"`
	DeviceId    string `json:"deviceId" validate:"required" field:"deviceId"`
}

type SessionUpdateRequest struct {
	FcmKey      string `json:"fcmKey"`
	Version     string `json:"version"`
	Language    string `json:"language"`
	Country     string `json:"country"`
	City        string `json:"city"`
	Platform    string `json:"platform"`
	VoipKey     string `json:"voipKey"`
	OsType      string `json:"osType"`
	OsVersion   string `json:"osVersion"`
	DeviceModel string `json:"deviceModel"`
	DeviceId    string `json:"deviceId"`
}

type UpdateFCMTokenRequest struct {
	FcmKey string `json:"fcmKey" validate:"required" field:"fcmKey"`
}

type UpdateActivityTokenRequest struct {
	ActivityKey string `json:"activityKey" validate:"required" field:"activityKey"`
}

func (ts *Session) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *Session) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
