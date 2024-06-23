package autoslog

import (
	"context"
	"log/slog"
	"os"
)

const (
	CTX_SPAN_NAME    context_key = "SpanName"
	CTX_TRACE_ID     context_key = "TraceID" // same as OTEL
	CTX_SPAN_ID      context_key = "SpanID"  // same as OTEL
	CTX_REQUEST_ID   context_key = "RequestID"
	CTX_USER_ID      context_key = "UserID"
	CTX_HOST_NAME    context_key = "HostName"
	CTX_ENVIRONMENT  context_key = "Environment"
	CTX_SERVICE_NAME context_key = "ServiceName"
	LOGGER_KEY       context_key = "Logger"
)

type context_key string

// AddLoggerToContext adds the logger to the provided context
func AddLoggerToContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, LOGGER_KEY, logger)
}

// FromContext retrieves the logger from the provided context
func FromContext(ctx context.Context) (Logger, bool) {
	logger, ok := ctx.Value(LOGGER_KEY).(Logger)
	return logger, ok
}

// defaultLogger is used when no logger is found in the context
var defaultLogger = NewLogger(os.Stdout, JSON_HANDLER, nil)

// InfoCtx logs an info message using the logger from the context, or the default logger if none is found
func InfoCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	logger, ok := FromContext(ctx)
	if !ok {
		logger = defaultLogger
	}
	logger.Info(msg, attrs...)
}

// DebugCtx logs a debug message using the logger from the context, or the default logger if none is found
func DebugCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	logger, ok := FromContext(ctx)
	if !ok {
		logger = defaultLogger
	}
	logger.Debug(msg, attrs...)
}

// WarnCtx logs a warning message using the logger from the context, or the default logger if none is found
func WarnCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	logger, ok := FromContext(ctx)
	if !ok {
		logger = defaultLogger
	}
	logger.Warn(msg, attrs...)
}

// ErrorCtx logs an error message using the logger from the context, or the default logger if none is found
func ErrorCtx(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	logger, ok := FromContext(ctx)
	if !ok {
		logger = defaultLogger
	}
	logger.Error(msg, err, attrs...)
}
