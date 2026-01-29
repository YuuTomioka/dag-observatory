package intentlog

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"dag-observatory/demo-go/internal/domain/observability/ctxprop"
	"dag-observatory/demo-go/internal/domain/observability/semantics"
	"dag-observatory/demo-go/internal/infrastructure/observability/applog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	lp             *sdklog.LoggerProvider
	serviceName    string
	invalidCounter metric.Int64Counter
	summaryLog     *slog.Logger
}

const (
	invalidReasonMissingRequired = "missing_required"
	invalidReasonTypeMismatch    = "type_mismatch"
	invalidReasonRangeViolation  = "range_violation"
	invalidReasonSchemaMismatch  = "schema_mismatch"
)

func New(lp *sdklog.LoggerProvider, serviceName string, invalidCounter metric.Int64Counter) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return &Logger{
		lp:             lp,
		serviceName:    serviceName,
		invalidCounter: invalidCounter,
		summaryLog:     slog.New(handler),
	}
}

func (l *Logger) ClockTickReceived(ctx context.Context, e semantics.ClockTickReceived) {
	if !l.valid(ctx, semantics.EventClockTickReceived, e.Validate()) {
		return
	}
	l.emit(ctx, semantics.EventClockTickReceived, log.SeverityInfo,
		log.String(semantics.KeyClockUTC, e.ClockUTC.UTC().Format(time.RFC3339Nano)),
		log.String(semantics.KeySymbol, e.Symbol),
	)
}

func (l *Logger) DAGRunStarted(ctx context.Context, e semantics.DAGRunStarted) {
	if !l.valid(ctx, semantics.EventDAGRunStarted, e.Validate()) {
		return
	}
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
	if !l.valid(ctx, semantics.EventDAGRunFinished, e.Validate()) {
		return
	}
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
	if !l.valid(ctx, semantics.EventDAGRunFailed, e.Validate()) {
		return
	}
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
	if !l.valid(ctx, semantics.EventDAGRunStateChanged, e.Validate()) {
		return
	}
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
	if !l.valid(ctx, semantics.EventDAGNodeStarted, e.Validate()) {
		return
	}
	attrs := []log.KeyValue{
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyDAGNodeID, e.DAGNodeID),
	}
	attrs = appendParentAttrs(attrs, e.ParentNodeID, e.ParentNodeIDs)
	l.emit(ctx, semantics.EventDAGNodeStarted, log.SeverityInfo, attrs...)
}

func (l *Logger) DAGNodeFinished(ctx context.Context, e semantics.DAGNodeFinished) {
	if !l.valid(ctx, semantics.EventDAGNodeFinished, e.Validate()) {
		return
	}
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
	if !l.valid(ctx, semantics.EventDAGNodeFailed, e.Validate()) {
		return
	}
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
	if !l.valid(ctx, semantics.EventDAGNodeTimeout, e.Validate()) {
		return
	}
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
	if !l.valid(ctx, semantics.EventDAGNodeSkipped, e.Validate()) {
		return
	}
	attrs := []log.KeyValue{
		log.String(semantics.KeyDAGRunID, e.DAGRunID),
		log.String(semantics.KeyDAGNodeID, e.DAGNodeID),
		log.String(semantics.KeyReason, e.Reason),
	}
	attrs = appendParentAttrs(attrs, e.ParentNodeID, e.ParentNodeIDs)
	l.emit(ctx, semantics.EventDAGNodeSkipped, log.SeverityWarn, attrs...)
}

func (l *Logger) DAGNodeStateChanged(ctx context.Context, e semantics.DAGNodeStateChanged) {
	if !l.valid(ctx, semantics.EventDAGNodeStateChanged, e.Validate()) {
		return
	}
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

func (l *Logger) valid(ctx context.Context, eventName string, err error) bool {
	if err == nil {
		return true
	}
	l.emitInvalid(ctx, eventName, err)
	return false
}

func (l *Logger) emitInvalid(ctx context.Context, eventName string, err error) {
	reason := classifyInvalidReason(err)
	logger := l.lp.Logger(l.serviceName)
	var record log.Record
	record.SetEventName("intentlog.invalid")
	record.SetBody(log.StringValue("intentlog.invalid"))
	record.SetSeverity(log.SeverityWarn)
	record.SetSeverityText(log.SeverityWarn.String())
	errorMessage, errorDetail := summarizeError(err.Error(), 200)
	record.AddAttributes(
		log.String("event.name", "intentlog.invalid"),
		log.String(semantics.KeyEvent, eventName),
		log.String("invalid.reason", reason),
		log.String(semantics.KeyErrorType, "intent/"+reason),
		log.String("error.message", errorMessage),
	)
	if errorDetail != "" {
		record.AddAttributes(log.String("error.detail", errorDetail))
	}
	if runID, ok := ctxprop.RunID(ctx); ok {
		record.AddAttributes(log.String(semantics.KeyDAGRunID, runID))
	}
	addTraceCorrelation(&record, ctx)
	logger.Emit(ctx, record)
	if l.invalidCounter != nil {
		l.invalidCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("invalid.reason", reason),
			attribute.String(semantics.KeyEvent, eventName),
		))
	}
	l.emitInvalidSummary(ctx, eventName, reason, errorMessage)
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

func (l *Logger) emitInvalidSummary(ctx context.Context, eventName, reason, errorMessage string) {
	if l.summaryLog == nil {
		return
	}
	attrs := []slog.Attr{
		slog.String("event.name", "intentlog.invalid"),
		slog.String(semantics.KeyEvent, eventName),
		slog.String("invalid.reason", reason),
		slog.String("error.message", errorMessage),
	}
	if runID, ok := ctxprop.RunID(ctx); ok {
		attrs = append(attrs, slog.String(semantics.KeyDAGRunID, runID))
	}
	attrs = append(attrs, traceAttrs(ctx)...)
	l.summaryLog.LogAttrs(ctx, slog.LevelWarn, "intentlog.invalid", attrs...)
}

func summarizeError(message string, max int) (string, string) {
	trimmed := strings.TrimSpace(message)
	if max <= 0 || len(trimmed) <= max {
		return applog.MaskSensitive(trimmed), ""
	}
	return applog.MaskSensitive(trimmed[:max]), applog.MaskSensitive(trimmed)
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

func classifyInvalidReason(err error) string {
	if err == nil {
		return invalidReasonSchemaMismatch
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "is required"):
		return invalidReasonMissingRequired
	case strings.Contains(msg, "must be >="),
		strings.Contains(msg, "must be <="),
		strings.Contains(msg, "must be >"),
		strings.Contains(msg, "must be <"):
		return invalidReasonRangeViolation
	case strings.Contains(msg, "invalid"),
		strings.Contains(msg, "unsupported"),
		strings.Contains(msg, "type"):
		return invalidReasonTypeMismatch
	default:
		return invalidReasonSchemaMismatch
	}
}
