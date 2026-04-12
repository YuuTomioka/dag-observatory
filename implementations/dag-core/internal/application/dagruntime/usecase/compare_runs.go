package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
	"dag-observatory/dag-core/internal/domain/dagruntime/events"
)

type CompareRuns struct {
	Reader port.RunReader
}

type CompareRunsRequest struct {
	BaseRunID   string
	TargetRunID string
}

type RunSummaryDeltaView struct {
	StatusChanged      bool  `json:"status_changed"`
	DurationMSDelta    int64 `json:"duration_ms_delta"`
	RetryCountDelta    int64 `json:"retry_count_delta"`
	StepCountDelta     int   `json:"step_count_delta"`
	FailedStepDelta    int   `json:"failed_step_delta"`
	SkippedStepDelta   int   `json:"skipped_step_delta"`
	SucceededStepDelta int   `json:"succeeded_step_delta"`
}

type RunStateFieldDeltaView struct {
	Field       string `json:"field"`
	BaseAfter   any    `json:"base_after,omitempty"`
	TargetAfter any    `json:"target_after,omitempty"`
	Changed     bool   `json:"changed"`
}

type RunCompareView struct {
	Base             RunView                  `json:"base"`
	Target           RunView                  `json:"target"`
	SummaryDelta     RunSummaryDeltaView      `json:"summary_delta"`
	StateFieldDeltas []RunStateFieldDeltaView `json:"state_field_deltas"`
}

func (u *CompareRuns) Execute(ctx context.Context, req CompareRunsRequest) (RunCompareView, bool, error) {
	_ = ctx
	if u == nil || u.Reader == nil {
		return RunCompareView{}, false, nil
	}
	if req.BaseRunID == "" || req.TargetRunID == "" {
		return RunCompareView{}, false, fmt.Errorf("base_run_id and target_run_id are required")
	}

	baseRecord, ok := u.Reader.GetRun(req.BaseRunID)
	if !ok {
		return RunCompareView{}, false, nil
	}
	targetRecord, ok := u.Reader.GetRun(req.TargetRunID)
	if !ok {
		return RunCompareView{}, false, nil
	}
	baseSteps := u.Reader.ListRunSteps(req.BaseRunID)
	targetSteps := u.Reader.ListRunSteps(req.TargetRunID)

	baseView := mapRunView(baseRecord, len(baseSteps))
	targetView := mapRunView(targetRecord, len(targetSteps))

	baseStatusCounts := countStepStatuses(baseSteps)
	targetStatusCounts := countStepStatuses(targetSteps)

	baseLatest := latestStateDiffByField(baseSteps)
	targetLatest := latestStateDiffByField(targetSteps)
	fields := unionFields(baseLatest, targetLatest)

	fieldDeltas := make([]RunStateFieldDeltaView, 0, len(fields))
	for _, field := range fields {
		baseValue, baseOK := baseLatest[field]
		targetValue, targetOK := targetLatest[field]
		changed := !baseOK || !targetOK || !equalCanonical(baseValue, targetValue)
		fieldDeltas = append(fieldDeltas, RunStateFieldDeltaView{
			Field:       field,
			BaseAfter:   baseValue,
			TargetAfter: targetValue,
			Changed:     changed,
		})
	}

	return RunCompareView{
		Base:   baseView,
		Target: targetView,
		SummaryDelta: RunSummaryDeltaView{
			StatusChanged:      baseView.Status != targetView.Status,
			DurationMSDelta:    targetView.DurationMS - baseView.DurationMS,
			RetryCountDelta:    targetView.RetryCount - baseView.RetryCount,
			StepCountDelta:     targetView.StepCount - baseView.StepCount,
			FailedStepDelta:    targetStatusCounts["failed"] - baseStatusCounts["failed"],
			SkippedStepDelta:   targetStatusCounts["skipped"] - baseStatusCounts["skipped"],
			SucceededStepDelta: targetStatusCounts["succeeded"] - baseStatusCounts["succeeded"],
		},
		StateFieldDeltas: fieldDeltas,
	}, true, nil
}

func countStepStatuses(steps []events.NodeExecutionEvent) map[string]int {
	out := map[string]int{
		"succeeded": 0,
		"failed":    0,
		"skipped":   0,
	}
	for _, step := range steps {
		out[string(step.Status)]++
	}
	return out
}

func latestStateDiffByField(steps []events.NodeExecutionEvent) map[string]any {
	out := make(map[string]any)
	sorted := make([]events.NodeExecutionEvent, len(steps))
	copy(sorted, steps)
	slices.SortFunc(sorted, func(a, b events.NodeExecutionEvent) int {
		if a.SequenceNo < b.SequenceNo {
			return -1
		}
		if a.SequenceNo > b.SequenceNo {
			return 1
		}
		return 0
	})
	for _, step := range sorted {
		for _, diff := range step.StateDiff {
			out[diff.Field] = diff.After
		}
	}
	return out
}

func unionFields(base, target map[string]any) []string {
	set := make(map[string]struct{}, len(base)+len(target))
	for field := range base {
		set[field] = struct{}{}
	}
	for field := range target {
		set[field] = struct{}{}
	}
	fields := make([]string, 0, len(set))
	for field := range set {
		fields = append(fields, field)
	}
	slices.Sort(fields)
	return fields
}

func equalCanonical(a, b any) bool {
	ab, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bb, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(ab) == string(bb)
}
