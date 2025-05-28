package tenantcontext

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// PropagateToSpan adds tenant information to the current span
func PropagateToSpan(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	tenant, ok := GetTenant(ctx)
	if !ok {
		return
	}

	// Add tenant attributes to the span
	span.SetAttributes(
		attribute.String("tenant.id", tenant.ID),
		attribute.String("tenant.name", tenant.Name),
		attribute.Bool("tenant.is_active", tenant.IsActive),
	)

	// Add datasource count if available
	if len(tenant.Datasources) > 0 {
		span.SetAttributes(attribute.Int("tenant.datasources.count", len(tenant.Datasources)))
	}
}

// WithTracing wraps a function with tenant context propagation to tracing
func WithTracing(ctx context.Context, fn func(context.Context) error) error {
	PropagateToSpan(ctx)
	return fn(ctx)
}

// WithTracingFunc wraps a function with tenant context propagation to tracing (generic version)
func WithTracingFunc[T any](ctx context.Context, fn func(context.Context) (T, error)) (T, error) {
	PropagateToSpan(ctx)
	return fn(ctx)
}

// StartSpanWithTenant starts a new span and automatically adds tenant information
func StartSpanWithTenant(ctx context.Context, tracer trace.Tracer, spanName string) (context.Context, trace.Span) {
	ctx, span := tracer.Start(ctx, spanName)
	PropagateToSpan(ctx)
	return ctx, span
}
