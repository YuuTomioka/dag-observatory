package semantics

const (
	EventClockTickReceived   = "clock.tick.received"
	EventDAGRunStarted       = "dag.run.started"
	EventDAGRunFinished      = "dag.run.finished"
	EventDAGRunFailed        = "dag.run.failed"
	EventDAGRunStateChanged  = "dag.run.state_changed"
	EventDAGNodeStarted      = "dag.node.started"
	EventDAGNodeFinished     = "dag.node.finished"
	EventDAGNodeFailed       = "dag.node.failed"
	EventDAGNodeTimeout      = "dag.node.timeout"
	EventDAGNodeSkipped      = "dag.node.skipped"
	EventDAGNodeStateChanged = "dag.node.state_changed"
)
