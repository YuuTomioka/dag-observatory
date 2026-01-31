package di

import (
	"context"

	"dag-observatory/dag-core/internal/application/observability/port"
	"dag-observatory/dag-core/internal/infrastructure/observability/intentlog"
	"dag-observatory/dag-core/internal/infrastructure/observability/metrics"
	"dag-observatory/dag-core/internal/infrastructure/observability/otelcore"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log"
	otellog "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type OTelContainer struct {
	Shutdown  func(ctx context.Context) error
	IntentLog port.IntentLog
	Tracer    trace.Tracer
	Meter     metric.Meter
	AppLogger log.Logger

	Metrics *metrics.Instruments
}

func NewOTelContainer(cfg Config) (*OTelContainer, error) {
	coreCfg := otelcore.Config{
		ServiceName:         cfg.ServiceName,
		ServiceNamespace:    cfg.ServiceNamespace,
		ServiceVersion:      cfg.ServiceVersion,
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
	otellog.SetLoggerProvider(core.LoggerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tr := otel.Tracer(cfg.ServiceName)
	m := otel.Meter(cfg.ServiceName)
	appLogger := otellog.Logger(cfg.ServiceName)

	inst, err := metrics.NewInstruments(m)
	if err != nil {
		return nil, err
	}

	il := intentlog.New(core.LoggerProvider, cfg.ServiceName, inst.IntentLogInvalidCounter)

	return &OTelContainer{
		Shutdown:  core.Shutdown,
		IntentLog: il,
		Tracer:    tr,
		Meter:     m,
		AppLogger: appLogger,
		Metrics:   inst,
	}, nil
}
