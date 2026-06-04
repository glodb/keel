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
	onError              func(c *SocketSession, err error)
	onNewSessionCallback func(c *SocketSession)
	onSessionDisconnect  func(c *SocketSession, reason string)

	onMessage   func(c *SocketSession, message socketmodels.Message)
	socketUsers map[string]*utilsdatatypes.Set[*SocketSession] //To send a message to specific user based on user id
	mutex       sync.RWMutex
}

var allowOriginFunc = func(r *http.Request) bool {
	return true
}

type RequestHandler struct {
	incomingRequests int
	servedRequests   int
}

// Listen starts the socket.io server on the configured address and blocks until
// the underlying HTTP server exits. It uses its own http.ServeMux so it does
// not pollute http.DefaultServeMux or conflict with the Gin HTTP server.
func (s *SocketIO) Listen() {
	server := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: allowOriginFunc,
			},
			&websocket.Transport{
				CheckOrigin: allowOriginFunc,
			},
		},
	})

	s.server = server
	s.registerTopics(s.server)
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

func (s *SocketIO) JoinRoom(c *SocketSession, roomName string) {
	c.Conn().Join(roomName)
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
	room, ok := s.socketUsers[c.UserId]

	if ok {
		roomsList := room.GetMap()
		for k := range roomsList {
			if k.conn.ID() == c.conn.ID() {
				if len(k.Conn().Rooms()) == 0 {
					room.Remove(k)
				}
			}
		}

		if room.Size() == 0 {
			delete(s.socketUsers, c.UserId)
		} else {
			s.socketUsers[c.UserId] = room
		}
	}
}

func (s *SocketIO) LeaveAll(c *SocketSession) {
	c.conn.LeaveAll()
	s.mutex.Lock()
	defer s.mutex.Unlock()

	room, ok := s.socketUsers[c.UserId]
	if ok {
		room.Clear()
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
