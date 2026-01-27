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
	DAGRunID string
	Symbol   string
}

func (e DAGRunStarted) Validate() error {
	if e.DAGRunID == "" {
		return errors.New("dag.run_id is required")
	}
	if e.Symbol == "" {
		return errors.New("symbol is required")
	}
	return nil
}

type DAGRunFinished struct {
	DAGRunID   string
	Status     string
	DurationMS int64
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
	return nil
}

type DAGNodeStarted struct {
	DAGRunID  string
	DAGNodeID string
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
	DAGRunID   string
	DAGNodeID  string
	Status     string
	DurationMS int64
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
	return nil
}
