package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"dag-observatory/demo-go/internal/domain/observability/ctxprop"
	"dag-observatory/demo-go/internal/domain/observability/semantics"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type dagRunReq struct {
	Symbol string `json:"symbol"`
	Mode   string `json:"mode"`
}

func (h *Handlers) DagRun(c echo.Context) error {
	ctx := c.Request().Context()

	var req dagRunReq
	if err := c.Bind(&req); err != nil {
		h.appLog.Warn(ctx, "dag run request bind failed",
			slog.String("error", err.Error()),
		)
		span := trace.SpanFromContext(ctx)
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request")
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "invalid request",
		})
	}
	if req.Symbol == "" {
		req.Symbol = "USDJPY"
	}
	mode := req.Mode
	if mode == "" {
		mode = c.QueryParam("mode")
	}

	runID := uuid.NewString()
	ctx = ctxprop.WithRunID(ctx, runID)

	// span: pretend this is a DAG run
	ctx, span := h.tracer.Start(ctx, "dag.run",
		// 重要キーは属性に
		trace.WithAttributes(
			attribute.String("symbol", req.Symbol),
			attribute.String(semantics.KeyDAGRunID, runID),
		),
	)
	defer span.End()

	start := time.Now()
	inputSize := int64(len(req.Symbol))
	baseAttrs := []attribute.KeyValue{
		attribute.String("symbol", req.Symbol),
		attribute.String("mode", mode),
	}

	h.appLog.Info(ctx, "dag run accepted",
		slog.String("symbol", req.Symbol),
		slog.String("mode", mode),
	)

	h.intentLog.DAGRunStateChanged(ctx, semantics.DAGRunStateChanged{
		DAGRunID:  runID,
		FromState: "queued",
		ToState:   "running",
	})

	// 意図ログ：DAG開始
	h.intentLog.DAGRunStarted(ctx, semantics.DAGRunStarted{
		DAGRunID:  runID,
		Symbol:    req.Symbol,
		InputSize: inputSize,
	})

	nodeID := "node-1"
	nodeStart := time.Now()

	queueWaitMS := int64(15)
	nodeBaseAttrs := []attribute.KeyValue{
		attribute.String("dag.node_id", nodeID),
		attribute.String("mode", mode),
	}

	if mode == "skip" {
		nodeAttrs := append(nodeBaseAttrs, attribute.String("status", "skipped"))
		h.intentLog.DAGNodeStateChanged(ctx, semantics.DAGNodeStateChanged{
			DAGRunID:    runID,
			DAGNodeID:   nodeID,
			FromState:   "queued",
			ToState:     "skipped",
			QueueWaitMS: queueWaitMS,
		})
		h.intentLog.DAGNodeSkipped(ctx, semantics.DAGNodeSkipped{
			DAGRunID:  runID,
			DAGNodeID: nodeID,
			Reason:    "filtered",
		})
		h.metrics.DAGNodeCounter.Add(ctx, 1, metric.WithAttributes(nodeAttrs...))
		h.metrics.DAGNodeQueueWait.Record(ctx, queueWaitMS, metric.WithAttributes(nodeAttrs...))
	} else {
		h.intentLog.DAGNodeStateChanged(ctx, semantics.DAGNodeStateChanged{
			DAGRunID:    runID,
			DAGNodeID:   nodeID,
			FromState:   "queued",
			ToState:     "running",
			QueueWaitMS: queueWaitMS,
		})
		h.intentLog.DAGNodeStarted(ctx, semantics.DAGNodeStarted{
			DAGRunID:  runID,
			DAGNodeID: nodeID,
		})

		// pretend to do work
		time.Sleep(120 * time.Millisecond)

		nodeDuration := time.Since(nodeStart).Milliseconds()
		h.metrics.DAGNodeQueueWait.Record(ctx, queueWaitMS, metric.WithAttributes(nodeBaseAttrs...))
		switch mode {
		case "timeout":
			nodeAttrs := append(nodeBaseAttrs, attribute.String("status", "timeout"))
			h.intentLog.DAGNodeStateChanged(ctx, semantics.DAGNodeStateChanged{
				DAGRunID:   runID,
				DAGNodeID:  nodeID,
				FromState:  "running",
				ToState:    "timeout",
				DurationMS: nodeDuration,
			})
			h.intentLog.DAGNodeTimeout(ctx, semantics.DAGNodeTimeout{
				DAGRunID:   runID,
				DAGNodeID:  nodeID,
				DurationMS: nodeDuration,
			})
			h.metrics.DAGNodeCounter.Add(ctx, 1, metric.WithAttributes(nodeAttrs...))
			h.metrics.DAGNodeLatency.Record(ctx, float64(nodeDuration), metric.WithAttributes(nodeAttrs...))
		case "fail":
			nodeAttrs := append(nodeBaseAttrs, attribute.String("status", "failed"))
			h.intentLog.DAGNodeStateChanged(ctx, semantics.DAGNodeStateChanged{
				DAGRunID:   runID,
				DAGNodeID:  nodeID,
				FromState:  "running",
				ToState:    "failed",
				DurationMS: nodeDuration,
			})
			h.intentLog.DAGNodeFailed(ctx, semantics.DAGNodeFailed{
				DAGRunID:   runID,
				DAGNodeID:  nodeID,
				ErrorType:  "demo.node_failed",
				ErrorMsg:   "node execution failed",
				DurationMS: nodeDuration,
			})
			h.metrics.DAGNodeCounter.Add(ctx, 1, metric.WithAttributes(nodeAttrs...))
			h.metrics.DAGNodeLatency.Record(ctx, float64(nodeDuration), metric.WithAttributes(nodeAttrs...))
		default:
			nodeAttrs := append(nodeBaseAttrs, attribute.String("status", "succeeded"))
			h.intentLog.DAGNodeStateChanged(ctx, semantics.DAGNodeStateChanged{
				DAGRunID:   runID,
				DAGNodeID:  nodeID,
				FromState:  "running",
				ToState:    "succeeded",
				DurationMS: nodeDuration,
			})
			h.intentLog.DAGNodeFinished(ctx, semantics.DAGNodeFinished{
				DAGRunID:   runID,
				DAGNodeID:  nodeID,
				Status:     "succeeded",
				DurationMS: nodeDuration,
			})
			h.metrics.DAGNodeCounter.Add(ctx, 1, metric.WithAttributes(nodeAttrs...))
			h.metrics.DAGNodeLatency.Record(ctx, float64(nodeDuration), metric.WithAttributes(nodeAttrs...))
		}
	}

	// metrics
	runDuration := time.Since(start).Milliseconds()

	switch mode {
	case "timeout":
		span := trace.SpanFromContext(ctx)
		span.RecordError(errors.New("dag run timeout"))
		span.SetAttributes(
			attribute.String("error.type", "demo.timeout"),
			attribute.String("error.message", "run timed out"),
		)
		span.SetStatus(codes.Error, "timeout")
		runAttrs := append(baseAttrs, attribute.String("status", "failed"))
		h.intentLog.DAGRunStateChanged(ctx, semantics.DAGRunStateChanged{
			DAGRunID:   runID,
			FromState:  "running",
			ToState:    "failed",
			DurationMS: runDuration,
		})
		h.intentLog.DAGRunFailed(ctx, semantics.DAGRunFailed{
			DAGRunID:   runID,
			ErrorType:  "demo.timeout",
			ErrorMsg:   "run timed out",
			DurationMS: runDuration,
		})
		h.appLog.Error(ctx, "dag run failed",
			slog.String("reason", "timeout"),
		)
		h.metrics.DAGRunCounter.Add(ctx, 1, metric.WithAttributes(runAttrs...))
		h.metrics.DAGRunLatency.Record(ctx, float64(runDuration), metric.WithAttributes(runAttrs...))
	case "fail":
		span := trace.SpanFromContext(ctx)
		span.RecordError(errors.New("dag run failed"))
		span.SetAttributes(
			attribute.String("error.type", "demo.failed"),
			attribute.String("error.message", "run failed"),
		)
		span.SetStatus(codes.Error, "failed")
		runAttrs := append(baseAttrs, attribute.String("status", "failed"))
		h.intentLog.DAGRunStateChanged(ctx, semantics.DAGRunStateChanged{
			DAGRunID:   runID,
			FromState:  "running",
			ToState:    "failed",
			DurationMS: runDuration,
		})
		h.intentLog.DAGRunFailed(ctx, semantics.DAGRunFailed{
			DAGRunID:   runID,
			ErrorType:  "demo.failed",
			ErrorMsg:   "run failed",
			DurationMS: runDuration,
		})
		h.appLog.Error(ctx, "dag run failed",
			slog.String("reason", "failed"),
		)
		h.metrics.DAGRunCounter.Add(ctx, 1, metric.WithAttributes(runAttrs...))
		h.metrics.DAGRunLatency.Record(ctx, float64(runDuration), metric.WithAttributes(runAttrs...))
	default:
		span := trace.SpanFromContext(ctx)
		span.SetStatus(codes.Ok, "succeeded")
		runAttrs := append(baseAttrs, attribute.String("status", "succeeded"))
		h.intentLog.DAGRunStateChanged(ctx, semantics.DAGRunStateChanged{
			DAGRunID:   runID,
			FromState:  "running",
			ToState:    "succeeded",
			DurationMS: runDuration,
		})
		h.intentLog.DAGRunFinished(ctx, semantics.DAGRunFinished{
			DAGRunID:   runID,
			Status:     "succeeded",
			DurationMS: runDuration,
		})
		h.appLog.Info(ctx, "dag run finished",
			slog.String("status", "succeeded"),
		)
		h.metrics.DAGRunCounter.Add(ctx, 1, metric.WithAttributes(runAttrs...))
		h.metrics.DAGRunLatency.Record(ctx, float64(runDuration), metric.WithAttributes(runAttrs...))
	}

	return c.JSON(200, map[string]any{
		"status": "started",
		"symbol": req.Symbol,
		"run_id": runID,
	})
}
