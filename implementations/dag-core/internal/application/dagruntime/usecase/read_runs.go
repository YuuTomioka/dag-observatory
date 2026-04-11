package usecase

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type ListRuns struct {
	Reader port.RunReader
}

type ListRunsRequest struct {
	Partition string
}

type RunView struct {
	RunID      string
	Partition  string
	Status     string
	EventType  string
	EventTime  string
	StartedAt  string
	EndedAt    string
	DurationMS int64
	RetryCount int64
	Error      string
	StepCount  int
}

func (u *ListRuns) Execute(ctx context.Context, req ListRunsRequest) ([]RunView, error) {
	_ = ctx
	if u == nil || u.Reader == nil {
		return nil, nil
	}
	records := u.Reader.ListRuns(state.Partition(req.Partition))
	items := make([]RunView, 0, len(records))
	for _, run := range records {
		items = append(items, mapRunView(run, len(u.Reader.ListRunSteps(run.RunID))))
	}
	return items, nil
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
	sequenceNo, err := strconv.ParseInt(req.SequenceNo, 10, 64)
	if err != nil {
		return RunStepView{}, false, fmt.Errorf("sequence_no must be int64")
	}
	for _, step := range u.Reader.ListRunSteps(req.RunID) {
		if step.SequenceNo != sequenceNo {
			continue
		}
		return mapRunStep(step), true, nil
	}
	return RunStepView{}, false, nil
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
