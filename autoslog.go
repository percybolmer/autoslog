package autoslog

import (
	"context"
	"io"
	"log/slog"
	"os"
)

const (
	TEXT_HANDLER = iota
	JSON_HANDLER
)
const (
	CTX_SPAN_NAME context_key = "span_name"
	LOGGER_KEY    context_key = "logger"
)

type context_key string

type Logger struct {
	logger     *slog.Logger
	middleware []MiddlewareFunc
	ctx        context.Context
}

// NewLogger creates a new logger with the attributes applied and the Handler selected
// The handler should be one of the predefined consts in this package
func NewLogger(output io.Writer, handler int, defaultAttributes []slog.Attr) Logger {
	if output == nil {
		output = os.Stdout
	}

	var hand slog.Handler
	switch handler {
	case 0:
		hand = slog.NewTextHandler(output, &slog.HandlerOptions{}).WithAttrs(defaultAttributes)
	case 1:
		hand = slog.NewJSONHandler(output, &slog.HandlerOptions{}).WithAttrs(defaultAttributes)
	default:
		hand = slog.NewJSONHandler(output, &slog.HandlerOptions{}).WithAttrs(defaultAttributes)
	}
	return Logger{logger: slog.New(hand)}
}

// NewLoggerWithHandler is a wrapper to allow custom Handlers
func NewLoggerWithHandler(handler slog.Handler) Logger {
	return Logger{logger: slog.New(handler)}
}

// With is used to add a middleware to a logger
func (l Logger) With(mw MiddlewareFunc) Logger {
	l.middleware = append(l.middleware, mw)
	return l
}

// WithContext adds a context to the logger
func (l Logger) WithContext(ctx context.Context) Logger {
	l.ctx = ctx
	return l
}

// WithSpan adds a span
func (l Logger) WithSpan(span string) Logger {
	if l.ctx == nil {
		l.ctx = context.Background()
	}
	l.ctx = context.WithValue(l.ctx, CTX_SPAN_NAME, span)
	return l
}

// Info is used to print Info logs, will also trigger Middlewares
func (l Logger) Info(msg string, args ...any) {
	for _, mw := range l.middleware {
		msg, args = mw(l.ctx, msg, args...)
	}
	l.logger.Info(msg, args...)
}

// Debug is used to print Debug logs, will also trigger Middlewares
func (l Logger) Debug(msg string, args ...any) {
	for _, mw := range l.middleware {
		msg, args = mw(l.ctx, msg, args...)
	}
	l.logger.Debug(msg, args...)
}

// Warn is used to print Warn logs, will also trigger Middlewares
func (l Logger) Warn(msg string, args ...any) {
	for _, mw := range l.middleware {
		msg, args = mw(l.ctx, msg, args...)
	}
	l.logger.Warn(msg, args...)
}

// Error will print a error log message, will add the "error" attribute with the err msg
func (l Logger) Error(msg string, err error, args ...any) {
	for _, mw := range l.middleware {
		msg, args = mw(l.ctx, msg, args...)
	}
	args = append(args, "error", err)
	l.logger.Error(msg, args...)
}

// AddLoggerToContext adds the logger to the provided context
func AddLoggerToContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, LOGGER_KEY, logger)
}

// FromContext retrieves the logger from the provided context
func FromContext(ctx context.Context) (Logger, bool) {
	logger, ok := ctx.Value(LOGGER_KEY).(Logger)
	return logger, ok
}
