package autoslog

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// MiddlewareFunc is a alias for a function that can be used as a Middleware between each log call
type MiddlewareFunc func(ctx context.Context, msg string, attrs ...slog.Attr) (string, []slog.Attr)

// TimestampMiddleware returns a middleware that formats the default "time" attribute using the provided format.
func TimestampMiddleware(format string) MiddlewareFunc {
	return func(ctx context.Context, msg string, attrs ...slog.Attr) (string, []slog.Attr) {
		timestamp := slog.String("time", time.Now().Format(format))
		return msg, append(attrs, timestamp)
	}
}

// AutoEnvironment adds environmental attributes automatically to logger
func AutoEnvironment(serviceName, environment, hostName string) MiddlewareFunc {
	return func(ctx context.Context, msg string, attrs ...slog.Attr) (string, []slog.Attr) {
		defaultAttrs := []slog.Attr{
			slog.String(string(CTX_SERVICE_NAME), serviceName),
			slog.String(string(CTX_HOST_NAME), hostName),
			slog.String(string(CTX_ENVIRONMENT), environment),
		}
		return msg, append(attrs, defaultAttrs...)
	}
}

// AutoTracing ensures that TRACEID and SpanID are present in the context and adds them to the log attributes.
func AutoTracing(tracer trace.Tracer) MiddlewareFunc {
	return func(ctx context.Context, msg string, attrs ...slog.Attr) (string, []slog.Attr) {
		span := trace.SpanFromContext(ctx)
		var traceID, spanID string

		spanName, _ := ctx.Value(CTX_SPAN_NAME).(string)
		if spanName == "" {
			spanName = "auto-generated-span"
		}

		if span.SpanContext().IsValid() {
			traceID = span.SpanContext().TraceID().String()
			spanID = span.SpanContext().SpanID().String()
		} else {
			_, span = tracer.Start(ctx, spanName)
			defer span.End()
			traceID = span.SpanContext().TraceID().String()
			spanID = span.SpanContext().SpanID().String()
		}

		return msg, append(attrs, slog.String(string(CTX_TRACE_ID), traceID), slog.String(string(CTX_SPAN_ID), spanID))
	}
}
