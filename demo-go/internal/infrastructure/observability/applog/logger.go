package applog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"dag-observatory/demo-go/internal/domain/observability/ctxprop"
	"dag-observatory/demo-go/internal/domain/observability/semantics"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	base *slog.Logger
	otel log.Logger
}

func New(level, output string, otelLogger log.Logger) *Logger {
	handler := slog.NewJSONHandler(outputWriter(output), &slog.HandlerOptions{
		Level: slogLevel(level),
	})
	return &Logger{
		base: slog.New(handler),
		otel: otelLogger,
	}
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
	l.emitOTel(ctx, level, msg, attrs...)
}

func (l *Logger) emitOTel(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	if l.otel == nil {
		return
	}
	severity := severityFromLevel(level)
	if !l.otel.Enabled(ctx, log.EnabledParameters{Severity: severity}) {
		return
	}

	var record log.Record
	now := time.Now()
	record.SetTimestamp(now)
	record.SetObservedTimestamp(now)
	record.SetSeverity(severity)
	record.SetSeverityText(level.String())
	record.SetBody(log.StringValue(msg))
	record.AddAttributes(slogAttrsToLog(attrs)...)
	l.otel.Emit(ctx, record)
}

func slogAttrsToLog(attrs []slog.Attr) []log.KeyValue {
	out := make([]log.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		out = appendSlogAttr(out, attr, "")
	}
	return out
}

func appendSlogAttr(out []log.KeyValue, attr slog.Attr, prefix string) []log.KeyValue {
	if attr.Equal(slog.Attr{}) {
		return out
	}
	key := prefix + attr.Key
	switch attr.Value.Kind() {
	case slog.KindGroup:
		group := attr.Value.Group()
		if attr.Key == "" {
			for _, g := range group {
				out = appendSlogAttr(out, g, prefix)
			}
			return out
		}
		for _, g := range group {
			out = appendSlogAttr(out, g, key+".")
		}
		return out
	default:
		if key == "" {
			return out
		}
		out = append(out, slogValueToLog(key, attr.Value))
		return out
	}
}

func slogValueToLog(key string, value slog.Value) log.KeyValue {
	switch value.Kind() {
	case slog.KindString:
		return log.String(key, value.String())
	case slog.KindInt64:
		return log.Int64(key, value.Int64())
	case slog.KindUint64:
		return log.String(key, fmt.Sprintf("%d", value.Uint64()))
	case slog.KindFloat64:
		return log.Float64(key, value.Float64())
	case slog.KindBool:
		return log.Bool(key, value.Bool())
	case slog.KindDuration:
		return log.Int64(key, value.Duration().Milliseconds())
	case slog.KindTime:
		return log.String(key, value.Time().Format(time.RFC3339Nano))
	case slog.KindAny:
		if err, ok := value.Any().(error); ok {
			return log.String(key, err.Error())
		}
		if str, ok := value.Any().(fmt.Stringer); ok {
			return log.String(key, str.String())
		}
		return log.String(key, fmt.Sprint(value.Any()))
	default:
		return log.Empty(key)
	}
}

func outputWriter(output string) io.Writer {
	switch strings.ToLower(strings.TrimSpace(output)) {
	case "stderr":
		return os.Stderr
	default:
		return os.Stdout
	}
}

func slogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func severityFromLevel(level slog.Level) log.Severity {
	switch {
	case level <= slog.LevelDebug:
		return log.SeverityDebug
	case level <= slog.LevelInfo:
		return log.SeverityInfo
	case level <= slog.LevelWarn:
		return log.SeverityWarn
	default:
		return log.SeverityError
	}
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
