package metrics

import (
	"go.opentelemetry.io/otel/metric"
)

type Instruments struct {
	DAGRunCounter           metric.Int64Counter
	DAGRunLatency           metric.Float64Histogram
	DAGNodeCounter          metric.Int64Counter
	DAGNodeLatency          metric.Float64Histogram
	DAGNodeQueueWait        metric.Int64Histogram
	IntentLogInvalidCounter metric.Int64Counter
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
	nodeCnt, err := m.Int64Counter("dag.node.count")
	if err != nil {
		return nil, err
	}
	nodeLat, err := m.Float64Histogram("dag.node.latency_ms")
	if err != nil {
		return nil, err
	}
	nodeQueue, err := m.Int64Histogram("dag.node.queue_wait_ms")
	if err != nil {
		return nil, err
	}
	intentInvalid, err := m.Int64Counter("intentlog.invalid.count")
	if err != nil {
		return nil, err
	}
	return &Instruments{
		DAGRunCounter:           cnt,
		DAGRunLatency:           lat,
		DAGNodeCounter:          nodeCnt,
		DAGNodeLatency:          nodeLat,
		DAGNodeQueueWait:        nodeQueue,
		IntentLogInvalidCounter: intentInvalid,
	}, nil
}
