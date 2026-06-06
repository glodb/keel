package socketio

import (
	"errors"
	"net/http"
	"os"
	"runtime/debug"
	"sync"

	"github.com/glodb/keel/models/socketmodels"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/socket.io/socketevents"
	"github.com/glodb/keel/settings/utilsdatatypes"

	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

type SocketIO struct {
	address              string
	server               *socketio.Server
	serveOnce            sync.Once
	onError              func(c *SocketSession, err error)
	onNewSessionCallback func(c *SocketSession)
	onSessionDisconnect  func(c *SocketSession, reason string)

	onMessage   func(c *SocketSession, message socketmodels.Message)
	socketUsers map[string]*utilsdatatypes.Set[*SocketSession] //To send a message to specific user based on user id
	mutex       sync.RWMutex

	// events holds consumer-registered inbound event handlers, keyed by event
	// name. They are bound onto the underlying socket.io server in build().
	events map[string]interface{}

	// checkOrigin decides whether a cross-origin request is accepted. Defaults
	// to allow-all for backward compatibility; override with SetCheckOrigin.
	checkOrigin func(r *http.Request) bool
}

// allowOriginFunc is the default permissive CORS check (allow all origins).
var allowOriginFunc = func(r *http.Request) bool {
	return true
}

type RequestHandler struct {
	incomingRequests int
	servedRequests   int
}

// build constructs the underlying socket.io server, applies the configured
// CORS check, binds the lifecycle/message topics, and binds every consumer-
// registered custom event. It is shared by Handler() (Gin-mounted mode) and
// Listen() (standalone-port mode). It does NOT start serving.
func (s *SocketIO) build() *socketio.Server {
	origin := s.checkOrigin
	if origin == nil {
		origin = allowOriginFunc
	}

	server := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: origin,
			},
			&websocket.Transport{
				CheckOrigin: origin,
			},
		},
	})

	s.server = server
	s.registerTopics(s.server)
	s.registerEvents(s.server)
	return server
}

// Handler builds the socket.io server (if not already built), starts its
// background event loop, and returns it as an http.Handler so it can be mounted
// onto an existing router (e.g. Gin via gin.WrapH). Use this for the shared-port
// deployment where sockets live on the main HTTP server.
//
// Register events and callbacks (RegisterEvent, OnMessage, etc.) BEFORE calling
// Handler(); bindings are applied during the build and are not picked up after.
func (s *SocketIO) Handler() http.Handler {
	if s.server == nil {
		s.build()
	}
	s.serveOnce.Do(func() {
		go s.server.Serve()
	})
	return s.server
}

// Listen starts the socket.io server on the configured address and blocks until
// the underlying HTTP server exits. It uses its own http.ServeMux so it does
// not pollute http.DefaultServeMux or conflict with the Gin HTTP server. Use
// this only for the standalone (HTTP-less) socket deployment; for the shared
// Gin port use Handler() instead.
func (s *SocketIO) Listen() {
	s.build()
	go s.server.Serve()
	defer s.server.Close()

	mux := http.NewServeMux()
	mux.Handle("/socket.io/", s.server)

	srv := &http.Server{
		Addr:    s.address,
		Handler: mux,
	}

	logger.Log().Info("socket.io server listening", logger.StringField("address", s.address))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log().Error("socket.io server error", logger.ErrorField("error", err))
	}
}

// Shutdown gracefully stops the socket.io server.
func (s *SocketIO) Shutdown() {
	if s.server != nil {
		s.server.Close()
	}
}

// SetCheckOrigin overrides the CORS origin check. Pass a function that returns
// true only for allowed origins. Must be called before Handler()/Listen().
func (s *SocketIO) SetCheckOrigin(fn func(r *http.Request) bool) {
	s.checkOrigin = fn
}

// RegisterEvent registers an inbound event handler for an arbitrary event name.
// The payload is delivered to the handler as a raw JSON string; unmarshal it
// into your own type. Must be called before Handler()/Listen().
//
//	s.RegisterEvent("chatMessage", func(c *socketio.SocketSession, data string) {
//	    var m MyPayload
//	    _ = json.Unmarshal([]byte(data), &m)
//	    // ...
//	})
func (s *SocketIO) RegisterEvent(name string, handler func(c *SocketSession, data string)) {
	if s.events == nil {
		s.events = make(map[string]interface{})
	}
	s.events[name] = func(sock socketio.Conn, data string) {
		if sock.Context() == nil {
			logger.Log().Error("socket event received before session context was set",
				logger.StringField("event", name))
			return
		}
		session := sock.Context().(*SocketSession)
		handler(session, data)
	}
}

// RegisterRawEvent registers an inbound event handler with full control over
// the payload type. handler must be a func whose first parameter is
// socketio.Conn; go-socket.io decodes the event arguments into the remaining
// parameters by reflection (e.g. func(socketio.Conn, MyStruct)). Use this when
// you want typed decoding instead of the raw JSON string from RegisterEvent.
// Must be called before Handler()/Listen().
func (s *SocketIO) RegisterRawEvent(name string, handler interface{}) {
	if s.events == nil {
		s.events = make(map[string]interface{})
	}
	s.events[name] = handler
}

func (h RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func RequestServerHandler() RequestHandler {

	return RequestHandler{
		incomingRequests: 0,
		servedRequests:   0,
	}
}

func RecoverPanic(callerName string, callerPath string, w http.ResponseWriter) {

	var err error
	runError := recover()
	if runError != nil {
		switch t := runError.(type) {
		case string:
			err = errors.New(t)
		case error:
			err = t
		default:
			err = errors.New("unknown error")
		}
		logger.Log().Error("Panic Recovered", logger.ErrorField("error", err), logger.StringField("caller", callerName), logger.StringField("path", callerPath))
		debug.PrintStack()
	}
}

func RecoverWrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		path := os.Args[0]
		defer RecoverPanic("WEB SERVER"+r.URL.Path, path, w)
		//enableCors(&w)
		h.ServeHTTP(w, r)
	})
}

