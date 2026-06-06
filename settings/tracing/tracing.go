package tracing

import (
	"context"
	"sync"

	"github.com/glodb/keel/settings/configmanager"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Tracing struct {
	tracer  trace.Tracer
	tp      *sdktrace.TracerProvider
	logger  *zap.Logger
	mu      sync.RWMutex
	enabled bool
}

var (
	instance *Tracing
	once     sync.Once
)

func GetInstance() *Tracing {
	once.Do(func() {
		instance = newTracing()
	})
	return instance
}

func newTracing() *Tracing {
	t := &Tracing{logger: zap.L()}
	cfg := configmanager.GetInstance()
	if !cfg.UseTracing {
		t.setupNoOp(cfg.ServiceLBName)
		return t
	}
	t.initialize(cfg.ServiceLBName, cfg.JaegerEndpoint, cfg.ServiceVersion.String(), cfg.DeploymentEnv)
	return t
}

func NewTracing(serviceName string, jaegerEndpoint string, version string, environment string) *Tracing {
	t := &Tracing{logger: zap.L()}
	if jaegerEndpoint == "" {
		t.setupNoOp(serviceName)
		return t
	}
	t.initialize(serviceName, jaegerEndpoint, version, environment)
	return t
}

func (t *Tracing) setupNoOp(serviceName string) {
	if serviceName == "" {
		serviceName = "keel-backend"
	}
	t.enabled = false
	t.tp = sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.NeverSample()))
	otel.SetTracerProvider(t.tp)
	t.tracer = otel.Tracer(serviceName)
}

// Enabled reports whether tracing export is active.
func (t *Tracing) Enabled() bool {
	return t.enabled
}

func (t *Tracing) initialize(serviceName string, jaegerEndpoint string, version string, environment string) {
	if serviceName == "" {
		serviceName = "keel-backend"
	}

	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		t.logger.Error("Failed to create Jaeger exporter, using no-op tracer", zap.Error(err))
		t.setupNoOp(serviceName)
		return
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
			attribute.String("environment", environment),
		),
	)
	if err != nil {
		t.logger.Error("Failed to create resource, using no-op tracer", zap.Error(err))
		t.setupNoOp(serviceName)
		return
	}

	t.tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(t.tp)
	t.tracer = otel.Tracer(serviceName)
	t.enabled = true

	t.logger.Info("Tracing initialized",
		zap.String("service", serviceName),
		zap.String("jaeger_endpoint", jaegerEndpoint),
	)
}

// Shutdown flushes and stops the tracer provider. Call on process exit.
func (t *Tracing) Shutdown(ctx context.Context) error {
	if t.tp != nil {
		return t.tp.Shutdown(ctx)
	}
	return nil
}

// GetTraceID returns the trace ID string from the context.
func (t *Tracing) GetTraceID(ctx context.Context) string {
	return trace.SpanFromContext(ctx).SpanContext().TraceID().String()
}

// GetSpanID returns the span ID string from the context.
func (t *Tracing) GetSpanID(ctx context.Context) string {
	return trace.SpanFromContext(ctx).SpanContext().SpanID().String()
}

// TraceHTTPRequest starts a server span for an incoming HTTP request and returns
// the derived context (carrying the span) along with the span itself. The caller
// owns the span and must call span.End() when the request completes (typically
// via defer). route is optional; pass "" when no matched route template exists.
func (t *Tracing) TraceHTTPRequest(ctx context.Context, method, path, route string) (context.Context, trace.Span) {
	if t.tracer == nil {
		t.tracer = otel.Tracer("keel-backend")
	}

	ctx, span := t.tracer.Start(ctx, method+" "+path,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.target", path),
		),
	)
	if route != "" {
		span.SetAttributes(attribute.String("http.route", route))
	}
	return ctx, span
}

// RecordError records err on the span associated with ctx and marks the span
// status as Error. It is a no-op when err is nil, so callers can invoke it
// unconditionally.
func (t *Tracing) RecordError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
