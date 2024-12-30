package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Layer represents a traceable layer in the application
type Layer string

const (
	HTTPLayer       Layer = "http"
	ServiceLayer    Layer = "service"
	DatabaseLayer   Layer = "database"
	ControllerLayer Layer = "controller"
	RepositoryLayer Layer = "repository"
)

// TracerProvider encapsulates the tracing functionality
type TracerProvider struct {
	tracer trace.Tracer
}

// NewTracerProvider creates a new TracerProvider
func NewTracerProvider(serviceName string) *TracerProvider {
	return &TracerProvider{
		tracer: Tracer,
	}
}

// StartSpan starts a new span with layer-specific attributes
func (tp *TracerProvider) StartSpan(ctx context.Context, layer Layer, operation string) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("%s.%s", layer, operation)

	ctx, span := tp.tracer.Start(ctx, spanName,
		trace.WithAttributes(
			attribute.String("layer", string(layer)),
			attribute.String("operation", operation),
		),
	)

	return ctx, span
}
