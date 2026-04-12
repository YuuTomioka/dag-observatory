package observer

import (
	"context"
	"errors"
	"strings"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	appport "dag-observatory/dag-core/internal/application/observability/port"
	"dag-observatory/dag-core/internal/domain/observability/semantics"
	"dag-observatory/dag-core/internal/infrastructure/observability/metrics"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type OTelObserver struct {
	intentLog appport.IntentLog
	metrics   *metrics.Instruments
}

func NewOTelObserver(intentLog appport.IntentLog, instruments *metrics.Instruments) *OTelObserver {
	return &OTelObserver{
		intentLog: intentLog,
		metrics:   instruments,
	}
}

func (o *OTelObserver) OnCompile(ctx context.Context, info port.CompileInfo) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("dag.compile", trace.WithAttributes(
		attribute.String("workflow", info.WorkflowName),
		attribute.Int("workflow.node_count", info.NodeCount),
	))
}

func (o *OTelObserver) OnCycleStart(ctx context.Context, info port.CycleInfo) {
	runID := info.Event.EventID
	partition := string(info.Partition)
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(correlationSpanAttrs(runID, partition, 0, "", "", "")...)
	if o.intentLog != nil && runID != "" {
		o.intentLog.DAGRunStateChanged(ctx, semantics.DAGRunStateChanged{
			DAGRunID:  runID,
			Partition: partition,
			FromState: "queued",
			ToState:   "running",
		})
		symbol := partition
		if symbol != "" {
			o.intentLog.DAGRunStarted(ctx, semantics.DAGRunStarted{
				DAGRunID:  runID,
				Partition: partition,
				Symbol:    symbol,
				InputSize: 0,
			})
		}
	}
}

func (o *OTelObserver) OnCycleEnd(ctx context.Context, info port.CycleResult) {
	runID := info.Event.EventID
	if runID == "" {
		return
	}
	durationMS := info.Duration.Milliseconds()
	partition := string(info.Partition)
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(correlationSpanAttrs(runID, partition, 0, "", "", "")...)
	attrs := cycleMetricAttrs(info.WorkflowName)

	if info.Err == nil {
		if o.intentLog != nil {
			o.intentLog.DAGRunStateChanged(ctx, semantics.DAGRunStateChanged{
				DAGRunID:   runID,
				Partition:  partition,
				FromState:  "running",
				ToState:    "succeeded",
				DurationMS: durationMS,
			})
			o.intentLog.DAGRunFinished(ctx, semantics.DAGRunFinished{
				DAGRunID:   runID,
				Partition:  partition,
				Status:     "succeeded",
				DurationMS: durationMS,
				RetryCount: info.RetryCount,
			})
		}
		o.recordRunMetrics(ctx, attrs, "succeeded", info.Duration)
		return
	}

	if o.intentLog != nil {
		o.intentLog.DAGRunStateChanged(ctx, semantics.DAGRunStateChanged{
			DAGRunID:   runID,
			Partition:  partition,
			FromState:  "running",
			ToState:    "failed",
			DurationMS: durationMS,
		})
		o.intentLog.DAGRunFailed(ctx, semantics.DAGRunFailed{
			DAGRunID:   runID,
			Partition:  partition,
			ErrorType:  "dagruntime.error",
			ErrorMsg:   info.Err.Error(),
			DurationMS: durationMS,
			RetryCount: info.RetryCount,
		})
	}
	o.recordRunMetrics(ctx, attrs, "failed", info.Duration)
}

func (o *OTelObserver) OnNodeStart(ctx context.Context, info port.NodeInfo) {
	partition := string(info.Partition)
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(correlationSpanAttrs(
		info.RunID,
		partition,
		info.SequenceNo,
		info.IntentID,
		info.ExecutionID,
		info.TradeID,
	)...)
	if o.intentLog == nil || info.RunID == "" {
		return
	}
	o.intentLog.DAGNodeStateChanged(ctx, semantics.DAGNodeStateChanged{
		DAGRunID:    info.RunID,
		Partition:   partition,
		SequenceNo:  info.SequenceNo,
		IntentID:    info.IntentID,
		ExecutionID: info.ExecutionID,
		TradeID:     info.TradeID,
		DAGNodeID:   info.NodeName,
		FromState:   "queued",
		ToState:     "running",
		QueueWaitMS: info.QueueWaitMS,
	})
	o.intentLog.DAGNodeStarted(ctx, semantics.DAGNodeStarted{
		DAGRunID:    info.RunID,
		Partition:   partition,
		SequenceNo:  info.SequenceNo,
		IntentID:    info.IntentID,
		ExecutionID: info.ExecutionID,
		TradeID:     info.TradeID,
		DAGNodeID:   info.NodeName,
	})
}

