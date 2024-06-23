package autoslog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestLogger_Middlewares(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	ctx := context.Background()

	tracer := InitTracer(ctx, "test-service")
	// Create a logger that writes to the buffer
	logger := NewLogger(&buf, TEXT_HANDLER, []slog.Attr{
		slog.String("service.name", "test-service"),
	}).
		WithContext(ctx).
		WithMiddleware(AutoTracing(tracer))

	// Log messages
	logger.Error("Test error message", errors.New("test error"))

	// Check the output for attributes and messages
	logOutput := buf.String()

	if !contains(logOutput, "Test error message") {
		t.Errorf("Expected log message to contain 'Test error message', got: %s", logOutput)
	}

	if !contains(logOutput, "test error") {
		t.Errorf("Expected log message to contain 'test error', got: %s", logOutput)
	}

	// make sure Trace and SPAN id are set
	if !contains(logOutput, "TraceID") {
		t.Errorf("Expected log message to attribute 'TraceID', got: %s", logOutput)
	}

	// Set trace and SPAN into Predetermined IDs and apply on context and make sure the attributes are added correctly
	// Create a context with predetermined trace and span IDs
	traceID := trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
	ctx = trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	}))

	// Override logger with new Context
	buf.Reset()
	logger.WithContext(ctx).WithSpan("test-span").Info("new log with predetermined ID")

	logOutput = buf.String()

	// Check for TraceID and SpanID
	if !contains(logOutput, "TraceID=0102030405060708090a0b0c0d0e0f10") {
		t.Errorf("Expected log message to contain 'TraceID=0102030405060708090a0b0c0d0e0f10', got: %s", logOutput)
	}

	if !contains(logOutput, "SpanID=0102030405060708") {
		t.Errorf("Expected log message to contain 'SpanID=0102030405060708', got: %s", logOutput)
	}
}

// TestAutoEnvironmentMiddleware tests the AutoEnvironment middleware
func TestLogger_Middlewares_AutoEnvironment(t *testing.T) {
	var buf bytes.Buffer

	// Create a new logger with the AutoEnvironment middleware
	logger := NewLogger(&buf, TEXT_HANDLER, []slog.Attr{
		slog.String("dont-remove", "test-dont-remove")},
	).WithMiddleware(AutoEnvironment("test-service", "test-environment", "test-host"))

	// Log a message
	logger.Info("Test log message")

	// Check the log output
	logOutput := buf.String()

	// Verify that the log output contains the expected attributes
	expectedAttributes := []string{
		fmt.Sprintf("%s=test-service", string(CTX_SERVICE_NAME)),
		fmt.Sprintf("%s=test-environment", string(CTX_ENVIRONMENT)),
		fmt.Sprintf("%s=test-host", string(CTX_HOST_NAME)),
		fmt.Sprintf("%s=test-dont-remove", "dont-remove"),
	}

	for _, attr := range expectedAttributes {
		if !bytes.Contains([]byte(logOutput), []byte(attr)) {
			t.Errorf("Expected log output to contain '%s', got: %s", attr, logOutput)
		}
	}
}
