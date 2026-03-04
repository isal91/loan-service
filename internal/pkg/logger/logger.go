package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/google/uuid"
)

type contextKey string

const TraceIDKey contextKey = "trace_id"
const SpanIDKey contextKey = "span_id"

var globalLogger *slog.Logger

func Init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	globalLogger = slog.New(handler)
	slog.SetDefault(globalLogger)
}

func FromContext(ctx context.Context) *slog.Logger {
	l := globalLogger
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		l = l.With(slog.String("trace_id", traceID))
	}
	if spanID, ok := ctx.Value(SpanIDKey).(string); ok {
		l = l.With(slog.String("span_id", spanID))
	}
	return l
}

func StartSpan(ctx context.Context, name string) (context.Context, *slog.Logger) {
	spanID := uuid.New().String()[:16]
	newCtx := context.WithValue(ctx, SpanIDKey, spanID)
	l := FromContext(newCtx).With(slog.String("span_name", name))
	l.Debug("span started")
	return newCtx, l
}

func Error(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Error(msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Warn(msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Info(msg, args...)
}
