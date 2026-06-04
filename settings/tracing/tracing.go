package tracing

import (
	"context"
	"fmt"
	"sync"
	"time"

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

// Tracing provides OpenTelemetry distributed tracing
type Tracing struct {
	tracer trace.Tracer
	tp     *sdktrace.TracerProvider
	logger *zap.Logger
	mu     sync.RWMutex
}

var (
	instance *Tracing
	once     sync.Once
)

// GetInstance returns the singleton tracing instance
func GetInstance() *Tracing {
	once.Do(func() {
		instance = newTracing()
	})
	return instance
}

// newTracing creates a new tracing instance
func newTracing() *Tracing {
	t := &Tracing{
		logger: zap.L(),
	}

	t.initializeTracing()
	return t
}

// initializeTracing initializes OpenTelemetry tracing
func (t *Tracing) initializeTracing() {
	// Get service configuration from environment
	serviceName := configmanager.GetInstance().ServiceLBName
	if serviceName == "" {
		serviceName = "keel-backend"
	}

	serviceVersion := configmanager.GetInstance().ServiceVersion

	// Create Jaeger exporter
	jaegerEndpoint := configmanager.GetInstance().JaegerEndpoint
	var (
		exporter sdktrace.SpanExporter
		err      error
	)

	exporter, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))

	if err != nil {
		t.logger.Error("Failed to create exporter", zap.Error(err))
		return
	}

	// Create resource with service information
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion.String()),
			attribute.String("environment", configmanager.GetInstance().DeploymentEnv),
		),
	)
	if err != nil {
		t.logger.Error("Failed to create resource, using no-op tracer", zap.Error(err))

		// Create a no-op tracer provider as fallback
		t.tp = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.NeverSample()),
		)
		otel.SetTracerProvider(t.tp)
		t.tracer = otel.Tracer(serviceName)

		t.logger.Warn("Tracing initialized with no-op tracer (resource creation failed)",
			zap.String("service_name", serviceName),
			zap.String("service_version", serviceVersion.String()))
		return
	}

	// Create trace provider
	t.tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global trace provider
	otel.SetTracerProvider(t.tp)

	// Create tracer
	t.tracer = otel.Tracer(serviceName)

	t.logger.Info("Tracing initialized successfully",
		zap.String("service_name", serviceName),
		zap.String("service_version", serviceVersion.String()),
		zap.String("jaeger_endpoint", jaegerEndpoint),
	)
}

// Shutdown gracefully shuts down the trace provider
func (t *Tracing) Shutdown(ctx context.Context) error {
	if t.tp != nil {
		return t.tp.Shutdown(ctx)
	}
	return nil
}

// TraceHTTPRequest creates a span for HTTP request tracing
func (t *Tracing) TraceHTTPRequest(ctx context.Context, method, path, status string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, fmt.Sprintf("HTTP %s %s", method, path),
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.route", path),
			attribute.String("http.status_code", status),
		),
	)
}

// TraceDBOperation creates a span for database operation tracing
func (t *Tracing) TraceDBOperation(ctx context.Context, operation, collection string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, fmt.Sprintf("DB %s %s", operation, collection),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.operation", operation),
			attribute.String("db.collection", collection),
		),
	)
}

// TraceBusinessOperation creates a span for business operation tracing
func (t *Tracing) TraceBusinessOperation(ctx context.Context, operation, entityType, entityID string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, operation,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("business.operation", operation),
			attribute.String("business.entity_type", entityType),
			attribute.String("business.entity_id", entityID),
		),
	)
}

// TraceAPICall creates a span for API call tracing
func (t *Tracing) TraceAPICall(ctx context.Context, endpoint, method string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, fmt.Sprintf("API %s %s", method, endpoint),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("api.endpoint", endpoint),
			attribute.String("api.method", method),
		),
	)
}

// TraceCacheOperation creates a span for cache operation tracing
func (t *Tracing) TraceCacheOperation(ctx context.Context, operation, cacheType, key string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, fmt.Sprintf("Cache %s", operation),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("cache.operation", operation),
			attribute.String("cache.type", cacheType),
			attribute.String("cache.key", key),
		),
	)
}

// TraceRateLimit creates a span for rate limiting tracing
func (t *Tracing) TraceRateLimit(ctx context.Context, clientID, path string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, "Rate Limit Check",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("rate_limit.client_id", clientID),
			attribute.String("rate_limit.path", path),
		),
	)
}

// RecordError records an error in the current span
func (t *Tracing) RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// AddEvent adds an event to the current span
func (t *Tracing) AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttributes sets attributes on the current span
func (t *Tracing) SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// GetTraceID returns the trace ID from the context
func (t *Tracing) GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	return span.SpanContext().TraceID().String()
}

// GetSpanID returns the span ID from the context
func (t *Tracing) GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	return span.SpanContext().SpanID().String()
}

// IsSampled returns whether the current span is sampled
func (t *Tracing) IsSampled(ctx context.Context) bool {
	span := trace.SpanFromContext(ctx)
	return span.SpanContext().IsSampled()
}

// InjectTraceContext injects trace context into HTTP headers
func (t *Tracing) InjectTraceContext(ctx context.Context, headers map[string]string) {
	span := trace.SpanFromContext(ctx)

	// Inject trace context into headers
	headers["X-Trace-ID"] = span.SpanContext().TraceID().String()
	headers["X-Span-ID"] = span.SpanContext().SpanID().String()
}

// ExtractTraceContext extracts trace context from HTTP headers
func (t *Tracing) ExtractTraceContext(ctx context.Context, headers map[string]string) context.Context {
	traceID := headers["X-Trace-ID"]
	spanID := headers["X-Span-ID"]

	if traceID != "" && spanID != "" {
		// Create a new span context with the extracted trace and span IDs
		// This is a simplified version - in production, you'd use proper trace context propagation
		t.logger.Debug("Extracted trace context",
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
		)
	}

	return ctx
}

// TraceWithTimeout creates a span with a timeout
func (t *Tracing) TraceWithTimeout(ctx context.Context, name string, timeout time.Duration) (context.Context, trace.Span, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	spanCtx, span := t.tracer.Start(ctx, name)

	return spanCtx, span, cancel
}

// TraceWithDeadline creates a span with a deadline
func (t *Tracing) TraceWithDeadline(ctx context.Context, name string, deadline time.Time) (context.Context, trace.Span, context.CancelFunc) {
	ctx, cancel := context.WithDeadline(ctx, deadline)
	spanCtx, span := t.tracer.Start(ctx, name)

	return spanCtx, span, cancel
}

// GetTracer returns the underlying tracer
func (t *Tracing) GetTracer() trace.Tracer {
	return t.tracer
}

// GetTracerProvider returns the underlying tracer provider
func (t *Tracing) GetTracerProvider() *sdktrace.TracerProvider {
	return t.tp
}

// IsInitialized returns whether tracing is properly initialized
func (t *Tracing) IsInitialized() bool {
	return t.tp != nil && t.tracer != nil
}
