package rpcreplymodels

type CreateOtpReply struct {
	BaseReply
	OTP        string `json:"otp,omitempty"`
	ExpiringAt int64  `json:"expiringAt,omitempty"`
}
