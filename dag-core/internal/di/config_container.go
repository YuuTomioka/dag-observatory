package di

import "os"

type Config struct {
	ServiceName      string
	ServiceNamespace string
	ServiceVersion   string
	Env              string
	HTTPPort         string
	AppLogLevel      string
	AppLogOutput     string
	StateStoreType   string
	EventStoreType   string
	BoltPath         string
	KafkaBrokers     string
	KafkaTopic       string
	KafkaTaskTopic   string
	KafkaEventTopic  string
	KafkaGroupID     string

	OTLPEndpoint        string // e.g. http://otel-collector:4318
	OTLPLogsEndpoint    string // e.g. http://otel-collector:4318/v1/logs
	OTLPTracesEndpoint  string // e.g. http://otel-collector:4318/v1/traces
	OTLPMetricsEndpoint string // e.g. http://otel-collector:4318/v1/metrics
}

func NewConfig() Config {
	return Config{
		ServiceName:      getenv("SERVICE_NAME", "dag-observatory-dag-core"),
		ServiceNamespace: getenv("SERVICE_NAMESPACE", ""),
		ServiceVersion:   getenv("SERVICE_VERSION", ""),
		Env:              getenv("ENV", "dev"),
		HTTPPort:         getenv("HTTP_PORT", "8080"),
		AppLogLevel:      getenv("APP_LOG_LEVEL", "info"),
		AppLogOutput:     getenv("APP_LOG_OUTPUT", "stdout"),
		StateStoreType:   getenv("STATE_STORE_TYPE", "memory"),
		EventStoreType:   getenv("EVENT_STORE_TYPE", "memory"),
		BoltPath:         getenv("BOLT_PATH", "dagruntime.bolt"),
		KafkaBrokers:     getenv("KAFKA_BROKERS", ""),
		KafkaTopic:       getenv("KAFKA_TOPIC", ""),
		KafkaTaskTopic:   getenv("KAFKA_TASK_TOPIC", ""),
		KafkaEventTopic:  getenv("KAFKA_EVENT_TOPIC", ""),
		KafkaGroupID:     getenv("KAFKA_GROUP_ID", ""),
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
