package observer

import (
	"context"
	"errors"
	"strings"
	"time"

	appport "dag-observatory/demo-go/internal/application/port"
	"dag-observatory/demo-go/internal/application/dagruntime/port"
	"dag-observatory/demo-go/internal/domain/observability/semantics"
	"dag-observatory/demo-go/internal/infrastructure/observability/metrics"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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

func (o *OTelObserver) OnCompile(ctx context.Context, info port.CompileInfo) {}

func (o *OTelObserver) OnCycleStart(ctx context.Context, info port.CycleInfo) {
	runID := info.Event.EventID
	if o.intentLog != nil && runID != "" {
		o.intentLog.DAGRunStateChanged(ctx, semantics.DAGRunStateChanged{
			DAGRunID:  runID,
			FromState: "queued",
			ToState:   "running",
		})
		symbol := string(info.Partition)
		if symbol != "" {
			o.intentLog.DAGRunStarted(ctx, semantics.DAGRunStarted{
				DAGRunID:  runID,
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
	symbol := string(info.Partition)
	attrs := cycleAttrs(info.WorkflowName, symbol, runID)

	if info.Err == nil {
		if o.intentLog != nil {
			o.intentLog.DAGRunStateChanged(ctx, semantics.DAGRunStateChanged{
				DAGRunID:   runID,
				FromState:  "running",
				ToState:    "succeeded",
				DurationMS: durationMS,
			})
			o.intentLog.DAGRunFinished(ctx, semantics.DAGRunFinished{
				DAGRunID:   runID,
				Status:     "succeeded",
				DurationMS: durationMS,
			})
		}
		o.recordRunMetrics(ctx, attrs, "succeeded", info.Duration)
		return
	}

	if o.intentLog != nil {
		o.intentLog.DAGRunStateChanged(ctx, semantics.DAGRunStateChanged{
			DAGRunID:   runID,
			FromState:  "running",
			ToState:    "failed",
			DurationMS: durationMS,
		})
		o.intentLog.DAGRunFailed(ctx, semantics.DAGRunFailed{
			DAGRunID:   runID,
			ErrorType:  "dagruntime.error",
			ErrorMsg:   info.Err.Error(),
			DurationMS: durationMS,
		})
	}
	o.recordRunMetrics(ctx, attrs, "failed", info.Duration)
}

func (o *OTelObserver) OnNodeStart(ctx context.Context, info port.NodeInfo) {
	if o.intentLog == nil || info.RunID == "" {
		return
	}
	o.intentLog.DAGNodeStateChanged(ctx, semantics.DAGNodeStateChanged{
		DAGRunID:  info.RunID,
		DAGNodeID: info.NodeName,
		FromState: "queued",
		ToState:   "running",
	})
	o.intentLog.DAGNodeStarted(ctx, semantics.DAGNodeStarted{
		DAGRunID:  info.RunID,
		DAGNodeID: info.NodeName,
	})
}

func (o *OTelObserver) OnNodeEnd(ctx context.Context, info port.NodeResult) {
	if info.RunID == "" {
		return
	}
	durationMS := info.Duration.Milliseconds()
	attrs := nodeAttrs(info.NodeName, info.RunID)

	if info.Err == nil {
		if o.intentLog != nil {
			o.intentLog.DAGNodeStateChanged(ctx, semantics.DAGNodeStateChanged{
				DAGRunID:   info.RunID,
				DAGNodeID:  info.NodeName,
				FromState:  "running",
				ToState:    "succeeded",
				DurationMS: durationMS,
			})
			o.intentLog.DAGNodeFinished(ctx, semantics.DAGNodeFinished{
				DAGRunID:   info.RunID,
				DAGNodeID:  info.NodeName,
				Status:     "succeeded",
				DurationMS: durationMS,
			})
		}
		o.recordNodeMetrics(ctx, attrs, "succeeded", info.Duration)
		return
	}

	if errors.Is(info.Err, context.DeadlineExceeded) {
		if o.intentLog != nil {
			o.intentLog.DAGNodeStateChanged(ctx, semantics.DAGNodeStateChanged{
				DAGRunID:   info.RunID,
				DAGNodeID:  info.NodeName,
				FromState:  "running",
				ToState:    "timeout",
				DurationMS: durationMS,
			})
			o.intentLog.DAGNodeTimeout(ctx, semantics.DAGNodeTimeout{
				DAGRunID:   info.RunID,
				DAGNodeID:  info.NodeName,
				DurationMS: durationMS,
			})
		}
		o.recordNodeMetrics(ctx, attrs, "timeout", info.Duration)
		return
	}

	if o.intentLog != nil {
		o.intentLog.DAGNodeStateChanged(ctx, semantics.DAGNodeStateChanged{
			DAGRunID:   info.RunID,
			DAGNodeID:  info.NodeName,
			FromState:  "running",
			ToState:    "failed",
			DurationMS: durationMS,
		})
		o.intentLog.DAGNodeFailed(ctx, semantics.DAGNodeFailed{
			DAGRunID:   info.RunID,
			DAGNodeID:  info.NodeName,
			ErrorType:  classifyNodeErrorType(info.Err),
			ErrorMsg:   info.Err.Error(),
			DurationMS: durationMS,
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

func cycleAttrs(workflow, symbol, runID string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String(semantics.KeyDAGRunID, runID),
	}
	if workflow != "" {
		attrs = append(attrs, attribute.String("workflow", workflow))
	}
	if symbol != "" {
		attrs = append(attrs, attribute.String(semantics.KeySymbol, symbol))
	}
	return attrs
}

func nodeAttrs(nodeName, runID string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String(semantics.KeyDAGRunID, runID),
		attribute.String(semantics.KeyDAGNodeID, nodeName),
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
