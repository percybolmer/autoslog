package autoslog

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
)

const (
	TEXT_HANDLER = iota
	JSON_HANDLER
)

type Logger struct {
	handler    slog.Handler
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
	return Logger{handler: hand}
}

// WithMiddleware is used to add a middleware to a logger
func (l Logger) WithMiddleware(mw MiddlewareFunc) Logger {
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

func (l Logger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.handler.Enabled(ctx, level)
}

func (l *Logger) Handle(ctx context.Context, record slog.Record) error {
	msg := record.Message
	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	for _, mw := range l.middleware {
		msg, attrs = mw(l.ctx, msg, attrs...)
	}

	newRecord := slog.NewRecord(record.Time, record.Level, msg, record.PC)
	for _, attr := range attrs {
		newRecord.AddAttrs(attr)
	}

	return l.handler.Handle(ctx, newRecord)
}

// log is a generic log func that simply logs the message
func (l Logger) log(level slog.Level, msg string, attrs ...slog.Attr) {
	record := slog.NewRecord(time.Now(), level, msg, 0)
	record.AddAttrs(attrs...)
	l.Handle(l.ctx, record)
}

// Logging methods for different levels using the generic log method
func (l Logger) Info(msg string, attrs ...slog.Attr) {
	l.log(slog.LevelInfo, msg, attrs...)
}

func (l Logger) Debug(msg string, attrs ...slog.Attr) {
	l.log(slog.LevelDebug, msg, attrs...)
}

func (l Logger) Warn(msg string, attrs ...slog.Attr) {
	l.log(slog.LevelWarn, msg, attrs...)
}

func (l Logger) Error(msg string, err error, attrs ...slog.Attr) {
	attrs = append(attrs, slog.Any("error", err))
	l.log(slog.LevelError, msg, attrs...)
}

func (l Logger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Logger{
		handler: l.handler.WithAttrs(attrs),
		ctx:     l.ctx,
	}
}

func (l Logger) WithGroup(name string) slog.Handler {
	return &Logger{
		handler: l.handler.WithGroup(name),
		ctx:     l.ctx,
	}
}
