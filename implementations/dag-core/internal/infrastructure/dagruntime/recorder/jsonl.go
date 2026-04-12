package recorder

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type JSONLRecorder struct {
	mu       sync.Mutex
	file     *os.File
	path     string
	next     port.Recorder
	maxBytes int64
}

const jsonlTimeLayout = "2006-01-02T15:04:05.999999999Z07:00"
const defaultJSONLRotateMaxBytes int64 = 10 * 1024 * 1024

type ReplayStats struct {
	TotalLines   int64 `json:"total_lines"`
	AppliedLines int64 `json:"applied_lines"`
	SkippedLines int64 `json:"skipped_lines"`
	DecodeErrors int64 `json:"decode_errors"`
	ReplayErrors int64 `json:"replay_errors"`
	UnknownKinds int64 `json:"unknown_kinds"`
}

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
		file:     file,
		path:     path,
		next:     next,
		maxBytes: defaultJSONLRotateMaxBytes,
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
	r.rotateIfNeededLocked()
	_ = json.NewEncoder(r.file).Encode(record)
}

func recorderErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func ReplayJSONL(path string, sink port.Recorder) (ReplayStats, error) {
	stats := ReplayStats{}
	if sink == nil {
		return stats, nil
	}
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return stats, nil
		}
		return stats, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Allow reasonably large JSON lines (1 MiB).
	scanner.Buffer(make([]byte, 0, 64*1024), 1*1024*1024)
	for scanner.Scan() {
		stats.TotalLines++
		line := scanner.Bytes()
		var record jsonlObservationRecord
		if err := json.Unmarshal(line, &record); err != nil {
			stats.SkippedLines++
			stats.DecodeErrors++
			continue
		}
		applied, err := replayRecord(record, sink)
		if err != nil {
			stats.SkippedLines++
			stats.ReplayErrors++
			continue
		}
		if applied {
			stats.AppliedLines++
		} else {
			stats.SkippedLines++
			stats.UnknownKinds++
		}
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		return stats, err
	}
	return stats, nil
}

func replayRecord(record jsonlObservationRecord, sink port.Recorder) (bool, error) {
	eventTime, err := parseJSONLTime(record.EventTime)
	if err != nil {
		return false, fmt.Errorf("parse event_time: %w", err)
	}
	partition := state.Partition(record.Partition)

	switch record.Kind {
	case "run_event":
		sink.RecordEvent(context.Background(), events.Event{
			EventID:   record.RunID,
			EventTime: eventTime,
			Partition: partition,
			Type:      record.EventType,
		})
		return true, nil
	case "node_execution":
		sink.RecordNodeExecution(context.Background(), events.NodeExecutionEvent{
			RunID:             record.RunID,
			Partition:         partition,
			EventTime:         eventTime,
			SequenceNo:        record.SequenceNo,
			IntentID:          record.IntentID,
			ExecutionID:       record.ExecutionID,
			TradeID:           record.TradeID,
			NodeID:            record.NodeID,
			NodeName:          record.NodeName,
			Status:            events.NodeExecutionStatus(record.Status),
			TriggerReason:     record.TriggerReason,
			SkipReason:        record.SkipReason,
			InputRef:          record.InputRef,
			OutputRef:         record.OutputRef,
			SnapshotBeforeRef: record.SnapshotBeforeRef,
			SnapshotAfterRef:  record.SnapshotAfterRef,
			StateDiff:         record.StateDiff,
			DurationNS:        record.DurationNS,
			Error:             record.Error,
		})
		return true, nil
	case "cycle_result":
		var cycleErr error
		if record.Error != "" {
			cycleErr = errors.New(record.Error)
		}
		sink.RecordCycleResult(context.Background(), port.CycleResult{
			Partition: partition,
			Event: events.Event{
				EventID:   record.RunID,
				EventTime: eventTime,
				Partition: partition,
				Type:      record.EventType,
			},
			Duration:   time.Duration(record.DurationNS),
			RetryCount: record.RetryCount,
			Err:        cycleErr,
		})
		return true, nil
	default:
		// Ignore unknown kinds for forward compatibility.
		return false, nil
	}
}

func parseJSONLTime(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(jsonlTimeLayout, raw)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}

func (r *JSONLRecorder) rotateIfNeededLocked() {
	if r == nil || r.file == nil || r.path == "" || r.maxBytes <= 0 {
		return
	}
	stat, err := r.file.Stat()
	if err != nil || stat.Size() < r.maxBytes {
		return
	}
	_ = r.file.Close()
	backupPath := r.path + ".1"
	_ = os.Remove(backupPath)
	if err := os.Rename(r.path, backupPath); err != nil {
		return
	}
	file, err := os.OpenFile(r.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	r.file = file
}
