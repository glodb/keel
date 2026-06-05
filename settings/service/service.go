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

	"github.com/glodb/keel/httpHandler/controllers"
	httpHandler "github.com/glodb/keel/internal/httpHandler"
	"github.com/glodb/keel/internal/topics"
	"github.com/glodb/keel/middlewares/basemiddlewares"
	"github.com/glodb/keel/models/socketmodels"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/panicrecovery"
	"github.com/glodb/keel/settings/serviceutils"
	socketio "github.com/glodb/keel/settings/socket.io"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

// socketConfig holds socket.io configuration registered by the consumer via the
// builder methods (OnSocketConnect, RegisterSocketEvent, etc.) before Run is
// called. It is applied onto the socket server during Run, before serving.
type socketConfig struct {
	onConnect    func(c *socketio.SocketSession)
	onDisconnect func(c *socketio.SocketSession, reason string)
	onError      func(c *socketio.SocketSession, err error)
	onMessage    func(c *socketio.SocketSession, message socketmodels.Message)
	events       map[string]func(c *socketio.SocketSession, data string)
	checkOrigin  func(r *http.Request) bool
}

type Service struct {
	socketServer *socketio.SocketIO
	address      string
	socketCfg    socketConfig
}

var getInstance = sync.OnceValue(func() *Service {
	return &Service{}
})

func GetInstance() *Service {
	return getInstance()
}

func NewService() *Service {
	return &Service{}
}

// GetSocket returns the socket.io server if it was started with SERVICE_TYPE_SOCKET,
// or nil otherwise. Use it for outbound sends (BroadcastToUser, BroadcastToRoom,
// BroadcastToAll) from HTTP handlers, NATS subscribers, etc.
func (s *Service) GetSocket() *socketio.SocketIO {
	return s.socketServer
}

// OnSocketConnect registers the callback invoked when a new socket session
// connects. Must be called before Run.
func (s *Service) OnSocketConnect(cb func(c *socketio.SocketSession)) *Service {
	s.socketCfg.onConnect = cb
	return s
}

// OnSocketDisconnect registers the callback invoked when a socket session
// disconnects. Must be called before Run.
func (s *Service) OnSocketDisconnect(cb func(c *socketio.SocketSession, reason string)) *Service {
	s.socketCfg.onDisconnect = cb
	return s
}

// OnSocketError registers the callback invoked on a socket session error.
// Must be called before Run.
func (s *Service) OnSocketError(cb func(c *socketio.SocketSession, err error)) *Service {
	s.socketCfg.onError = cb
	return s
}

// OnSocketMessage registers the handler for the built-in "socketMessage" event.
// Must be called before Run.
func (s *Service) OnSocketMessage(cb func(c *socketio.SocketSession, message socketmodels.Message)) *Service {
	s.socketCfg.onMessage = cb
	return s
}

// RegisterSocketEvent registers an inbound handler for an arbitrary event name.
// The payload arrives as a raw JSON string; unmarshal it into your own type.
// Must be called before Run.
func (s *Service) RegisterSocketEvent(name string, handler func(c *socketio.SocketSession, data string)) *Service {
	if s.socketCfg.events == nil {
		s.socketCfg.events = make(map[string]func(c *socketio.SocketSession, data string))
	}
	s.socketCfg.events[name] = handler
	return s
}

// SetSocketCORS overrides the socket.io CORS origin check (default allow-all).
// Must be called before Run.
func (s *Service) SetSocketCORS(fn func(r *http.Request) bool) *Service {
	s.socketCfg.checkOrigin = fn
	return s
}

// applySocketConfig copies the consumer-registered socket configuration onto the
// socket server. Called during Run before the server starts serving.
func (s *Service) applySocketConfig() {
	if s.socketCfg.onConnect != nil {
		s.socketServer.OnNewSessionCallback(s.socketCfg.onConnect)
	}
	if s.socketCfg.onDisconnect != nil {
		s.socketServer.OnSessionDisconnect(s.socketCfg.onDisconnect)
	}
	if s.socketCfg.onError != nil {
		s.socketServer.OnError(s.socketCfg.onError)
	}
	if s.socketCfg.onMessage != nil {
		s.socketServer.OnMessage(s.socketCfg.onMessage)
	}
	if s.socketCfg.checkOrigin != nil {
		s.socketServer.SetCheckOrigin(s.socketCfg.checkOrigin)
	}
	for name, handler := range s.socketCfg.events {
		s.socketServer.RegisterEvent(name, handler)
	}
}

