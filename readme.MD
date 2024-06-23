# Autoslog
Autoslog is a library used to help developers easily start using `Tracing` and `Structured logs`

It will help by having being a strongly opinionated library that sets best practices for
logs and tracing.

Autoslog is built as a functional library, allowing you to chain items.

```go
	logger := NewLogger(&buf, TEXT_HANDLER, []slog.Attr{
		slog.String("service.name", "test-service"),
	}).
		WithContext(ctx).
		WithMiddleware(AutoTracing(tracer))
```

You can create a default logger like that. 

The functions all return a replica, not pointers so you can chain items how much you want.
```go
	logger := NewLogger(&buf, TEXT_HANDLER, []slog.Attr{
		slog.String("service.name", "test-service"),
	}).
		WithContext(ctx).
		WithMiddleware(AutoTracing(tracer))

	// Add Middleware only ONCE by not overriding main logger
	logger.WithMiddleware(middleware).Info("test")
```

Setting the default slog to use autoslog this way
```go
	logger := NewLogger(&buf, TEXT_HANDLER, []slog.Attr{
		slog.String("service.name", "test-service"),
	}).
		WithContext(ctx).
		WithMiddleware(AutoTracing(tracer))

	// Create a slog.Logger using the custom logger as handler
	slogLogger := slog.New(&logger)
	slog.SetDefault(slogLogger)
```

## Attributes
You can add attributes that are defaults when you create the logger in the third parameter.

```go
logger := NewLogger(&buf, TEXT_HANDLER, []slog.Attr{
		slog.String("service.name", "test-service"),
})
```

There is also middlewares that adds some defaults depending on your need
For instance, often you want the host, service name and maybe the Environment it runs on
```go
logger := NewLogger(&buf, TEXT_HANDLER, nil).WithMiddleware(AutoEnvironment("test-service", "staging-server", "test"))
```

## Context Logging
If you are using Pass by context for your logger, then you can leverage the help utils
that helps you easily log from that logger the following way.

```go
	logger := NewLogger(&buf, TEXT_HANDLER, []slog.Attr{
		slog.String("service.name", "test-service"),
	}).
		WithContext(ctx).
		WithMiddleware(AutoTracing(tracer))

	// Create a context with the logger
	ctx := autoslog.AddLoggerToContext(context.Background(), logger)
	// Call the Logger with the Library Log funcs instead directly using the logger
	autoslog.InfoCtx(ctx, "Doing something important")

```

## Tracing
You can add automatic tracing to logs by using the `AutoTracing` found in the `middleware.go`.
This will automatically add trace ID and span ID to any Log that DOES NOT contain it.
It the context does contain the values, it will add those instead of generating new ones.

If you wish to not use a generated span name, you can control that with `WithSpan`

```go
	logger.WithContext(ctx).WithSpan("test-span").Info("new log with predetermined ID")
```