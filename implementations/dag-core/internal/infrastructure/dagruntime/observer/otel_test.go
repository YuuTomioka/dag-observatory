package observer

import (
	"context"
	"errors"
	"testing"
	"time"

	"dag-observatory/dag-core/internal/domain/observability/semantics"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	appport "dag-observatory/dag-core/internal/application/observability/port"
)

type fakeIntentLog struct {
	nodeStarted      []semantics.DAGNodeStarted
	nodeFinished     []semantics.DAGNodeFinished
	nodeFailed       []semantics.DAGNodeFailed
	nodeStateChanged []semantics.DAGNodeStateChanged
}

func (f *fakeIntentLog) ClockTickReceived(ctx context.Context, e semantics.ClockTickReceived) {}
func (f *fakeIntentLog) DAGRunStarted(ctx context.Context, e semantics.DAGRunStarted)         {}
func (f *fakeIntentLog) DAGRunFinished(ctx context.Context, e semantics.DAGRunFinished)       {}
func (f *fakeIntentLog) DAGRunFailed(ctx context.Context, e semantics.DAGRunFailed)           {}
func (f *fakeIntentLog) DAGRunStateChanged(ctx context.Context, e semantics.DAGRunStateChanged) {
}
func (f *fakeIntentLog) DAGNodeStarted(ctx context.Context, e semantics.DAGNodeStarted) {
	f.nodeStarted = append(f.nodeStarted, e)
}
func (f *fakeIntentLog) DAGNodeFinished(ctx context.Context, e semantics.DAGNodeFinished) {
	f.nodeFinished = append(f.nodeFinished, e)
}
func (f *fakeIntentLog) DAGNodeFailed(ctx context.Context, e semantics.DAGNodeFailed) {
	f.nodeFailed = append(f.nodeFailed, e)
}
func (f *fakeIntentLog) DAGNodeTimeout(ctx context.Context, e semantics.DAGNodeTimeout) {}
func (f *fakeIntentLog) DAGNodeSkipped(ctx context.Context, e semantics.DAGNodeSkipped) {}
func (f *fakeIntentLog) DAGNodeStateChanged(ctx context.Context, e semantics.DAGNodeStateChanged) {
	f.nodeStateChanged = append(f.nodeStateChanged, e)
}

func TestOTelObserverPropagatesCorrelationFieldsToIntentLog(t *testing.T) {
	t.Parallel()

	var intentLog appport.IntentLog = &fakeIntentLog{}
	observer := NewOTelObserver(intentLog, nil)

	observer.OnNodeStart(context.Background(), port.NodeInfo{
		RunID:       "run-1",
		Partition:   "partition-1",
		SequenceNo:  12,
		IntentID:    "intent-1",
		ExecutionID: "exec-1",
		TradeID:     "trade-1",
		NodeName:    "node-a",
		QueueWaitMS: 2,
	})
	observer.OnNodeEnd(context.Background(), port.NodeResult{
		RunID:       "run-1",
		Partition:   "partition-1",
		SequenceNo:  12,
		IntentID:    "intent-1",
		ExecutionID: "exec-1",
		TradeID:     "trade-1",
		NodeName:    "node-a",
		Duration:    10 * time.Millisecond,
		Err:         errors.New("boom"),
	})

	log := intentLog.(*fakeIntentLog)
	if len(log.nodeStarted) != 1 || len(log.nodeFailed) != 1 || len(log.nodeStateChanged) < 2 {
		t.Fatalf("unexpected emitted intent events: %+v", log)
	}
	start := log.nodeStarted[0]
	if start.Partition != "partition-1" || start.SequenceNo != 12 || start.IntentID != "intent-1" || start.ExecutionID != "exec-1" || start.TradeID != "trade-1" {
		t.Fatalf("missing start correlation fields: %#v", start)
	}
	failed := log.nodeFailed[0]
	if failed.Partition != "partition-1" || failed.SequenceNo != 12 || failed.IntentID != "intent-1" || failed.ExecutionID != "exec-1" || failed.TradeID != "trade-1" {
		t.Fatalf("missing failed correlation fields: %#v", failed)
	}
}

func TestMetricAttrsDoNotIncludeHighCardinalityCorrelationIDs(t *testing.T) {
	t.Parallel()

	cycle := cycleMetricAttrs("wf-1")
	for _, attr := range cycle {
		if attr.Key == semantics.KeyDAGRunID || attr.Key == semantics.KeyDAGIntentID || attr.Key == semantics.KeyDAGExecutionID || attr.Key == semantics.KeyDAGTradeID || attr.Key == semantics.KeyDAGSequenceNo {
			t.Fatalf("cycle metric attrs should not contain high-cardinality keys: %#v", cycle)
		}
	}

	node := nodeMetricAttrs("node-a")
	for _, attr := range node {
		if attr.Key == semantics.KeyDAGRunID || attr.Key == semantics.KeyDAGIntentID || attr.Key == semantics.KeyDAGExecutionID || attr.Key == semantics.KeyDAGTradeID || attr.Key == semantics.KeyDAGSequenceNo {
			t.Fatalf("node metric attrs should not contain high-cardinality keys: %#v", node)
		}
	}
}
