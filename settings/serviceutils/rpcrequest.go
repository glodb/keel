package serviceutils

import (
	"time"

	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/topics"

	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
)

type RpcRequest struct {
}

func (r *RpcRequest) New() {

}

func (s *RpcRequest) rpcRequest(request interface{}, topic string) (*nats.Msg, error) {
	if topics.GetInstance().ValidateRequestTopics(topic) {
		if data, err := sonic.Marshal(request); err == nil {
			return GetInstance().GetNats().Request(topic, data, time.Duration(configmanager.GetInstance().RPCRequestExpirySeconds)*time.Second)
		} else {
			return nil, err
		}
	} else {
		logger.Log().Error("RPC Error", logger.StringField("service_name", configmanager.GetInstance().MicroServiceName), logger.StringField("topic", topic))
		return nil, nil
	}
}
