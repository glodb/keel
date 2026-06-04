package rpcreplymodels

import (
	"errors"
	"fmt"
	"net/http"
)

type BaseReply struct {
	Error    string `json:"error,omitempty"`
	Code     int    `json:"code,omitempty"`
	HttpCode int32  `json:"httpCode,omitempty"`
}

func (br *BaseReply) IsError() bool {
	if br.Error != "" || br.HttpCode != http.StatusOK {
		return true
	}
	return false
}
func (br *BaseReply) GetError() error {
	if br.Error != "" {
		return errors.New(br.Error)
	}
	return fmt.Errorf("error code %d", br.Code)
}

func (br *BaseReply) GetCode() int {
	return br.Code
}

func (br *BaseReply) GetHttpCode() int32 {
	return br.HttpCode
}

func (br *BaseReply) SetError(err error) {
	br.Error = err.Error()
}

func (br *BaseReply) SetCode(code int) {
	br.Code = code
}

func (br *BaseReply) SetHttpCode(code int) {
	br.HttpCode = int32(code)
}

func (br *BaseReply) SetErrorReply(err error, localCode int, httpCode int) {
	br.Error = err.Error()
	br.Code = localCode
	br.HttpCode = int32(httpCode)
}
