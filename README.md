# keel

keel is a Go microservice framework that provides opinionated, batteries-included infrastructure for building production microservices: HTTP routing, NATS messaging, database access (Mongo/MySQL/PostgreSQL), Redis caching, search (Meilisearch), notifications (FCM/email/SMS/WhatsApp), distributed tracing, metrics, circuit breakers, and more.

## Import

```bash
go get github.com/glodb/keel
```

## Quick Start

A keel service has three parts: register, implement, boot.

### 1. Register your service

In any `init()` in your service package, register a factory function:

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

`Boot` accepts the deployment environment and service name. Your `main()` owns how those values are obtained (flags, env vars, config, etc.):

```go
package main

import (
    "flag"
    "log"

    _ "myapp/myservice" // trigger init() registration
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

Environment variables override JSON values. See `settings/configmanager` for the full config schema.

## Key packages

| Package | Purpose |
|---|---|
| `settings/configmanager` | Configuration loading; call `Configure(env, service)` before `Boot` if needed |
| `settings/servicehandler` | Service registration and lifecycle |
| `settings/logger` | Structured logging |
| `settings/metrics` | Prometheus metrics |
| `settings/cachesettings` | Redis cache |
| `settings/searchsettings` | Meilisearch integration |
| `settings/notificationsettings` | FCM, email, SMS, WhatsApp notifications |
| `settings/tracing` | OpenTelemetry distributed tracing |
| `settings/circuitbreaker` | Circuit breaker |
| `settings/errors` | Structured error types |
| `database/basefunctions` | Mongo/MySQL/PostgreSQL CRUD helpers |
| `httpHandler` | Gin-based HTTP server setup |
| `settings/socket.io` | Socket.IO server |

## Dependency note

Importing `github.com/glodb/keel` pulls in the full dependency set (Mongo, MySQL, PostgreSQL, Redis, NATS, Firebase, Twilio, Stripe, OpenTelemetry, etc.). This is intentional for a framework; opt-in sub-package splitting is on the roadmap.

## Ops / Docker reference

See [docs/ops.md](docs/ops.md) for Docker run commands and deployment notes.
