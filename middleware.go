package autoslog

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// MiddlewareFunc is a alias for a function that can be used as a Middleware between each log call
type MiddlewareFunc func(ctx context.Context, msg string, args ...any) (string, []any)

// TimestampMiddleware returns a middleware that formats the default "time" attribute using the provided format.
func TimestampMiddleware(format string) MiddlewareFunc {
	return func(ctx context.Context, msg string, args ...any) (string, []any) {
		timestamp := slog.String("time", time.Now().Format(format))
		return msg, append(args, timestamp)
	}
}

// TraceMiddleware ensures that TRACEID and SpanID are present in the context and adds them to the log attributes.
// This assumes you are using OTEL for tracing
func TraceMiddleware(tracer trace.Tracer) MiddlewareFunc {
	return func(ctx context.Context, msg string, args ...any) (string, []any) {
		span := trace.SpanFromContext(ctx)
		var traceID, spanID string

		// Retrieve the span name from the context if available
		spanName, _ := ctx.Value(CTX_SPAN_NAME).(string)
		if spanName == "" {
			spanName = "auto-generated-span"
		}

		if span.SpanContext().IsValid() {
			traceID = span.SpanContext().TraceID().String()
			spanID = span.SpanContext().SpanID().String()
		} else {
			// Create a new span to generate valid trace and span IDs
			_, span = tracer.Start(ctx, spanName)
			defer span.End()
			traceID = span.SpanContext().TraceID().String()
			spanID = span.SpanContext().SpanID().String()
		}

		return msg, append(args, slog.String("trace_id", traceID), slog.String("span_id", spanID))
	}
}
