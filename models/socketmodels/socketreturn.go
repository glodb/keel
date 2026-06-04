package socketmodels

import "fmt"

type SocketReturn struct {
	ErrorCode        int32       `json:"errorCode" bson:"errorCode"`
	HttpResponseCode int         `json:"httpCode" bson:"httpCode"`
	ErrorMessage     string      `json:"errorMessage" bson:"errorMessage"`
	Data             interface{} `json:"data" bson:"data"`
}

func (a *SocketReturn) GetLength() int {
	return 4
}

func (d *SocketReturn) String() string {
	return "{Code:" + fmt.Sprintf("%d", d.ErrorCode) + " ErrorMessage:" + d.ErrorMessage + " Data: " + fmt.Sprintf("%s", d.Data) + "}"
}