func (s *SocketIO) registerTopics(server *socketio.Server) {
	server.OnConnect("/", func(sock socketio.Conn) error {
		session := &SocketSession{
			conn:   sock,
			Server: s,
		}
		sock.SetContext(session)
		s.onNewSessionCallback(session)
		return nil
	})
	server.OnDisconnect("/", func(sock socketio.Conn, reason string) {

		if sock != nil {
			if sock.Context() != nil {
				session := sock.Context().(*SocketSession)
				if session != nil {
					s.LeaveAll(session)
					s.onSessionDisconnect(session, reason)
				}
				session = nil
				sock = nil
			}
		}
	})
	server.OnError("/", func(sock socketio.Conn, e error) {

		if sock != nil {
			session := sock.Context().(*SocketSession)
			if session != nil {
				s.LeaveAll(session)
				s.onError(session, e)
			}
			session = nil
			sock = nil
		}
	})

	server.OnEvent("/", socketevents.MESSAGE, func(sock socketio.Conn, variables socketmodels.Message) {
		session := sock.Context().(*SocketSession)
		s.onMessage(session, variables)
	})
}

// registerEvents binds every consumer-registered custom event onto the server.
func (s *SocketIO) registerEvents(server *socketio.Server) {
	for name, handler := range s.events {
		server.OnEvent("/", name, handler)
	}
}

func (s *SocketIO) OnNewSessionCallback(callback func(c *SocketSession)) {
	s.onNewSessionCallback = callback
}

func (s *SocketIO) OnSessionDisconnect(callback func(c *SocketSession, reason string)) {
	s.onSessionDisconnect = callback
}

func (s *SocketIO) OnError(callback func(c *SocketSession, err error)) {
	s.onError = callback
}

func (s *SocketIO) OnMessage(callback func(c *SocketSession, message socketmodels.Message)) {
	s.onMessage = callback
}

func (s *SocketIO) GetUserSessions(userId string) *utilsdatatypes.Set[*SocketSession] {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.socketUsers[userId]
}

func (s *SocketIO) CheckUserSession(userId string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, ok := s.socketUsers[userId]
	return ok
}

func (s *SocketIO) BroadcastToRoom(eventName string, roomId string, data socketmodels.SocketReturn) {

	// //fmt.Println("BroadCast to room", eventName, deviceId, data)
	s.server.BroadcastToRoom("/", roomId, eventName, data)
}

// BroadcastToUser emits an event to every live session belonging to userId.
// Use this to push a message to a specific user from outside the socket layer
// (e.g. an HTTP handler or NATS subscriber) via service.GetInstance().GetSocket().
func (s *SocketIO) BroadcastToUser(userId string, eventName string, data interface{}) {
	s.mutex.RLock()
	sessions := s.socketUsers[userId]
	s.mutex.RUnlock()

	if sessions == nil {
		return
	}
	for session := range sessions.GetMap() {
		session.SendEvent(eventName, data)
	}
}

// BroadcastToAll emits an event to every connected client on the default namespace.
func (s *SocketIO) BroadcastToAll(eventName string, data interface{}) {
	if s.server != nil {
		s.server.BroadcastToNamespace("/", eventName, data)
	}
}

// JoinRoom joins the session to a socket.io room and registers the session
// under its user id so it can be addressed later via BroadcastToUser. The
// user-id registration is skipped when UserId is empty (e.g. unauthenticated).
func (s *SocketIO) JoinRoom(c *SocketSession, roomName string) {
	c.Conn().Join(roomName)
	s.RegisterUserSession(c)
}

// RegisterUserSession records a session under its user id (UserId) so messages
// can be pushed to that user with BroadcastToUser, without joining any room.
// Call this once UserId is known (typically right after authentication). It is
// a no-op when UserId is empty.
func (s *SocketIO) RegisterUserSession(c *SocketSession) {
	if c == nil || c.UserId == "" {
		return
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.socketUsers[c.UserId] == nil {
		s.socketUsers[c.UserId] = utilsdatatypes.NewSet[*SocketSession]()
	}
	s.socketUsers[c.UserId].Add(c)
}

func (s *SocketIO) Leave(c *SocketSession, roomName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	c.Conn().Leave(roomName)
	sessions, ok := s.socketUsers[c.UserId]

	if ok {
		for session := range sessions.GetMap() {
			if session.conn.ID() == c.conn.ID() {
				if len(session.Conn().Rooms()) == 0 {
					sessions.Remove(session)
				}
			}
		}

		if sessions.Size() == 0 {
			delete(s.socketUsers, c.UserId)
		} else {
			s.socketUsers[c.UserId] = sessions
		}
	}
}

func (s *SocketIO) LeaveAll(c *SocketSession) {
	c.conn.LeaveAll()
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sessions, ok := s.socketUsers[c.UserId]
	if ok {
		sessions.Clear()
		delete(s.socketUsers, c.UserId)
	}
}

func New(address string) *SocketIO {

	server := &SocketIO{
		address:     address,
		socketUsers: make(map[string]*utilsdatatypes.Set[*SocketSession]),
	}

	server.OnNewSessionCallback(func(c *SocketSession) {})
	server.OnSessionDisconnect(func(c *SocketSession, reason string) {})
	server.OnError(func(c *SocketSession, err error) {})
	server.OnMessage(func(c *SocketSession, message socketmodels.Message) {})
	return server
}
