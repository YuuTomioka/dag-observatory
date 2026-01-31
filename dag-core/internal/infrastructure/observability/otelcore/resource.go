package otelcore

import (
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func newResource(cfg Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(cfg.ServiceName),
		semconv.DeploymentEnvironmentName(cfg.Environment),
	}
	if cfg.ServiceNamespace != "" {
		attrs = append(attrs, semconv.ServiceNamespace(cfg.ServiceNamespace))
	}
	if cfg.ServiceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersion(cfg.ServiceVersion))
	}
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		attrs = append(attrs, attribute.String("service.instance.id", hostname))
	}

	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)
}
