package rpcinterface

type RPCReplyInterface interface {
	IsError() bool
	GetError() error
	GetCode() int
	SetError(error)
	SetCode(int)
	SetHttpCode(int)
	SetErrorReply(err error, localCode int, httpCode int)
}
