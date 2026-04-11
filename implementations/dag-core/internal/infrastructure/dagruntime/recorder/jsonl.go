package recorder

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

type JSONLRecorder struct {
	mu   sync.Mutex
	file *os.File
	next port.Recorder
}

const jsonlTimeLayout = "2006-01-02T15:04:05.999999999Z07:00"

type jsonlObservationRecord struct {
	Kind              string                  `json:"kind"`
	RunID             string                  `json:"run_id,omitempty"`
	Partition         string                  `json:"partition,omitempty"`
	EventType         string                  `json:"event_type,omitempty"`
	EventTime         string                  `json:"event_time,omitempty"`
	SequenceNo        int64                   `json:"sequence_no,omitempty"`
	IntentID          string                  `json:"intent_id,omitempty"`
	ExecutionID       string                  `json:"execution_id,omitempty"`
	TradeID           string                  `json:"trade_id,omitempty"`
	NodeID            string                  `json:"node_id,omitempty"`
	NodeName          string                  `json:"node_name,omitempty"`
	Status            string                  `json:"status,omitempty"`
	TriggerReason     string                  `json:"trigger_reason,omitempty"`
	SkipReason        string                  `json:"skip_reason,omitempty"`
	InputRef          string                  `json:"input_ref,omitempty"`
	OutputRef         string                  `json:"output_ref,omitempty"`
	SnapshotBeforeRef string                  `json:"snapshot_before_ref,omitempty"`
	SnapshotAfterRef  string                  `json:"snapshot_after_ref,omitempty"`
	StateDiff         []events.StateDiffField `json:"state_diff,omitempty"`
	DurationNS        int64                   `json:"duration_ns,omitempty"`
	RetryCount        int64                   `json:"retry_count,omitempty"`
	Error             string                  `json:"error,omitempty"`
}

func NewJSONLRecorder(path string, next port.Recorder) (*JSONLRecorder, error) {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &JSONLRecorder{
		file: file,
		next: next,
	}, nil
}

func (r *JSONLRecorder) Close() error {
	if r == nil || r.file == nil {
		return nil
	}
	return r.file.Close()
}

func (r *JSONLRecorder) RecordEvent(ctx context.Context, event events.Event) {
	if r.next != nil {
		r.next.RecordEvent(ctx, event)
	}
	r.append(jsonlObservationRecord{
		Kind:      "run_event",
		RunID:     event.EventID,
		Partition: string(event.Partition),
		EventType: event.Type,
		EventTime: event.EventTime.UTC().Format(jsonlTimeLayout),
	})
}

func (r *JSONLRecorder) RecordNodeExecution(ctx context.Context, event events.NodeExecutionEvent) {
	if r.next != nil {
		r.next.RecordNodeExecution(ctx, event)
	}
	r.append(jsonlObservationRecord{
		Kind:              "node_execution",
		RunID:             event.RunID,
		Partition:         string(event.Partition),
		EventTime:         event.EventTime.UTC().Format(jsonlTimeLayout),
		SequenceNo:        event.SequenceNo,
		IntentID:          event.IntentID,
		ExecutionID:       event.ExecutionID,
		TradeID:           event.TradeID,
		NodeID:            event.NodeID,
		NodeName:          event.NodeName,
		Status:            string(event.Status),
		TriggerReason:     event.TriggerReason,
		SkipReason:        event.SkipReason,
		InputRef:          event.InputRef,
		OutputRef:         event.OutputRef,
		SnapshotBeforeRef: event.SnapshotBeforeRef,
		SnapshotAfterRef:  event.SnapshotAfterRef,
		StateDiff:         event.StateDiff,
		DurationNS:        event.DurationNS,
		Error:             event.Error,
	})
}

func (r *JSONLRecorder) RecordNodeResult(ctx context.Context, result port.NodeResult) {
	if r.next != nil {
		r.next.RecordNodeResult(ctx, result)
	}
}

func (r *JSONLRecorder) RecordCycleResult(ctx context.Context, result port.CycleResult) {
	if r.next != nil {
		r.next.RecordCycleResult(ctx, result)
	}
	r.append(jsonlObservationRecord{
		Kind:       "cycle_result",
		RunID:      result.Event.EventID,
		Partition:  string(result.Partition),
		EventType:  result.Event.Type,
		EventTime:  result.Event.EventTime.UTC().Format(jsonlTimeLayout),
		DurationNS: result.Duration.Nanoseconds(),
		RetryCount: result.RetryCount,
		Error:      recorderErrorString(result.Err),
	})
}

func (r *JSONLRecorder) append(record jsonlObservationRecord) {
	if r == nil || r.file == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	_ = json.NewEncoder(r.file).Encode(record)
}

func recorderErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
