package otelcore

type Config struct {
	ServiceName         string
	ServiceNamespace    string
	ServiceVersion      string
	Environment         string
	OTLPEndpoint        string
	OTLPLogsEndpoint    string
	OTLPTracesEndpoint  string
	OTLPMetricsEndpoint string
}
