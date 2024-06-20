package autoslog

import (
	"bytes"
	"context"
	"errors"
	"log"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
)

func TestLogger_Attributes(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a logger that writes to the buffer
	logger := NewLogger(&buf, TEXT_HANDLER, []slog.Attr{
		slog.String("service.name", "test-service"),
	})

	// Create a context and test applying it
	ctx := context.Background()
	logger = logger.WithContext(ctx)

	// Log messages
	logger.Info("Test info message")
	logger.Error("Test error message", errors.New("test error"))

	// Check the output for attributes and messages
	logOutput := buf.String()

	if !contains(logOutput, "Test info message") {
		t.Errorf("Expected log message to contain 'Test info message', got: %s", logOutput)
	}
	if !contains(logOutput, "test-service") {
		t.Errorf("Expected log message to contain attribute 'service.name', got: %s", logOutput)
	}
	if !contains(logOutput, "Test error message") {
		t.Errorf("Expected log message to contain 'Test error message', got: %s", logOutput)
	}

	if !contains(logOutput, "test error") {
		t.Errorf("Expected log message to contain 'test error', got: %s", logOutput)
	}
}

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
		With(TraceMiddleware(tracer))

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
	if !contains(logOutput, "trace_id") {
		t.Errorf("Expected log message to attribute 'trace_id', got: %s", logOutput)
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

	// Check for trace_id and span_id
	if !contains(logOutput, "trace_id=0102030405060708090a0b0c0d0e0f10") {
		t.Errorf("Expected log message to contain 'trace_id=0102030405060708090a0b0c0d0e0f10', got: %s", logOutput)
	}

	if !contains(logOutput, "span_id=0102030405060708") {
		t.Errorf("Expected log message to contain 'span_id=0102030405060708', got: %s", logOutput)
	}
}

func TestAddLoggerToContext(t *testing.T) {
	// Create a new logger instance
	logger := NewLogger(nil, TEXT_HANDLER, nil)

	// Create a context and add the logger to it
	ctx := AddLoggerToContext(context.Background(), logger)

	// Retrieve the logger from the context
	retrievedLogger, ok := FromContext(ctx)
	if !ok {
		t.Errorf("Expected logger to be found in context")
	}

	if retrievedLogger.logger == nil {
		t.Errorf("failed to grab the logger")
	}
}

func TestFromContext_LoggerNotFound(t *testing.T) {
	// Create a context without adding a logger
	ctx := context.Background()

	// Attempt to retrieve the logger from the context
	_, ok := FromContext(ctx)
	if ok {
		t.Errorf("Expected no logger to be found in context")
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// Init returns an instance of Jaeger Tracer.
func InitTracer(ctx context.Context, service string) trace.Tracer {
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
	)
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		log.Fatal("creating OTLP trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(newResource(service)),
	)

	return tp.Tracer(service)
}

func newResource(service string) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(service),
		semconv.ServiceVersion("0.0.1"),
	)
}
