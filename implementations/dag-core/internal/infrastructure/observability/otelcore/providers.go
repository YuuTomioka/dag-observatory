package otelcore

import (
	"context"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Core struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	LoggerProvider *sdklog.LoggerProvider
	Shutdown       func(ctx context.Context) error
}

func New(cfg Config) (*Core, error) {
	ctx := context.Background()

	res, err := newResource(cfg)
	if err != nil {
		return nil, err
	}

	exps, err := newExporters(ctx, cfg)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exps.traceExp),
	)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exps.metricExp)),
	)

	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exps.logExp)),
	)

	return &Core{
		TracerProvider: tp,
		MeterProvider:  mp,
		LoggerProvider: lp,
		Shutdown: func(ctx context.Context) error {
			var firstErr error
			if err := lp.Shutdown(ctx); err != nil {
				firstErr = err
			}
			if err := mp.Shutdown(ctx); err != nil && firstErr == nil {
				firstErr = err
			}
			if err := tp.Shutdown(ctx); err != nil && firstErr == nil {
				firstErr = err
			}
			return firstErr
		},
	}, nil
}
