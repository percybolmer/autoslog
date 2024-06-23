package autoslog

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
)

func TestAutoSlog_CtxUtil(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	ctx := context.Background()
	// Initialize tracer
	tracer := InitTracer(ctx, "test-service")
	// Create a logger that writes to the buffer
	logger := NewLogger(&buf, TEXT_HANDLER, []slog.Attr{
		slog.String("service.name", "test-service"),
	}).
		WithContext(ctx).
		WithMiddleware(AutoTracing(tracer))

	// Create a context with the custom logger
	ctx = AddLoggerToContext(context.Background(), logger)

	// Log messages using slog.Logger
	InfoCtx(ctx, "Test info message")
	ErrorCtx(ctx, "Test error message", errors.New("test error"))

	// Check the output for attributes and messages
	logOutput := buf.String()

	if !contains(logOutput, "Test info message") {
		t.Errorf("Expected log message to contain 'Test info message', got: %s", logOutput)
	}

	if !contains(logOutput, "test error") {
		t.Errorf("Expected log message to contain 'test error', got: %s", logOutput)
	}

	if !contains(logOutput, "TraceID") {
		t.Errorf("Expected log message to contain 'TraceID', got: %s", logOutput)
	}

	if !contains(logOutput, "SpanID") {
		t.Errorf("Expected log message to contain 'SpanID', got: %s", logOutput)
	}
}
