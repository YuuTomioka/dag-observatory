package usecase

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type ListRuns struct {
	Reader port.RunReader
}

type ListRunsRequest struct {
	Partition string
	Status    string
	Since     *time.Time
	Until     *time.Time
	Limit     int
	Cursor    string
}

type ListRunsResult struct {
	Items      []RunView `json:"items"`
	NextCursor string    `json:"next_cursor,omitempty"`
}

type RunView struct {
	RunID      string `json:"run_id"`
	Partition  string `json:"partition"`
	Status     string `json:"status"`
	EventType  string `json:"event_type"`
	EventTime  string `json:"event_time"`
	StartedAt  string `json:"started_at"`
	EndedAt    string `json:"ended_at"`
	DurationMS int64  `json:"duration_ms"`
	RetryCount int64  `json:"retry_count"`
	Error      string `json:"error,omitempty"`
	StepCount  int    `json:"step_count"`
}

func (u *ListRuns) Execute(ctx context.Context, req ListRunsRequest) (ListRunsResult, error) {
	_ = ctx
	if u == nil || u.Reader == nil {
		return ListRunsResult{}, nil
	}
	records := u.Reader.ListRuns(state.Partition(req.Partition))
	filtered := make([]port.RunRecord, 0, len(records))
	for _, run := range records {
		if req.Status != "" && !strings.EqualFold(run.Status, req.Status) {
			continue
		}
		if req.Since != nil && run.StartedAt.Before(*req.Since) {
			continue
		}
		if req.Until != nil && !run.StartedAt.Before(*req.Until) {
			continue
		}
		filtered = append(filtered, run)
	}
	slices.SortFunc(filtered, func(a, b port.RunRecord) int {
		if a.StartedAt.After(b.StartedAt) {
			return -1
		}
		if b.StartedAt.After(a.StartedAt) {
			return 1
		}
		if a.RunID > b.RunID {
			return -1
		}
		if a.RunID < b.RunID {
			return 1
		}
		return 0
	})

	offset := 0
	if req.Cursor != "" {
		parsed, err := strconv.Atoi(req.Cursor)
		if err != nil || parsed < 0 {
			return ListRunsResult{}, fmt.Errorf("cursor must be non-negative integer offset")
		}
		offset = parsed
	}
	if offset >= len(filtered) {
		return ListRunsResult{Items: []RunView{}}, nil
	}
	end := len(filtered)
	if req.Limit > 0 && offset+req.Limit < end {
		end = offset + req.Limit
	}

	items := make([]RunView, 0, end-offset)
	for _, run := range filtered[offset:end] {
		items = append(items, mapRunView(run, len(u.Reader.ListRunSteps(run.RunID))))
	}
	nextCursor := ""
	if end < len(filtered) {
		nextCursor = strconv.Itoa(end)
	}
	return ListRunsResult{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
}

type GetRun struct {
	Reader port.RunReader
}

type GetRunRequest struct {
	RunID string
}

func (u *GetRun) Execute(ctx context.Context, req GetRunRequest) (RunView, bool, error) {
	_ = ctx
	if u == nil || u.Reader == nil {
		return RunView{}, false, nil
	}
	record, ok := u.Reader.GetRun(req.RunID)
	if !ok {
		return RunView{}, false, nil
	}
	return mapRunView(record, len(u.Reader.ListRunSteps(record.RunID))), true, nil
}

type ListRunSteps struct {
	Reader port.RunReader
}

type ListRunStepsRequest struct {
	RunID string
}

type RunStepView struct {
	NodeExecutionID   string                  `json:"node_execution_id"`
	Partition         string                  `json:"partition"`
	SequenceNo        int64                   `json:"sequence_no"`
	IntentID          string                  `json:"intent_id,omitempty"`
	ExecutionID       string                  `json:"execution_id,omitempty"`
	TradeID           string                  `json:"trade_id,omitempty"`
	NodeID            string                  `json:"node_id"`
	NodeName          string                  `json:"node_name"`
	Status            string                  `json:"status"`
	TriggerReason     string                  `json:"trigger_reason,omitempty"`
	SkipReason        string                  `json:"skip_reason,omitempty"`
	InputRef          string                  `json:"input_ref,omitempty"`
	OutputRef         string                  `json:"output_ref,omitempty"`
	SnapshotBeforeRef string                  `json:"snapshot_before_ref,omitempty"`
	SnapshotAfterRef  string                  `json:"snapshot_after_ref,omitempty"`
	StateDiff         []events.StateDiffField `json:"state_diff,omitempty"`
	DurationNS        int64                   `json:"duration_ns"`
	Error             string                  `json:"error,omitempty"`
}

func (u *ListRunSteps) Execute(ctx context.Context, req ListRunStepsRequest) ([]RunStepView, error) {
	_ = ctx
	if u == nil || u.Reader == nil {
		return nil, nil
	}
	steps := u.Reader.ListRunSteps(req.RunID)
	out := make([]RunStepView, 0, len(steps))
	for _, step := range steps {
		out = append(out, mapRunStep(step))
	}
	slices.SortFunc(out, func(a, b RunStepView) int {
		if a.SequenceNo < b.SequenceNo {
			return -1
		}
		if a.SequenceNo > b.SequenceNo {
			return 1
		}
		return 0
	})
	return out, nil
}

type GetRunNode struct {
	Reader port.RunReader
}

type GetRunNodeRequest struct {
	RunID      string
	SequenceNo string
}

func (u *GetRunNode) Execute(ctx context.Context, req GetRunNodeRequest) (RunStepView, bool, error) {
	_ = ctx
	if u == nil || u.Reader == nil {
		return RunStepView{}, false, nil
	}
	steps := u.Reader.ListRunSteps(req.RunID)
	sequenceNo, err := strconv.ParseInt(req.SequenceNo, 10, 64)
	if err == nil {
		for _, step := range steps {
			if step.SequenceNo != sequenceNo {
				continue
			}
			return mapRunStep(step), true, nil
		}
		return RunStepView{}, false, nil
	}

	// Compatibility path for legacy /nodes/:execution_id callers.
	for _, step := range steps {
		if step.ExecutionID != req.SequenceNo {
			continue
		}
		return mapRunStep(step), true, nil
	}
	return RunStepView{}, false, fmt.Errorf("sequence_no must be int64")
}

func mapRunView(record port.RunRecord, stepCount int) RunView {
	return RunView{
		RunID:      record.RunID,
		Partition:  string(record.Partition),
		Status:     record.Status,
		EventType:  record.EventType,
		EventTime:  record.EventTime.Format(timeLayout),
		StartedAt:  record.StartedAt.Format(timeLayout),
		EndedAt:    record.EndedAt.Format(timeLayout),
		DurationMS: record.Duration.Milliseconds(),
		RetryCount: record.RetryCount,
		Error:      record.Error,
		StepCount:  stepCount,
	}
}

func mapRunStep(step events.NodeExecutionEvent) RunStepView {
	return RunStepView{
		NodeExecutionID:   strconv.FormatInt(step.SequenceNo, 10),
		Partition:         string(step.Partition),
		SequenceNo:        step.SequenceNo,
		IntentID:          step.IntentID,
		ExecutionID:       step.ExecutionID,
		TradeID:           step.TradeID,
		NodeID:            step.NodeID,
		NodeName:          step.NodeName,
		Status:            string(step.Status),
		TriggerReason:     step.TriggerReason,
		SkipReason:        step.SkipReason,
		InputRef:          step.InputRef,
		OutputRef:         step.OutputRef,
		SnapshotBeforeRef: step.SnapshotBeforeRef,
		SnapshotAfterRef:  step.SnapshotAfterRef,
		StateDiff:         step.StateDiff,
		DurationNS:        step.DurationNS,
		Error:             step.Error,
	}
}

const timeLayout = "2006-01-02T15:04:05.999999999Z07:00"
