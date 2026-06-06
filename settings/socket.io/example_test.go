package socketio_test

import (
	"encoding/json"
	"net/http"

	"github.com/glodb/keel/models/socketmodels"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/service"
	socketio "github.com/glodb/keel/settings/socket.io"
)

// chatPayload is an example inbound event body sent by clients on the "chat" event.
type chatPayload struct {
	RoomID  string `json:"roomId"`
	Message string `json:"message"`
}

// Example_register shows how a consumer wires socket receiving (inbound event
// handlers and lifecycle callbacks) and sending (outbound emit) from outside the
// library, mounting sockets on the main Gin HTTP port.
//
// Enable sockets by OR-ing SERVICE_TYPE_SOCKET into the service type. With
// SERVICE_TYPE_HTTP also set, the socket.io endpoint is served at /socket.io/
// on the same address as the HTTP server.
func Example_register() {
	svc := service.GetInstance()

	// Inbound: lifecycle callbacks.
	svc.OnSocketConnect(func(c *socketio.SocketSession) {
		// A client connected. Set c.UserId after your own auth, then register
		// the session so it can be addressed by user id later.
		c.UserId = "user-123"
		svc.GetSocket().RegisterUserSession(c)
	})
	svc.OnSocketDisconnect(func(c *socketio.SocketSession, reason string) {
		// A client disconnected.
	})

	// Inbound: built-in "socketMessage" event (typed Message payload).
	svc.OnSocketMessage(func(c *socketio.SocketSession, m socketmodels.Message) {
		c.SendEvent("ack", "received: "+m.Content)
	})

	// Inbound: an arbitrary custom event. Payload arrives as a raw JSON string.
	svc.RegisterSocketEvent("chat", func(c *socketio.SocketSession, data string) {
		var p chatPayload
		if err := json.Unmarshal([]byte(data), &p); err != nil {
			return
		}
		// Outbound: fan the message out to everyone in the room.
		svc.GetSocket().BroadcastToRoom("chat", p.RoomID, socketmodels.SocketReturn{
			HttpResponseCode: http.StatusOK,
			Data:             p.Message,
		})
	})

	// Restrict CORS instead of the default allow-all.
	svc.SetSocketCORS(func(r *http.Request) bool {
		return r.Header.Get("Origin") == "https://example.com"
	})

	// Start with sockets mounted on the HTTP port. This blocks until a signal.
	_ = svc.Run(
		configmanager.GetInstance().MicroServiceName,
		service.SERVICE_TYPE_HTTP|service.SERVICE_TYPE_SOCKET|service.SERVICE_TYPE_SIMPLE,
		nil,
		nil,
		false,
	)
}

// Example_sendFromOutside shows pushing a message to a specific user from code
// that runs outside the socket layer (e.g. an HTTP controller or NATS handler).
func Example_sendFromOutside() {
	sock := service.GetInstance().GetSocket()
	if sock == nil {
		return // sockets not enabled for this service
	}
	sock.BroadcastToUser("user-123", "notification", socketmodels.SocketReturn{
		HttpResponseCode: http.StatusOK,
		Data:             "you have a new message",
	})
}
