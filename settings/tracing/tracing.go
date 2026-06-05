package tracing

import (
	"context"
	"sync"

	"github.com/glodb/keel/settings/configmanager"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

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

func GetInstance() *Tracing {
	once.Do(func() {
		instance = newTracing()
	})
	return instance
}

func newTracing() *Tracing {
	t := &Tracing{logger: zap.L()}
	t.initialize(configmanager.GetInstance().ServiceLBName, configmanager.GetInstance().JaegerEndpoint, configmanager.GetInstance().ServiceVersion.String(), configmanager.GetInstance().DeploymentEnv)
	return t
}

func NewTracing(serviceName string, jaegerEndpoint string, version string, environment string) *Tracing {
	t := &Tracing{logger: zap.L()}
	t.initialize(serviceName, jaegerEndpoint, version, environment)
	return t
}

func (t *Tracing) initialize(serviceName string, jaegerEndpoint string, version string, environment string) {
	if serviceName == "" {
		serviceName = "keel-backend"
	}

	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		t.logger.Error("Failed to create Jaeger exporter, using no-op tracer", zap.Error(err))
		t.tp = sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.NeverSample()))
		otel.SetTracerProvider(t.tp)
		t.tracer = otel.Tracer(serviceName)
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
		t.tp = sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.NeverSample()))
		otel.SetTracerProvider(t.tp)
		t.tracer = otel.Tracer(serviceName)
		return
	}

	t.tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(t.tp)
	t.tracer = otel.Tracer(serviceName)

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
