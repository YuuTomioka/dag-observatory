package metrics

import (
	"go.opentelemetry.io/otel/metric"
)

type Instruments struct {
	DAGRunCounter metric.Int64Counter
	DAGRunLatency metric.Float64Histogram
}

func NewInstruments(m metric.Meter) (*Instruments, error) {
	cnt, err := m.Int64Counter("dag.run.count")
	if err != nil {
		return nil, err
	}
	lat, err := m.Float64Histogram("dag.run.latency_ms")
	if err != nil {
		return nil, err
	}
	return &Instruments{DAGRunCounter: cnt, DAGRunLatency: lat}, nil
}
