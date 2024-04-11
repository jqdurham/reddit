package logger

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

type customLogger struct{}

func (c *customLogger) Info(_ string, _ ...any)  {}
func (c *customLogger) Debug(_ string, _ ...any) {}
func (c *customLogger) Warn(_ string, _ ...any)  {}
func (c *customLogger) Error(_ string, _ ...any) {}

func TestFromContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  context.Context
		want Logger
	}{
		{
			name: "Return custom logger when exists",
			ctx:  context.WithValue(context.Background(), loggerKey, &customLogger{}),
			want: &customLogger{},
		},
		{
			name: "Returns default logger when custom logger not exists",
			ctx:  context.Background(),
			want: slog.Default(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := FromContext(tt.ctx)

			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestNewContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  context.Context
		logr Logger
		want context.Context
	}{
		{
			name: "Stores logger in provided context",
			ctx:  context.Background(),
			logr: &customLogger{},
			want: context.WithValue(context.Background(), loggerKey, &customLogger{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewContext(tt.ctx, tt.logr)

			assert.EqualValues(t, tt.want, got)
		})
	}
}
