package handler

import (
	"dag-observatory/dag-core/internal/application/dagruntime/usecase"
	"dag-observatory/dag-core/internal/infrastructure/observability/applog"
	"dag-observatory/dag-core/internal/infrastructure/observability/metrics"

	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	appLog       *applog.Logger
	tracer       trace.Tracer
	metrics      *metrics.Instruments
	runWF        *usecase.RunWorkflow
	listRuns     *usecase.ListRuns
	getRun       *usecase.GetRun
	listRunSteps *usecase.ListRunSteps
	getRunNode   *usecase.GetRunNode
	compareRuns  *usecase.CompareRuns
}

type Dependencies struct {
	AppLog       *applog.Logger
	Tracer       trace.Tracer
	Metrics      *metrics.Instruments
	RunWorkflow  *usecase.RunWorkflow
	ListRuns     *usecase.ListRuns
	GetRun       *usecase.GetRun
	ListRunSteps *usecase.ListRunSteps
	GetRunNode   *usecase.GetRunNode
	CompareRuns  *usecase.CompareRuns
}

func New(d Dependencies) *Handler {
	return &Handler{
		appLog:       d.AppLog,
		tracer:       d.Tracer,
		metrics:      d.Metrics,
		runWF:        d.RunWorkflow,
		listRuns:     d.ListRuns,
		getRun:       d.GetRun,
		listRunSteps: d.ListRunSteps,
		getRunNode:   d.GetRunNode,
		compareRuns:  d.CompareRuns,
	}
}
