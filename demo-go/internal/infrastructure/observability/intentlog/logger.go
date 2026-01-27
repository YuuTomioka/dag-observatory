package intentlog

import (
	"context"
	"fmt"
	"time"

	"dag-observatory/demo-go/internal/domain/observability/semantics"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	lp          *sdklog.LoggerProvider
	serviceName string
}

func New(lp *sdklog.LoggerProvider, serviceName string) *Logger {
	return &Logger{lp: lp, serviceName: serviceName}
}

func (l *Logger) ClockTickReceived(ctx context.Context, e semantics.ClockTickReceived) {
	requireValid(semantics.EventClockTickReceived, e.Validate())
	l.emit(ctx, semantics.EventClockTickReceived,
		log.String(semantics.KeyClockUTC, e.ClockUTC.UTC().Format(time.RFC3339Nano)),
		log.String(semantics.KeySymbol, e.Symbol),
	)
}

func (l *Logger) DAGRunStarted(ctx context.Context, e semantics.DAGRunStarted) {
	requireValid(semantics.EventDAGRunStarted, e.Validate())
	l.emit(ctx, semantics.EventDAGRunStarted,
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeySymbol, e.Symbol),
	)
}

func (l *Logger) DAGRunFinished(ctx context.Context, e semantics.DAGRunFinished) {
	requireValid(semantics.EventDAGRunFinished, e.Validate())
	l.emit(ctx, semantics.EventDAGRunFinished,
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyStatus, e.Status),
		log.Int64(semantics.KeyDurationMS, e.DurationMS),
	)
}

func (l *Logger) DAGNodeStarted(ctx context.Context, e semantics.DAGNodeStarted) {
	requireValid(semantics.EventDAGNodeStarted, e.Validate())
	l.emit(ctx, semantics.EventDAGNodeStarted,
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyDAGNodeID, e.DAGNodeID),
	)
}

func (l *Logger) DAGNodeFinished(ctx context.Context, e semantics.DAGNodeFinished) {
	requireValid(semantics.EventDAGNodeFinished, e.Validate())
	l.emit(ctx, semantics.EventDAGNodeFinished,
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyDAGNodeID, e.DAGNodeID),
		log.String(semantics.KeyStatus, e.Status),
		log.Int64(semantics.KeyDurationMS, e.DurationMS),
	)
}

func (l *Logger) emit(ctx context.Context, eventName string, attrs ...log.KeyValue) {
	logger := l.lp.Logger(l.serviceName)
	var record log.Record
	record.SetEventName(eventName)
	record.SetBody(log.StringValue(eventName))
	record.AddAttributes(attrs...)
	addTraceCorrelation(&record, ctx)
	logger.Emit(ctx, record)
}

func addTraceCorrelation(record *log.Record, ctx context.Context) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return
	}
	record.AddAttributes(
		log.String("trace_id", sc.TraceID().String()),
		log.String("span_id", sc.SpanID().String()),
	)
}

func requireValid(eventName string, err error) {
	if err == nil {
		return
	}
	panic(fmt.Sprintf("intentlog: invalid %s: %v", eventName, err))
}
