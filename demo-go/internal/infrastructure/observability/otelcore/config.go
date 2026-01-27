package otelcore

type Config struct {
	ServiceName         string
	Environment         string
	OTLPEndpoint        string
	OTLPLogsEndpoint    string
	OTLPTracesEndpoint  string
	OTLPMetricsEndpoint string
}