// RegisterTopic marks a topic as valid. Use this when you only need the topic
// to be recognised (e.g. before publishing) but do not subscribe to it here.
func (s *Service) RegisterTopic(name string) *Service {
	topics.GetInstance().AddRegisteredTopic(name)
	return s
}

// RegisterPublishableTopic marks a topic as valid and publishable.
func (s *Service) RegisterPublishableTopic(name string) *Service {
	topics.GetInstance().AddPublishableTopic(name)
	return s
}

// RegisterSubscribedTopic registers a topic for queue subscription.
// handlerMethod is the name of the method on your SubscriptionInterface implementation
// that will be called when a message arrives on this topic.
func (s *Service) RegisterSubscribedTopic(name, handlerMethod string) *Service {
	topics.GetInstance().AddSubscribedTopic(name, handlerMethod)
	return s
}

// RegisterRpcRequestTopic marks a topic as a valid RPC request topic.
func (s *Service) RegisterRpcRequestTopic(name string) *Service {
	topics.GetInstance().AddRpcRequestTopic(name)
	return s
}

// RegisterRpcSubscribedTopic registers a topic for RPC queue subscription.
// handlerMethod is the name of the method on your SubscriptionInterface implementation
// that will be called when an RPC request arrives on this topic.
func (s *Service) RegisterRpcSubscribedTopic(name, handlerMethod string) *Service {
	topics.GetInstance().AddRpcSubscribedTopic(name, handlerMethod)
	return s
}

func (s *Service) SetAddress(address string) *Service {
	s.address = address
	return s
}

func (s *Service) Run(serviceName string, serviceType ServiceType, subscriber serviceutils.SubscriptionInterface, middlewares map[string][]basemiddlewares.Middleware) error {

	logger.Log().Info(fmt.Sprintf("%s starting...", serviceName))

	address := configmanager.GetInstance().Address
	if s.address != "" {
		address = s.address
	}
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

	// httpServer is captured here so the SERVICE_TYPE_SIMPLE block can shut it
	// down gracefully after a termination signal.
	var httpServer *http.Server

	// Create and configure the socket server up-front (if requested) so it can be
	// either mounted on the Gin engine (shared port) or served standalone.
	if serviceType&SERVICE_TYPE_SOCKET != 0 {
		s.socketServer = socketio.New(configmanager.GetInstance().SocketAddress)
		s.applySocketConfig()
	}

	if serviceType&SERVICE_TYPE_HTTP != 0 {
		server := httpHandler.NewGINServer()
		server.Setup(middlewares)

		// Shared-port mode: mount socket.io on the main Gin engine so HTTP and
		// sockets are served on the same address.
		if s.socketServer != nil {
			server.GetEngine().Any("/socket.io/*any", gin.WrapH(s.socketServer.Handler()))
			logger.Log().Info("socket.io mounted on HTTP server", logger.StringField("path", "/socket.io/"))
		}

		httpServer = &http.Server{
			Addr:    address,
			Handler: server.GetEngine(),
		}

		func() {
			defer panicrecovery.RecoverFromPanic(fmt.Sprintf("%s.RegisterControllers", serviceName))
			controllers.InitializeControllers()
		}()

		panicrecovery.SafeGo(func() {
			logger.Log().Info("Starting server on port", logger.StringField("port", address))
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Log().Error("Error starting server", logger.ErrorField("error", err))
			}
		}, fmt.Sprintf("%s HTTP server", serviceName))
	} else if s.socketServer != nil {
		// Standalone (HTTP-less) socket mode: serve on its own port.
		addr := configmanager.GetInstance().SocketAddress
		if addr == "" {
			logger.Log().Error("SERVICE_TYPE_SOCKET requested without HTTP but socketAddress is not set in config; socket server not started")
			s.socketServer = nil
		} else {
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

		if httpServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := httpServer.Shutdown(ctx); err != nil {
				logger.Log().Error("Server forced to shutdown", logger.ErrorField("error", err))
			}
		}
		if s.socketServer != nil {
			s.socketServer.Shutdown()
		}
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
