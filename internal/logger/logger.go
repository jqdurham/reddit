package logger

import (
	"context"
	"log/slog"
)

// Logger defines a logging interface for this project.
//
//go:generate mockery --name Logger
type Logger interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type key int

const (
	loggerKey key = iota
)

// NewContext creates a context with the provided logger inside.
func NewContext(ctx context.Context, logr Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logr)
}

// FromContext fetches logger from provided context. If a logger doesn't exist, it returns the DefaultLogger.
//
//nolint:ireturn // NewContext sets an interface and this needs to return it.
func FromContext(ctx context.Context) Logger {
	if logr := ctx.Value(loggerKey); logr != nil {
		if v, ok := logr.(Logger); ok {
			return v
		}
	}

	return slog.Default()
}
