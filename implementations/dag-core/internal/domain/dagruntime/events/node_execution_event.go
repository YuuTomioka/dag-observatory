package events

import (
	"time"

	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type NodeExecutionStatus string

const (
	NodeExecutionStatusUnknown   NodeExecutionStatus = "unknown"
	NodeExecutionStatusRunning   NodeExecutionStatus = "running"
	NodeExecutionStatusSucceeded NodeExecutionStatus = "succeeded"
	NodeExecutionStatusFailed    NodeExecutionStatus = "failed"
	NodeExecutionStatusSkipped   NodeExecutionStatus = "skipped"
)

// NodeExecutionEvent is the minimum observation unit for one node execution.
// The first version keeps string refs for payload/snapshot pointers so storage
// and transport decisions can be deferred.
type NodeExecutionEvent struct {
	RunID             string              `json:"run_id"`
	Partition         state.Partition     `json:"partition"`
	EventTime         time.Time           `json:"event_time"`
	SequenceNo        int64               `json:"sequence_no"`
	IntentID          string              `json:"intent_id,omitempty"`
	ExecutionID       string              `json:"execution_id,omitempty"`
	TradeID           string              `json:"trade_id,omitempty"`
	NodeID            string              `json:"node_id"`
	NodeName          string              `json:"node_name"`
	Status            NodeExecutionStatus `json:"status"`
	TriggerReason     string              `json:"trigger_reason,omitempty"`
	SkipReason        string              `json:"skip_reason,omitempty"`
	UpstreamNodeIDs   []string            `json:"upstream_node_ids,omitempty"`
	InputRef          string              `json:"input_ref,omitempty"`
	OutputRef         string              `json:"output_ref,omitempty"`
	SnapshotBeforeRef string              `json:"snapshot_before_ref,omitempty"`
	SnapshotAfterRef  string              `json:"snapshot_after_ref,omitempty"`
	StateDiff         []StateDiffField    `json:"state_diff,omitempty"`
	DurationNS        int64               `json:"duration_ns"`
	Error             string              `json:"error,omitempty"`
}

type StateDiffField struct {
	Field  string `json:"field"`
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
}
