package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	httpHandler "github.com/glodb/keel/httpHandler"
	"github.com/glodb/keel/httpHandler/controllers"
	"github.com/glodb/keel/middlewares/basemiddlewares"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/panicrecovery"
	socketio "github.com/glodb/keel/settings/socket.io"
	"github.com/glodb/keel/settings/serviceutils"
	"github.com/glodb/keel/settings/topics"

	"github.com/nats-io/nats.go"
)

type Service struct {
	socketServer *socketio.SocketIO
}

var getInstance = sync.OnceValue(func() *Service {
	return &Service{}
})

func GetInstance() *Service {
	return getInstance()
}

// GetSocket returns the socket.io server if it was started with SERVICE_TYPE_SOCKET,
// or nil otherwise. Consumers use this to register event callbacks after Run returns.
func (s *Service) GetSocket() *socketio.SocketIO {
	return s.socketServer
}

func (s *Service) Run(serviceName string, serviceType ServiceType, subscriber serviceutils.SubscriptionInterface, middlewares map[string][]basemiddlewares.Middleware) error {

	logger.Log().Info(fmt.Sprintf("%s starting...", serviceName))

	err := panicrecovery.InitializerWithRecovery(func() error {
		return s.registerSubscriber(subscriber)
	}, fmt.Sprintf("%s.AssignSubscriber", serviceName))
	if err != nil {
		logger.Log().Error("Failed to assign subscriber", logger.ErrorField("error", err))
		return err
	}

	panicrecovery.SafeGo(func() {
		serviceutils.GetInstance().RunService()
	}, fmt.Sprintf("%s.serviceutils", serviceName))

	if serviceType&SERVICE_TYPE_HTTP != 0 {
		httpHandler.Server().Setup(middlewares)
		srv := &http.Server{
			Addr:    configmanager.GetInstance().Address,
			Handler: httpHandler.Server().GetEngine(),
		}

		func() {
			defer panicrecovery.RecoverFromPanic(fmt.Sprintf("%s.RegisterControllers", serviceName))
			controllers.InitializeControllers()
		}()

		panicrecovery.SafeGo(func() {
			logger.Log().Info("Starting server on port", logger.StringField("port", configmanager.GetInstance().Address))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Log().Error("Error starting server", logger.ErrorField("error", err))
			}
		}, fmt.Sprintf("%s HTTP server", serviceName))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Log().Error("Server forced to shutdown", logger.ErrorField("error", err))
		}
		defer cancel()
	}

	if serviceType&SERVICE_TYPE_SOCKET != 0 {
		addr := configmanager.GetInstance().SocketAddress
		if addr == "" {
			logger.Log().Error("SERVICE_TYPE_SOCKET requested but socketAddress is not set in config; socket server not started")
		} else {
			s.socketServer = socketio.New(addr)
			panicrecovery.SafeGo(func() {
				logger.Log().Info("Starting socket.io server", logger.StringField("address", addr))
				s.socketServer.Listen()
			}, fmt.Sprintf("%s socket.io server", serviceName))
		}
	}

	if serviceType&SERVICE_TYPE_SIMPLE != 0 {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Log().Info("Shutting down server...")
	}
	return nil
}

func (s *Service) registerSubscriber(subscriber serviceutils.SubscriptionInterface) error {

	subscriber.Init()
	// Get subscribed topics
	subTopics := topics.GetInstance().GetSubscribedTopics()
	// Iterate over subscribed topics
	for k, v := range subTopics {
		logger.Log().Info("Registring Event Topic", logger.StringField("topic", k), logger.StringField("function", v.(string)))
		// Check if topic is valid
		if topics.GetInstance().ValidateTopic(k) {
			// Get method by name dynamically
			m := reflect.ValueOf(subscriber).MethodByName(v.(string))
			mCallable := m.Interface().(func(msg *nats.Msg))
			// Subscribe to the topic with the corresponding method
			serviceutils.GetInstance().GetNats().QueueSubscribe(k, configmanager.GetInstance().ClassName, mCallable)
		} else {
			logger.Log().Error("This Topic is not registered", logger.StringField("topic", k))
		}
	}
	rpcSubTopics := topics.GetInstance().GetRPCSubcribedTopics()
	// Get RPC subscribed topics
	for k, v := range rpcSubTopics {
		logger.Log().Info("Registring Event Topic", logger.StringField("topic", k), logger.StringField("function", v.(string)))
		// Check if topic is valid
		if topics.GetInstance().ValidateTopic(k) {
			// Get method by name dynamically
			m := reflect.ValueOf(subscriber).MethodByName(v.(string))
			mCallable := m.Interface().(func(msg *nats.Msg))
			// Subscribe to the topic with the corresponding method
			serviceutils.GetInstance().GetNats().QueueSubscribe(k, configmanager.GetInstance().ClassName, mCallable)
		} else {
			logger.Log().Error("This Topic is not registered", logger.StringField("topic", k))
		}
	}
	return nil
}
