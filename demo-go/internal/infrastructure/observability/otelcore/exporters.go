package otelcore

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

type exporters struct {
	traceExp  *otlptrace.Exporter
	metricExp *otlpmetrichttp.Exporter
	logExp    *otlploghttp.Exporter
}

func newExporters(ctx context.Context, cfg Config) (*exporters, error) {
	traceExp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpointURL(cfg.OTLPEndpoint, "traces", cfg.OTLPTracesEndpoint)),
	)
	if err != nil {
		return nil, err
	}
	metricExp, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL(endpointURL(cfg.OTLPEndpoint, "metrics", cfg.OTLPMetricsEndpoint)),
	)
	if err != nil {
		return nil, err
	}
	logExp, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpointURL(endpointURL(cfg.OTLPEndpoint, "logs", cfg.OTLPLogsEndpoint)),
	)
	if err != nil {
		return nil, err
	}
	return &exporters{traceExp: traceExp, metricExp: metricExp, logExp: logExp}, nil
}

func endpointURL(base, signal, override string) string {
	if override != "" {
		return override
	}
	base = strings.TrimRight(base, "/")
	return base + "/v1/" + signal
}
