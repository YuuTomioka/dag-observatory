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
	l.emit(ctx, semantics.EventClockTickReceived, log.SeverityInfo,
		log.String(semantics.KeyClockUTC, e.ClockUTC.UTC().Format(time.RFC3339Nano)),
		log.String(semantics.KeySymbol, e.Symbol),
	)
}

func (l *Logger) DAGRunStarted(ctx context.Context, e semantics.DAGRunStarted) {
	requireValid(semantics.EventDAGRunStarted, e.Validate())
	attrs := []log.KeyValue{
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeySymbol, e.Symbol),
	}
	if e.InputSize > 0 {
		attrs = append(attrs, log.Int64(semantics.KeyInputSize, e.InputSize))
	}
	l.emit(ctx, semantics.EventDAGRunStarted, log.SeverityInfo, attrs...)
}

func (l *Logger) DAGRunFinished(ctx context.Context, e semantics.DAGRunFinished) {
	requireValid(semantics.EventDAGRunFinished, e.Validate())
	attrs := []log.KeyValue{
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyStatus, e.Status),
		log.Int64(semantics.KeyDurationMS, e.DurationMS),
	}
	if e.RetryCount > 0 {
		attrs = append(attrs, log.Int64(semantics.KeyRetryCount, e.RetryCount))
	}
	l.emit(ctx, semantics.EventDAGRunFinished, log.SeverityInfo, attrs...)
}

func (l *Logger) DAGRunFailed(ctx context.Context, e semantics.DAGRunFailed) {
	requireValid(semantics.EventDAGRunFailed, e.Validate())
	attrs := []log.KeyValue{
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyErrorType, e.ErrorType),
		log.String(semantics.KeyErrorMsg, e.ErrorMsg),
		log.Int64(semantics.KeyDurationMS, e.DurationMS),
	}
	if e.RetryCount > 0 {
		attrs = append(attrs, log.Int64(semantics.KeyRetryCount, e.RetryCount))
	}
	l.emit(ctx, semantics.EventDAGRunFailed, log.SeverityError, attrs...)
}

func (l *Logger) DAGRunStateChanged(ctx context.Context, e semantics.DAGRunStateChanged) {
	requireValid(semantics.EventDAGRunStateChanged, e.Validate())
	attrs := []log.KeyValue{
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyFromState, e.FromState),
		log.String(semantics.KeyToState, e.ToState),
	}
	if e.DurationMS > 0 {
		attrs = append(attrs, log.Int64(semantics.KeyDurationMS, e.DurationMS))
	}
	l.emit(ctx, semantics.EventDAGRunStateChanged, log.SeverityInfo, attrs...)
}

func (l *Logger) DAGNodeStarted(ctx context.Context, e semantics.DAGNodeStarted) {
	requireValid(semantics.EventDAGNodeStarted, e.Validate())
	attrs := []log.KeyValue{
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyDAGNodeID, e.DAGNodeID),
	}
	attrs = appendParentAttrs(attrs, e.ParentNodeID, e.ParentNodeIDs)
	l.emit(ctx, semantics.EventDAGNodeStarted, log.SeverityInfo, attrs...)
}

func (l *Logger) DAGNodeFinished(ctx context.Context, e semantics.DAGNodeFinished) {
	requireValid(semantics.EventDAGNodeFinished, e.Validate())
	attrs := []log.KeyValue{
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyDAGNodeID, e.DAGNodeID),
		log.String(semantics.KeyStatus, e.Status),
		log.Int64(semantics.KeyDurationMS, e.DurationMS),
	}
	if e.RetryCount > 0 {
		attrs = append(attrs, log.Int64(semantics.KeyRetryCount, e.RetryCount))
	}
	attrs = appendParentAttrs(attrs, e.ParentNodeID, e.ParentNodeIDs)
	l.emit(ctx, semantics.EventDAGNodeFinished, log.SeverityInfo, attrs...)
}

