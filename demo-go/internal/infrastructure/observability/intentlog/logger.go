package intentlog

import (
	"context"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type Logger struct {
	lp          *sdklog.LoggerProvider
	serviceName string
}

func New(lp *sdklog.LoggerProvider, serviceName string) *Logger {
	return &Logger{lp: lp, serviceName: serviceName}
}

func (l *Logger) DAGRunStarted(ctx context.Context, e DAGRunEvent) {
	logger := l.lp.Logger(l.serviceName)
	var record log.Record
	record.SetBody(log.StringValue("intent.dag_run_started"))
	record.AddAttributes(
		log.String("dag.run_id", e.DAGRunID),
		log.String("symbol", e.Symbol),
	)
	logger.Emit(ctx, record)
}

func (l *Logger) DAGRunFinished(ctx context.Context, e DAGRunEvent) {
	logger := l.lp.Logger(l.serviceName)
	var record log.Record
	record.SetBody(log.StringValue("intent.dag_run_finished"))
	record.AddAttributes(
		log.String("dag.run_id", e.DAGRunID),
		log.String("symbol", e.Symbol),
	)
	logger.Emit(ctx, record)
}
