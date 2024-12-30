package tracer

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// JeagerContext encapsulates the tracing functionality
type JeagerContext struct {
	Endpoint string
}

// Tracer encapsulates the tracing functionality
var Tracer trace.Tracer

// NewJeagerContext creates a new JeagerContext
func (j JeagerContext) Open(servicesName string) (*sdktrace.TracerProvider, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(j.Endpoint)))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(servicesName),
		)),
	)

	otel.SetTracerProvider(tp)

	return tp, nil
}
