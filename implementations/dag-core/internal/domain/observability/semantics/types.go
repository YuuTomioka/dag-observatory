package semantics

import (
	"errors"
	"time"
)

type ClockTickReceived struct {
	ClockUTC time.Time
	Symbol   string
}

func (e ClockTickReceived) Validate() error {
	if e.ClockUTC.IsZero() {
		return errors.New("clock.utc is required")
	}
	if e.Symbol == "" {
		return errors.New("symbol is required")
	}
	return nil
}

type DAGRunStarted struct {
	DAGRunID  string
	Symbol    string
	InputSize int64
}

func (e DAGRunStarted) Validate() error {
	if e.DAGRunID == "" {
		return errors.New("dag.run_id is required")
	}
	if e.Symbol == "" {
		return errors.New("symbol is required")
	}
	if e.InputSize < 0 {
		return errors.New("input_size must be >= 0")
	}
	return nil
}

type DAGRunFinished struct {
	DAGRunID   string
	Status     string
	DurationMS int64
	RetryCount int64
}

func (e DAGRunFinished) Validate() error {
	if e.DAGRunID == "" {
		return errors.New("dag.run_id is required")
	}
	if e.Status == "" {
		return errors.New("status is required")
	}
	if e.DurationMS < 0 {
		return errors.New("duration_ms must be >= 0")
	}
	if e.RetryCount < 0 {
		return errors.New("retry_count must be >= 0")
	}
	return nil
}

type DAGRunFailed struct {
	DAGRunID   string
	ErrorType  string
	ErrorMsg   string
	DurationMS int64
	RetryCount int64
}

func (e DAGRunFailed) Validate() error {
	if e.DAGRunID == "" {
		return errors.New("dag.run_id is required")
	}
	if e.ErrorType == "" {
		return errors.New("error.type is required")
	}
	if e.ErrorMsg == "" {
		return errors.New("error.msg is required")
	}
	if e.DurationMS < 0 {
		return errors.New("duration_ms must be >= 0")
	}
	if e.RetryCount < 0 {
		return errors.New("retry_count must be >= 0")
	}
	return nil
}

type DAGRunStateChanged struct {
	DAGRunID   string
	FromState  string
	ToState    string
	DurationMS int64
}

func (e DAGRunStateChanged) Validate() error {
	if e.DAGRunID == "" {
		return errors.New("dag.run_id is required")
	}
	if e.FromState == "" {
		return errors.New("from_state is required")
	}
	if e.ToState == "" {
		return errors.New("to_state is required")
	}
	if e.DurationMS < 0 {
		return errors.New("duration_ms must be >= 0")
	}
	return nil
}

type DAGNodeStarted struct {
	DAGRunID      string
	DAGNodeID     string
	ParentNodeID  string
	ParentNodeIDs []string
}

func (e DAGNodeStarted) Validate() error {
	if e.DAGRunID == "" {
		return errors.New("dag.run_id is required")
	}
	if e.DAGNodeID == "" {
		return errors.New("dag.node_id is required")
	}
	return nil
}

type DAGNodeFinished struct {
	DAGRunID      string
	DAGNodeID     string
	Status        string
	DurationMS    int64
	RetryCount    int64
	ParentNodeID  string
	ParentNodeIDs []string
}

func (e DAGNodeFinished) Validate() error {
	if e.DAGRunID == "" {
		return errors.New("dag.run_id is required")
	}
	if e.DAGNodeID == "" {
		return errors.New("dag.node_id is required")
	}
	if e.Status == "" {
		return errors.New("status is required")
	}
	if e.DurationMS < 0 {
		return errors.New("duration_ms must be >= 0")
	}
	if e.RetryCount < 0 {
		return errors.New("retry_count must be >= 0")
	}
	return nil
}

type DAGNodeFailed struct {
	DAGRunID      string
	DAGNodeID     string
	ErrorType     string
	ErrorMsg      string
	DurationMS    int64
	RetryCount    int64
	ParentNodeID  string
	ParentNodeIDs []string
}

func (e DAGNodeFailed) Validate() error {
	if e.DAGRunID == "" {
		return errors.New("dag.run_id is required")
	}
	if e.DAGNodeID == "" {
		return errors.New("dag.node_id is required")
	}
	if e.ErrorType == "" {
		return errors.New("error.type is required")
	}
	if e.ErrorMsg == "" {
		return errors.New("error.msg is required")
	}
	if e.DurationMS < 0 {
		return errors.New("duration_ms must be >= 0")
	}
	if e.RetryCount < 0 {
		return errors.New("retry_count must be >= 0")
	}
	return nil
}

type DAGNodeTimeout struct {
	DAGRunID      string
	DAGNodeID     string
	DurationMS    int64
	RetryCount    int64
	ParentNodeID  string
	ParentNodeIDs []string
}

func (e DAGNodeTimeout) Validate() error {
	if e.DAGRunID == "" {
		return errors.New("dag.run_id is required")
	}
	if e.DAGNodeID == "" {
		return errors.New("dag.node_id is required")
	}
	if e.DurationMS < 0 {
		return errors.New("duration_ms must be >= 0")
	}
	if e.RetryCount < 0 {
		return errors.New("retry_count must be >= 0")
	}
	return nil
}

type DAGNodeSkipped struct {
	DAGRunID      string
	DAGNodeID     string
	Reason        string
	ParentNodeID  string
	ParentNodeIDs []string
}

func (e DAGNodeSkipped) Validate() error {
	if e.DAGRunID == "" {
		return errors.New("dag.run_id is required")
	}
	if e.DAGNodeID == "" {
		return errors.New("dag.node_id is required")
	}
	if e.Reason == "" {
		return errors.New("reason is required")
	}
	return nil
}

type DAGNodeStateChanged struct {
	DAGRunID      string
	DAGNodeID     string
	FromState     string
	ToState       string
	DurationMS    int64
	QueueWaitMS   int64
	ParentNodeID  string
	ParentNodeIDs []string
}

func (e DAGNodeStateChanged) Validate() error {
	if e.DAGRunID == "" {
		return errors.New("dag.run_id is required")
	}
	if e.DAGNodeID == "" {
		return errors.New("dag.node_id is required")
	}
	if e.FromState == "" {
		return errors.New("from_state is required")
	}
	if e.ToState == "" {
		return errors.New("to_state is required")
	}
	if e.DurationMS < 0 {
		return errors.New("duration_ms must be >= 0")
	}
	if e.QueueWaitMS < 0 {
		return errors.New("queue_wait_ms must be >= 0")
	}
	return nil
}
