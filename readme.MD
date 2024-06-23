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

## Pass By Structs
This approach is when you embed the logger into a Struct by storing it in the Struct.

This approach is often used when you need to encapsulate the logging functionality to the struct. 
It is an approach that is easy to use, and does not require each function of the struct to explicitly accept a logger.

This usually means maintaining a `Stateful` logger in the struct.

The negative part of this approach is the tight coupling between the struct and the logging.

This is suitable when you have objects that has a long lifespan and need consistent logging.

An example of how to perform this

```go
package main

import (
	"github.com/percybolmer/autoslog"
	"os"
)

type Service struct {
	logger autoslog.Logger
}

func NewService() *Service {
	logger := autoslog.NewLogger(os.Stdout, autoslog.TEXT_HANDLER, nil).
		WithMiddleware(autoslog.AutoEnvironment("my-service", "production", "localhost"))

	return &Service{
		logger: logger,
	}
}

func (s *Service) DoSomething() {
	s.logger.Info("Service is doing something")
}

func main() {
	service := NewService()
	service.DoSomething()
}
```

## Pass By Value
This approach is when every function accepts the Logger as a parameter.
Pretty much the same way we in the Go community use `ctx context.Context` as a first parameter. 

This allows logging without embedding the logger into a struct. 

Pros is that you dont need a tightly coupled struct, and functions without being attached to a struct can use this.

It is also very clear what functions that will depend on logging.

The negative part is that it is very verbose and repetetive to pass the logger.

```go
package main

import (
	"github.com/percybolmer/autoslog"
	"os"
)

func DoSomething(logger autoslog.Logger) {
	logger.Info("Doing something important")
}

func main() {
	logger := autoslog.NewLogger(os.Stdout, autoslog.TEXT_HANDLER, nil).
		WithMiddleware(autoslog.AutoEnvironment("my-service", "production", "localhost"))

	DoSomething(logger)
}

```

## Pass by Context
Pass by Context is an approach where we embedd the logger in a `context.Context` and then the functions that needs logging can fetch the logger and use it. 
This approach is good when a context is already passing through the application.

This approach is useful since its easy to start using and does not require an extra parameter to be passed, making it less bloated. 
It also easy to start using in functions that has access to a context.

A negative thing about it is that it creates a implicit dependency. Its no longer super clea that the function depends on a logger. 

It also increase context pollution if you use the context much, not sure if this really is an issue though.

It is however error prone, you can easily forget to add the logger and that can if not aken into account create risky behavior.

Very suitable for request-scoped services such as web services or microservices.

Autoslog tries to remove some of that risk by impementing a risk free approach, if there is no logger we will simply not log anything.

The pass by Context approach is usually recommended for large scalable applications

```go
package main

import (
	"context"
	"errors"
	"github.com/percybolmer/autoslog"
	"go.opentelemetry.io/otel/trace"
	"os"
)

func main() {
	tracer := trace.NewNoopTracerProvider().Tracer("example")

	// Create a new logger
	logger := autoslog.NewLogger(os.Stdout, autoslog.TEXT_HANDLER, nil).
		WithMiddleware(autoslog.TraceMiddleware(tracer)).
		WithMiddleware(autoslog.AutoEnvironment("my-service", "production", "localhost"))

	// Create a context with the logger
	ctx := autoslog.AddLoggerToContext(context.Background(), logger)

	// Start a span and add it to the context
	ctx, span := tracer.Start(ctx, "main-span")
	defer span.End()

	// Pass context to functions
	doSomething(ctx)
	doSomethingElse(ctx)
}

func doSomething(ctx context.Context) {
	autoslog.InfoCtx(ctx, "Doing something important")
	err := errors.New("something went wrong")
	autoslog.ErrorCtx(ctx, "An error occurred", err)
}

func doSomethingElse(ctx context.Context) {
	autoslog.WarnCtx(ctx, "This is a warning")
	autoslog.DebugCtx(ctx, "Debugging information")
}

```

## Log at Top Level
One common recommendation in the Go Community is to always log at the top of your application. 




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