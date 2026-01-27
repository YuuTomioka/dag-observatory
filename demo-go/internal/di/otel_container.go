package di

import (
	"context"

	"dag-observatory/demo-go/internal/application/port"
	"dag-observatory/demo-go/internal/infrastructure/observability/intentlog"
	"dag-observatory/demo-go/internal/infrastructure/observability/metrics"
	"dag-observatory/demo-go/internal/infrastructure/observability/otelcore"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type OTelContainer struct {
	Shutdown  func(ctx context.Context) error
	IntentLog port.IntentLog
	Tracer    trace.Tracer
	Meter     metric.Meter

	Metrics *metrics.Instruments
}

func NewOTelContainer(cfg Config) (*OTelContainer, error) {
	coreCfg := otelcore.Config{
		ServiceName:         cfg.ServiceName,
		Environment:         cfg.Env,
		OTLPEndpoint:        cfg.OTLPEndpoint,
		OTLPLogsEndpoint:    cfg.OTLPLogsEndpoint,
		OTLPTracesEndpoint:  cfg.OTLPTracesEndpoint,
		OTLPMetricsEndpoint: cfg.OTLPMetricsEndpoint,
	}

	core, err := otelcore.New(coreCfg)
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(core.TracerProvider)
	otel.SetMeterProvider(core.MeterProvider)

	tr := otel.Tracer(cfg.ServiceName)
	m := otel.Meter(cfg.ServiceName)

	inst, err := metrics.NewInstruments(m)
	if err != nil {
		return nil, err
	}

	il := intentlog.New(core.LoggerProvider, cfg.ServiceName)

	return &OTelContainer{
		Shutdown:  core.Shutdown,
		IntentLog: il,
		Tracer:    tr,
		Meter:     m,
		Metrics:   inst,
	}, nil
}
