# Socket.IO

keel embeds a [Socket.IO](https://socket.io/) server that you configure entirely
from your own service code. You register inbound event handlers and lifecycle
callbacks *before* the server starts, and you push outbound messages at runtime
through the live server handle.

## Enabling sockets

Sockets are off by default. Turn them on by OR-ing `SERVICE_TYPE_SOCKET` into the
service type passed to `service.GetInstance().Run(...)`.

With `SERVICE_TYPE_HTTP` also set (the recommended setup), the Socket.IO endpoint
is mounted on the **same address/port as the HTTP server** at the path
`/socket.io/`. No separate port or config is required.

```go
service.GetInstance().Run(
    configmanager.GetInstance().MicroServiceName,
    service.SERVICE_TYPE_HTTP|service.SERVICE_TYPE_SOCKET|service.SERVICE_TYPE_SIMPLE,
    subscriber,
    middlewares,
)
```

If you enable `SERVICE_TYPE_SOCKET` *without* `SERVICE_TYPE_HTTP`, the socket
server runs standalone on its own port, taken from the `socketAddress` config
field (e.g. `":9000"`). If `socketAddress` is empty in that case, the socket
server is not started.

## Receiving (inbound)

Register handlers and callbacks on the `Service` instance **before** calling
`Run`. They are applied to the socket server during startup. All builder methods
are chainable.

| Method | Purpose |
| --- | --- |
| `OnSocketConnect(func(c *socketio.SocketSession))` | New session connected |
| `OnSocketDisconnect(func(c *socketio.SocketSession, reason string))` | Session disconnected |
| `OnSocketError(func(c *socketio.SocketSession, err error))` | Session error |
| `OnSocketMessage(func(c *socketio.SocketSession, m socketmodels.Message))` | Built-in `socketMessage` event |
| `RegisterSocketEvent(name string, func(c *socketio.SocketSession, data string))` | Arbitrary custom event; payload is a raw JSON string |
| `SetSocketCORS(func(r *http.Request) bool)` | Override the default allow-all CORS check |

```go
svc := service.GetInstance()

svc.OnSocketConnect(func(c *socketio.SocketSession) {
    c.UserId = authenticate(c.Token)        // your own auth
    svc.GetSocket().RegisterUserSession(c)  // address this user later
})

svc.RegisterSocketEvent("chat", func(c *socketio.SocketSession, data string) {
    var p ChatPayload
    json.Unmarshal([]byte(data), &p)
    // ... handle inbound message ...
})
```

For fully-typed payload decoding (instead of a raw JSON string), use the lower
level `service.GetInstance().GetSocket().RegisterRawEvent(name, handler)` where
`handler` is a `func(socketio.Conn, YourStruct)`; go-socket.io decodes the event
arguments by reflection. This must be called before `Run`.

## Sending (outbound)

Get the live server handle with `service.GetInstance().GetSocket()` (returns
`nil` if sockets are not enabled) and emit from anywhere — an HTTP controller, a
NATS subscriber, a background job, etc.

| Method | Sends to |
| --- | --- |
| `SocketSession.SendEvent(event string, data interface{})` | A single session |
| `BroadcastToUser(userId, event string, data interface{})` | Every live session of a user |
| `BroadcastToRoom(event, roomId string, data socketmodels.SocketReturn)` | Every session in a room |
| `BroadcastToAll(event string, data interface{})` | Every connected client |

```go
sock := service.GetInstance().GetSocket()
sock.BroadcastToUser("user-123", "notification", socketmodels.SocketReturn{
    HttpResponseCode: http.StatusOK,
    Data:             "you have a new message",
})
```

## Rooms and users

- `JoinRoom(c, roomName)` joins the socket.io room and registers the session
  under its `UserId` (skipped when `UserId` is empty).
- `RegisterUserSession(c)` registers a session under its `UserId` without joining
  any room — call it once you know the user id (e.g. after auth).
- `Leave(c, roomName)` / `LeaveAll(c)` reverse the above.

## CORS

The default origin check allows all origins (development-friendly). Lock it down
in production with `SetSocketCORS`:

```go
svc.SetSocketCORS(func(r *http.Request) bool {
    return r.Header.Get("Origin") == "https://app.example.com"
})
```

## Reference example

See [`settings/socket.io/example_test.go`](../settings/socket.io/example_test.go)
for a complete, compilable wiring example covering inbound handlers, outbound
sends, and CORS.

## Using from keel-code (or any consumer)

The wiring lives in the consumer's `Run()`:

```go
func (s *MyService) Run() error {
    svc := service.GetInstance()

    svc.OnSocketConnect(func(c *socketio.SocketSession) { /* ... */ }).
        RegisterSocketEvent("chat", func(c *socketio.SocketSession, data string) { /* ... */ })

    svc.Run(
        configmanager.GetInstance().MicroServiceName,
        service.SERVICE_TYPE_HTTP|service.SERVICE_TYPE_SOCKET|service.SERVICE_TYPE_SIMPLE,
        &MyServiceSubscriptions{},
        middlewareregistry.GetInstance().GetMiddlewares(configmanager.GetInstance().MicroServiceName),
    )
    return nil
}
```

Note: a consumer that vendors keel (like keel-code) must re-vendor
(`go mod vendor`) to pick up these socket APIs before the snippet above will
compile there.
