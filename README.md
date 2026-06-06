# keel

![Go](https://img.shields.io/badge/go-1.21+-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Version](https://img.shields.io/badge/version-v0.1-orange)

keel is a Go microservice framework that provides opinionated, batteries-included infrastructure for building production microservices.

**New service from zero to running: under 5 minutes.**
Independently verified by Claude Sonnet in Cursor.

v0.1 — opinionated, batteries-included framework. Modular architecture planned for v0.2.

## Import

```bash
go get github.com/glodb/keel
```

## Why Keel

Most Go microservice projects repeat the same wiring: HTTP server, database connections, messaging, caching, observability — from scratch, every time. Keel provides a single opinionated core so you register your service and start writing business logic immediately.

Built in 2017. Battle-tested across 8 years and multiple production systems handling millions of daily requests.

## What's included

### Service & app lifecycle
- Service registry + factory pattern (`servicehandler.Register` → `Boot(env, service)` → `Run()/Stop()`)
- Service types you can OR together: `HTTP`, `SOCKET`, `SIMPLE`, `TCP`
- Graceful shutdown on SIGINT/SIGTERM (HTTP + socket drain)
- Panic recovery everywhere — `SafeGo`, `RecoverFromPanic`, `InitializerWithRecovery`
- Multi-service single binary — isolated `NewGINServer()` / `NewControllers()` graphs (no shared global state)

### Configuration
- Layered JSON config: global `config/<env>.json` + per-service `services/<svc>/config/<env>.json`
- Env var overrides win over JSON
- Programmatic config — `NewConfig()` + `LoadFromBytes()` + `SetGlobalInstance()` (no files needed)
- Typed accessors — `GetString/GetInt/...` with `Or`/`OK` variants
- Config validation at startup (`configvalidator`)

### Database
- Three drivers: MongoDB, MySQL, PostgreSQL with pooled connections
- Rich ops: `Add/AddMany`, `Update/UpdateOne/UpdateMany`, `FindOne/Find/FindCursor`, `FindOneAndUpdate/Delete`, `Delete/DeleteOne/DeleteMany`, `BulkWrite`, `Distinct`, `Aggregate`, `Count`, `RawQuery`
- Soft delete built-in (`SoftDeleteOne/Many` → `_deleted` collections, `deletedAt`/`deletedBy`)
- Index management (`EnsureIndex`, unique + partial filters)
- Pagination in 3 flavors: `Paginate`, `SqlPaginate`, `AggregatePaginate` + array pagination
- DB migrations (`MigrationModels` + runner)
- Cross-controller access — `GetController(MONGO, "name")`

### Messaging (NATS)
- Pub/Sub and RPC request/reply
- Topic registry (registered / publishable / subscribed / RPC-subscribed)
- Batched event publisher with queue-subscribe

### Real-time
- Socket.IO — connect/disconnect/error/message callbacks, custom events (typed + raw)
- Rooms, broadcast to room / user / all
- Shared-HTTP-port or standalone mode

### API layer
- Gin HTTP with middleware tiers (`open`/`auth`/`base`) via middleware registry
- Auto OpenAPI/Swagger — generated `openapi.json` + Swagger UI (non-PROD)
- Built-in `/health` + version endpoints on every HTTP service
- Custom validators with human-readable translated messages
- API validation built-in

### Reliability & observability
- Circuit breaker (closed/open/half-open)
- Health checks — readiness/liveness/startup, with DB/cache/breaker probes
- Distributed tracing — OpenTelemetry/Jaeger (toggleable via `useTracing`)
- pprof profiling endpoints (toggle)
- Structured logging (zap)
- Structured errors — `AppError` with codes + trace/span IDs

### Notifications
- FCM push, email (gomail), WhatsApp, activity sender
- Common `NotificationSender` interface with semaphore-limited concurrency

### Utilities
- Meilisearch full-text search integration
- Distributed Redis lock (`redislock`)
- Secure cookies (`gorilla/securecookie`)
- Crypto/utils — AES encrypt/decrypt, XID generation, OAuth2 Google client, BSON helpers
- Generic data structures — `Set[T]`, `Queue[T]`
- Redis cache built-in

## Quick Start

A keel service has three parts: register, implement, boot.

### 1. Register your service

```go
package myservice

import "github.com/glodb/keel/settings/servicehandler"

func init() {
    servicehandler.Register("myservice", func() servicehandler.ServiceBase {
        return &MyService{}
    })
}
```

### 2. Implement ServiceBase

```go
type MyService struct{}

func (s *MyService) Run() error {
    // start your HTTP handlers, NATS subscribers, etc.
    return nil
}

func (s *MyService) Stop() {
    // graceful shutdown
}
```

### 3. Boot from main

```go
package main

import (
    "flag"
    "log"

    _ "myapp/myservice"
    "github.com/glodb/keel"
)

func main() {
    env := flag.String("env", "DEV", "deployment environment")
    svc := flag.String("con", "MYSERVICE", "service name")
    flag.Parse()

    if err := keel.Boot(*env, *svc); err != nil {
        log.Fatal(err)
    }
}
```

## Configuration

keel loads config from JSON files at startup:

- Global config: `config/<env>.json`
- Service config: `services/<servicename>/config/<env>.json`

Environment variables override JSON values.

## Key packages

| Package | Purpose |
|---|---|
| `settings/configmanager` | Configuration loading |
| `settings/servicehandler` | Service registration and lifecycle |
| `settings/logger` | Structured logging (zap) |
| `settings/metrics` | Prometheus metrics |
| `settings/cachesettings` | Redis cache + distributed lock |
| `settings/searchsettings` | Meilisearch integration |
| `settings/notificationsettings` | FCM, email, WhatsApp notifications |
| `settings/tracing` | OpenTelemetry distributed tracing |
| `settings/circuitbreaker` | Circuit breaker |
| `settings/errors` | Structured error types |
| `database/basefunctions` | Mongo/MySQL/PostgreSQL CRUD helpers |
| `httpHandler` | Gin-based HTTP server + Swagger |
| `settings/socket.io` | Socket.IO server |

## Dependency note

Importing `github.com/glodb/keel` pulls in the full dependency set (Mongo, MySQL, PostgreSQL, Redis, NATS, Firebase, Twilio, Stripe, OpenTelemetry, etc.). This is intentional for a batteries-included framework. Opt-in sub-package splitting is on the roadmap for v0.2.

## See it in action

👉 [keel-code](https://github.com/glodb/keel-code) — a complete working example service

## Ops / Docker reference

See [docs/ops.md](docs/ops.md) for Docker run commands and deployment notes.
