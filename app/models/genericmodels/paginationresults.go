package genericmodels

import (
	"github.com/bytedance/sonic"
)

type PaginationResults struct {
	Pagination Pagination  `json:"pagination"`
	Data       interface{} `json:"data"`
}

func (ts *PaginationResults) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *PaginationResults) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
