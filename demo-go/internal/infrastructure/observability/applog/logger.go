package applog

import (
	"context"
	"log/slog"
	"os"

	"dag-observatory/demo-go/internal/domain/observability/ctxprop"
	"dag-observatory/demo-go/internal/domain/observability/semantics"

	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	base *slog.Logger
}

func New() *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return &Logger{base: slog.New(handler)}
}

func (l *Logger) Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, slog.LevelInfo, msg, attrs...)
}

func (l *Logger) Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, slog.LevelWarn, msg, attrs...)
}

func (l *Logger) Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.log(ctx, slog.LevelError, msg, attrs...)
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	attrs = append(attrs, traceAttrs(ctx)...)
	if runID, ok := ctxprop.RunID(ctx); ok {
		attrs = append(attrs, slog.String(semantics.KeyDAGRunID, runID))
	}
	l.base.LogAttrs(ctx, level, msg, attrs...)
}

func traceAttrs(ctx context.Context) []slog.Attr {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return nil
	}
	return []slog.Attr{
		slog.String("trace_id", spanCtx.TraceID().String()),
		slog.String("span_id", spanCtx.SpanID().String()),
	}
}
