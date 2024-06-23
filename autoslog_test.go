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

func TestAutoSlog_Handler(t *testing.T) {
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

	// Create a slog.Logger using the custom logger as handler
	slogLogger := slog.New(&logger)

	// Create a context with the custom logger
	ctx = AddLoggerToContext(context.Background(), logger)

	// Start a span and add it to the context
	ctx, span := tracer.Start(ctx, "test-span")
	defer span.End()

	// Log messages using slog.Logger
	slogLogger.InfoContext(ctx, "Test info message")
	slogLogger.ErrorContext(ctx, "Test error message", slog.Any("error", errors.New("test error")))

	// Check the output for attributes and messages
	logOutput := buf.String()
	t.Log(logOutput)

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

	if retrievedLogger.handler == nil {
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
