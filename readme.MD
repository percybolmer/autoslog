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
		With(TraceMiddleware(tracer))
```

You can create a default logger like that. If you wish to modify behavior in a single part 
of the code, you can just make to simply not override your default logger.

The functions all return a replica, not pointers so you can chain items how much you want.

## Attributes

## Context Logging

## Tracing
You can add automatic tracing to logs by using the `TraceMiddleware` found in the `middleware.go`.
This will automatically add trace ID and span ID to any Log that DOES NOT contain it.
It the context does contain the values, it will add those instead of generating new ones.

If you wish to not use a generated span name, you can control that with `WithSpan`

```go
	logger.WithContext(ctx).WithSpan("test-span").Info("new log with predetermined ID")
```