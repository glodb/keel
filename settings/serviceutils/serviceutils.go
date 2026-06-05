package serviceutils

import (
	"sync"

	"github.com/glodb/keel/models/rpcreplymodels/rpcinterface"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"

	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
)

type ServiceUtils struct {
	nats           *nats.Conn
	eventPublisher EventPublisher
	rpcRequest     RpcRequest
	exitChannel    chan bool
}

var (
	instance *ServiceUtils
	once     sync.Once
)

func NewServiceUtils(natsServerAddress string) *ServiceUtils {
	instance := &ServiceUtils{}
	instance.createConnection(natsServerAddress)
	return instance
}
func GetInstance() *ServiceUtils {
	once.Do(func() {
		logger.Log().Info("Connecting to NATS server", logger.StringField("address", configmanager.GetInstance().NatsServerAddress))

		instance.createConnection(configmanager.GetInstance().NatsServerAddress)
	})
	return instance
}

func (s *ServiceUtils) createConnection(natsServerAddress string) {
	// Configure NATS connection with error handlers and reconnection
	opts := []nats.Option{
		// Reconnect handler
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Log().Warn("NATS reconnected",
				logger.StringField("url", nc.ConnectedUrl()),
				logger.StringField("status", nc.Status().String()))
		}),
		// Disconnect handler
		nats.DisconnectHandler(func(nc *nats.Conn) {
			logger.Log().Warn("NATS disconnected",
				logger.StringField("status", nc.Status().String()))
		}),
		// Error handler
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			if sub != nil {
				logger.Log().Error("NATS subscription error",
					logger.StringField("subject", sub.Subject),
					logger.StringField("queue", sub.Queue),
					logger.ErrorField("error", err))
			} else {
				logger.Log().Error("NATS connection error",
					logger.StringField("status", nc.Status().String()),
					logger.ErrorField("error", err))
			}
		}),
		// Closed handler
		nats.ClosedHandler(func(nc *nats.Conn) {
			logger.Log().Error("NATS connection closed",
				logger.StringField("status", nc.Status().String()))
		}),
	}

	nc, err := nats.Connect(natsServerAddress, opts...)
	if err != nil {
		logger.Log().Error("Unable to connect to NATS",
			logger.StringField("address", natsServerAddress),
			logger.ErrorField("error", err))
		// Still create instance but with nil NATS - will be caught by checks
		instance = &ServiceUtils{}
		instance.nats = nil
		instance.eventPublisher = EventPublisher{}
		instance.eventPublisher.New()
		instance.exitChannel = make(chan bool)
		instance.rpcRequest.New()
		return
	}

	logger.Log().Info("Successfully connected to NATS server",
		logger.StringField("address", natsServerAddress),
		logger.StringField("url", nc.ConnectedUrl()),
		logger.StringField("status", nc.Status().String()))

	instance = &ServiceUtils{}
	instance.nats = nc
	instance.eventPublisher = EventPublisher{}
	instance.eventPublisher.New()
	instance.exitChannel = make(chan bool)
	instance.rpcRequest.New()
}

func (s *ServiceUtils) GetNats() *nats.Conn {
	return s.nats
}

func (s *ServiceUtils) PublishEvent(data interface{}, serviceName string, topic string) {

	logger.Log().Debug("Publishing Event", logger.StringField("topic", topic), logger.StringField("service_name", serviceName), logger.AnyField("data", data))
	s.eventPublisher.publishEvent(data, serviceName, configmanager.GetInstance().DeploymentEnv+topic)
}

func (s *ServiceUtils) PostReply(data rpcinterface.RPCReplyInterface, topic string) {
	dataArray, _ := sonic.Marshal(data)
	GetInstance().nats.Publish(topic, dataArray)
}

func (s *ServiceUtils) RpcRequest(request interface{}, topic string) (Msg, error) {
	if msg, err := s.rpcRequest.rpcRequest(request, configmanager.GetInstance().DeploymentEnv+topic); err == nil {
		return Msg{Data: msg.Data}, nil
	} else {
		return Msg{}, err
	}
}

func (s *ServiceUtils) Shutdown() {
	s.eventPublisher.StopTimer()

	// Check if the channel is already closed
	select {
	case <-s.exitChannel:
		// Channel is already closed
	default:
		// Channel is not closed, so close it
		close(s.exitChannel)
	}
}

func (s *ServiceUtils) RunService() {
}
