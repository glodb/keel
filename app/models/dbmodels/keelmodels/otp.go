package keelmodels

import (
	"github.com/glodb/keel/settings/customtypes"

	"github.com/bytedance/sonic"
)

type OtpDetail struct {
	SessionId  string            `json:"sessionId,omitempty" bson:"sessionId"`
	UserId     string            `json:"userId,omitempty" bson:"userId"`
	Otp        string            `json:"otp,omitempty" bson:"otp"`
	Mode       customtypes.Modes `json:"mode,omitempty" bson:"mode"`
	To         string            `json:"to,omitempty" bson:"to"`
	ResendAt   int64             `json:"resendAt,omitempty" bson:"resendAt"`
	CreatedAt  int64             `json:"createdAt,omitempty" bson:"createdAt"`
	UpdatedAt  int64             `json:"updatedAt,omitempty" bson:"updatedAt"`
	ExpiringAt int64             `json:"expiringAt,omitempty" bson:"expiringAt"`
	FromApp    string            `json:"fromApp,omitempty" bson:"fromApp"`
	Phone      string            `json:"phone,omitempty" bson:"phone"`
}

type OtpData struct {
	Otp     string `json:"otp,omitempty"`
	ToPhone string `json:"phone,omitempty"`
}

func (ts *OtpDetail) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *OtpDetail) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