func (l *Logger) DAGNodeFailed(ctx context.Context, e semantics.DAGNodeFailed) {
	requireValid(semantics.EventDAGNodeFailed, e.Validate())
	attrs := []log.KeyValue{
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyDAGNodeID, e.DAGNodeID),
		log.String(semantics.KeyErrorType, e.ErrorType),
		log.String(semantics.KeyErrorMsg, e.ErrorMsg),
		log.Int64(semantics.KeyDurationMS, e.DurationMS),
	}
	if e.RetryCount > 0 {
		attrs = append(attrs, log.Int64(semantics.KeyRetryCount, e.RetryCount))
	}
	attrs = appendParentAttrs(attrs, e.ParentNodeID, e.ParentNodeIDs)
	l.emit(ctx, semantics.EventDAGNodeFailed, log.SeverityError, attrs...)
}

func (l *Logger) DAGNodeTimeout(ctx context.Context, e semantics.DAGNodeTimeout) {
	requireValid(semantics.EventDAGNodeTimeout, e.Validate())
	attrs := []log.KeyValue{
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyDAGNodeID, e.DAGNodeID),
		log.Int64(semantics.KeyDurationMS, e.DurationMS),
	}
	if e.RetryCount > 0 {
		attrs = append(attrs, log.Int64(semantics.KeyRetryCount, e.RetryCount))
	}
	attrs = appendParentAttrs(attrs, e.ParentNodeID, e.ParentNodeIDs)
	l.emit(ctx, semantics.EventDAGNodeTimeout, log.SeverityError, attrs...)
}

func (l *Logger) DAGNodeSkipped(ctx context.Context, e semantics.DAGNodeSkipped) {
	requireValid(semantics.EventDAGNodeSkipped, e.Validate())
	attrs := []log.KeyValue{
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyDAGNodeID, e.DAGNodeID),
		log.String(semantics.KeyReason, e.Reason),
	}
	attrs = appendParentAttrs(attrs, e.ParentNodeID, e.ParentNodeIDs)
	l.emit(ctx, semantics.EventDAGNodeSkipped, log.SeverityWarn, attrs...)
}

func (l *Logger) DAGNodeStateChanged(ctx context.Context, e semantics.DAGNodeStateChanged) {
	requireValid(semantics.EventDAGNodeStateChanged, e.Validate())
	attrs := []log.KeyValue{
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyDAGNodeID, e.DAGNodeID),
		log.String(semantics.KeyFromState, e.FromState),
		log.String(semantics.KeyToState, e.ToState),
	}
	if e.DurationMS > 0 {
		attrs = append(attrs, log.Int64(semantics.KeyDurationMS, e.DurationMS))
	}
	if e.QueueWaitMS > 0 {
		attrs = append(attrs, log.Int64(semantics.KeyQueueWaitMS, e.QueueWaitMS))
	}
	attrs = appendParentAttrs(attrs, e.ParentNodeID, e.ParentNodeIDs)
	l.emit(ctx, semantics.EventDAGNodeStateChanged, log.SeverityInfo, attrs...)
}

func (l *Logger) emit(ctx context.Context, eventName string, severity log.Severity, attrs ...log.KeyValue) {
	logger := l.lp.Logger(l.serviceName)
	var record log.Record
	record.SetEventName(eventName)
	record.SetBody(log.StringValue(eventName))
	record.SetSeverity(severity)
	record.SetSeverityText(severity.String())
	record.AddAttributes(log.String(semantics.KeyEvent, eventName))
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

func appendParentAttrs(attrs []log.KeyValue, parentID string, parentIDs []string) []log.KeyValue {
	if parentID != "" {
		attrs = append(attrs, log.String(semantics.KeyDAGParentNodeID, parentID))
	}
	if len(parentIDs) > 0 {
		values := make([]log.Value, 0, len(parentIDs))
		for _, id := range parentIDs {
			if id == "" {
				continue
			}
			values = append(values, log.StringValue(id))
		}
		if len(values) > 0 {
			attrs = append(attrs, log.Slice(semantics.KeyDAGParentNodeIDs, values...))
		}
	}
	return attrs
}
