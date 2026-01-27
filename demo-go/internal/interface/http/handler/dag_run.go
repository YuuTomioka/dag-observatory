package handler

import (
	"time"

	"dag-observatory/demo-go/internal/domain/observability/semantics"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type dagRunReq struct {
	Symbol string `json:"symbol"`
}

func (h *Handlers) DagRun(c echo.Context) error {
	ctx := c.Request().Context()

	var req dagRunReq
	_ = c.Bind(&req)
	if req.Symbol == "" {
		req.Symbol = "USDJPY"
	}

	start := time.Now()

	// span: pretend this is a DAG run
	ctx, span := h.tracer.Start(ctx, "dag.run",
		// 重要キーは属性に
		trace.WithAttributes(
			attribute.String("symbol", req.Symbol),
		),
	)
	defer span.End()

	runID := "demo-" + time.Now().UTC().Format("20060102T150405Z")

	// 意図ログ：DAG開始
	h.intentLog.DAGRunStarted(ctx, semantics.DAGRunStarted{
		DAGRunID: runID,
		Symbol:   req.Symbol,
	})

	// pretend to do work
	time.Sleep(120 * time.Millisecond)

	// metrics
	h.metrics.DAGRunCounter.Add(ctx, 1)
	h.metrics.DAGRunLatency.Record(ctx, float64(time.Since(start).Milliseconds()))

	// 意図ログ：DAG終了
	h.intentLog.DAGRunFinished(ctx, semantics.DAGRunFinished{
		DAGRunID:   runID,
		Status:     "ok",
		DurationMS: time.Since(start).Milliseconds(),
	})

	return c.JSON(200, map[string]any{
		"status": "started",
		"symbol": req.Symbol,
	})
}
