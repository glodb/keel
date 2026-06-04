package rpcreplymodels

type LoginRequestReply struct {
	BaseReply
	SessionId string `json:"sessionId"`
}
