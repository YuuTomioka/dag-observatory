package metrics

import (
	"go.opentelemetry.io/otel/metric"
)

type Instruments struct {
	DAGRunCounter       metric.Int64Counter
	DAGRunLatency       metric.Float64Histogram
	HTTPRequestCounter  metric.Int64Counter
	HTTPRequestDuration metric.Float64Histogram
	DAGNodeCounter      metric.Int64Counter
	DAGNodeLatency      metric.Float64Histogram
	DAGNodeQueueWait    metric.Int64Histogram
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
	httpCnt, err := m.Int64Counter("http.server.request.count")
	if err != nil {
		return nil, err
	}
	httpDur, err := m.Float64Histogram("http.server.duration_ms")
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
	return &Instruments{
		DAGRunCounter:       cnt,
		DAGRunLatency:       lat,
		HTTPRequestCounter:  httpCnt,
		HTTPRequestDuration: httpDur,
		DAGNodeCounter:      nodeCnt,
		DAGNodeLatency:      nodeLat,
		DAGNodeQueueWait:    nodeQueue,
	}, nil
}