func (o *OTelObserver) OnNodeEnd(ctx context.Context, info port.NodeResult) {
	if info.RunID == "" {
		return
	}
	durationMS := info.Duration.Milliseconds()
	partition := string(info.Partition)
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(correlationSpanAttrs(
		info.RunID,
		partition,
		info.SequenceNo,
		info.IntentID,
		info.ExecutionID,
		info.TradeID,
	)...)
	attrs := nodeMetricAttrs(info.NodeName)

	if info.Err == nil {
		if o.intentLog != nil {
			o.intentLog.DAGNodeStateChanged(ctx, semantics.DAGNodeStateChanged{
				DAGRunID:    info.RunID,
				Partition:   partition,
				SequenceNo:  info.SequenceNo,
				IntentID:    info.IntentID,
				ExecutionID: info.ExecutionID,
				TradeID:     info.TradeID,
				DAGNodeID:   info.NodeName,
				FromState:   "running",
				ToState:     "succeeded",
				DurationMS:  durationMS,
				QueueWaitMS: info.QueueWaitMS,
			})
			o.intentLog.DAGNodeFinished(ctx, semantics.DAGNodeFinished{
				DAGRunID:    info.RunID,
				Partition:   partition,
				SequenceNo:  info.SequenceNo,
				IntentID:    info.IntentID,
				ExecutionID: info.ExecutionID,
				TradeID:     info.TradeID,
				DAGNodeID:   info.NodeName,
				Status:      "succeeded",
				DurationMS:  durationMS,
				RetryCount:  info.RetryCount,
			})
		}
		o.recordNodeMetrics(ctx, attrs, "succeeded", info.Duration)
		return
	}

	if errors.Is(info.Err, context.DeadlineExceeded) {
		if o.intentLog != nil {
			o.intentLog.DAGNodeStateChanged(ctx, semantics.DAGNodeStateChanged{
				DAGRunID:    info.RunID,
				Partition:   partition,
				SequenceNo:  info.SequenceNo,
				IntentID:    info.IntentID,
				ExecutionID: info.ExecutionID,
				TradeID:     info.TradeID,
				DAGNodeID:   info.NodeName,
				FromState:   "running",
				ToState:     "timeout",
				DurationMS:  durationMS,
				QueueWaitMS: info.QueueWaitMS,
			})
			o.intentLog.DAGNodeTimeout(ctx, semantics.DAGNodeTimeout{
				DAGRunID:    info.RunID,
				Partition:   partition,
				SequenceNo:  info.SequenceNo,
				IntentID:    info.IntentID,
				ExecutionID: info.ExecutionID,
				TradeID:     info.TradeID,
				DAGNodeID:   info.NodeName,
				DurationMS:  durationMS,
				RetryCount:  info.RetryCount,
			})
		}
		o.recordNodeMetrics(ctx, attrs, "timeout", info.Duration)
		return
	}

	if o.intentLog != nil {
		o.intentLog.DAGNodeStateChanged(ctx, semantics.DAGNodeStateChanged{
			DAGRunID:    info.RunID,
			Partition:   partition,
			SequenceNo:  info.SequenceNo,
			IntentID:    info.IntentID,
			ExecutionID: info.ExecutionID,
			TradeID:     info.TradeID,
			DAGNodeID:   info.NodeName,
			FromState:   "running",
			ToState:     "failed",
			DurationMS:  durationMS,
			QueueWaitMS: info.QueueWaitMS,
		})
		o.intentLog.DAGNodeFailed(ctx, semantics.DAGNodeFailed{
			DAGRunID:    info.RunID,
			Partition:   partition,
			SequenceNo:  info.SequenceNo,
			IntentID:    info.IntentID,
			ExecutionID: info.ExecutionID,
			TradeID:     info.TradeID,
			DAGNodeID:   info.NodeName,
			ErrorType:   classifyNodeErrorType(info.Err),
			ErrorMsg:    info.Err.Error(),
			DurationMS:  durationMS,
			RetryCount:  info.RetryCount,
		})
	}
	o.recordNodeMetrics(ctx, attrs, "failed", info.Duration)
}

func (o *OTelObserver) OnError(ctx context.Context, info port.ErrorInfo) {}

func (o *OTelObserver) recordRunMetrics(ctx context.Context, base []attribute.KeyValue, status string, duration time.Duration) {
	if o.metrics == nil {
		return
	}
	attrs := append(base, attribute.String(semantics.KeyStatus, status))
	o.metrics.DAGRunCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	o.metrics.DAGRunLatency.Record(ctx, float64(duration.Milliseconds()), metric.WithAttributes(attrs...))
}

func (o *OTelObserver) recordNodeMetrics(ctx context.Context, base []attribute.KeyValue, status string, duration time.Duration) {
	if o.metrics == nil {
		return
	}
	attrs := append(base, attribute.String(semantics.KeyStatus, status))
	o.metrics.DAGNodeCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	o.metrics.DAGNodeLatency.Record(ctx, float64(duration.Milliseconds()), metric.WithAttributes(attrs...))
}

func cycleMetricAttrs(workflow string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{}
	if workflow != "" {
		attrs = append(attrs, attribute.String("workflow", workflow))
	}
	return attrs
}

func nodeMetricAttrs(nodeName string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String(semantics.KeyDAGNodeID, nodeName),
	}
	return attrs
}

func correlationSpanAttrs(runID, partition string, sequenceNo int64, intentID, executionID, tradeID string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 6)
	if runID != "" {
		attrs = append(attrs, attribute.String(semantics.KeyDAGRunID, runID))
	}
	if partition != "" {
		attrs = append(attrs, attribute.String(semantics.KeyDAGPartition, partition))
	}
	if sequenceNo > 0 {
		attrs = append(attrs, attribute.Int64(semantics.KeyDAGSequenceNo, sequenceNo))
	}
	if intentID != "" {
		attrs = append(attrs, attribute.String(semantics.KeyDAGIntentID, intentID))
	}
	if executionID != "" {
		attrs = append(attrs, attribute.String(semantics.KeyDAGExecutionID, executionID))
	}
	if tradeID != "" {
		attrs = append(attrs, attribute.String(semantics.KeyDAGTradeID, tradeID))
	}
	return attrs
}

func classifyNodeErrorType(err error) string {
	if err == nil {
		return "dagruntime.node_error"
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout"),
		strings.Contains(msg, "deadline"):
		return "dagruntime.node_timeout"
	default:
		return "dagruntime.node_error"
	}
}
