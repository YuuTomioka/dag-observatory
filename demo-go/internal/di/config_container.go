package di

import "os"

type Config struct {
	ServiceName      string
	ServiceNamespace string
	ServiceVersion   string
	Env              string
	HTTPPort         string

	OTLPEndpoint        string // e.g. http://otel-collector:4318
	OTLPLogsEndpoint    string // e.g. http://otel-collector:4318/v1/logs
	OTLPTracesEndpoint  string // e.g. http://otel-collector:4318/v1/traces
	OTLPMetricsEndpoint string // e.g. http://otel-collector:4318/v1/metrics
}

func NewConfig() Config {
	return Config{
		ServiceName:      getenv("SERVICE_NAME", "dag-observatory-demo-go"),
		ServiceNamespace: getenv("SERVICE_NAMESPACE", ""),
		ServiceVersion:   getenv("SERVICE_VERSION", ""),
		Env:              getenv("ENV", "dev"),
		HTTPPort:         getenv("HTTP_PORT", "8080"),
		OTLPEndpoint: getenv("OTEL_EXPORTER_OTLP_ENDPOINT",
			"http://otel-collector:4318"),
		OTLPLogsEndpoint:    getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", ""),
		OTLPTracesEndpoint:  getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", ""),
		OTLPMetricsEndpoint: getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", ""),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
