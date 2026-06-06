package socketio

import (
	socketio "github.com/googollee/go-socket.io"
)

type SocketSession struct {
	conn      socketio.Conn
	Server    *SocketIO
	SessionId string
	Verified  bool
	UserId    string
	Username  string
	Token     string
	Phone     string
}

func (s *SocketSession) Conn() socketio.Conn {
	return s.conn
}

func (s *SocketSession) SendEvent(eventName string, data interface{}) {
	// Emit asynchronously: handlers run on the socket read goroutine and
	// go-socket.io uses an unbuffered write channel, so a synchronous Emit
	// from a handler would deadlock serveRead/serveWrite.
	conn := s.conn
	go conn.Emit(eventName, data)
}

// CompareTo compares two SocketSession instances based on UserID
func (s *SocketSession) CompareTo(other *SocketSession) int {

	if s.Conn().ID() == other.Conn().ID() {
		return 0
	}
	return 1

}
