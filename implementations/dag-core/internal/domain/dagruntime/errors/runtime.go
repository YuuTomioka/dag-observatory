package errors

type RuntimeError struct {
	Kind    string
	Message string
	Err     error
}

func (e RuntimeError) Error() string {
	if e.Message != "" {
		return e.Kind + ": " + e.Message
	}
	return e.Kind
}

func (e RuntimeError) Unwrap() error {
	return e.Err
}

const (
	RuntimeErrNodeFailed       = "node_failed"
	RuntimeErrNodeTimeout      = "node_timeout"
	RuntimeErrRetryExhausted   = "retry_exhausted"
	RuntimeErrTxnCommitFailure = "txn_commit_failed"
)

type TimeoutError struct {
	Node string
	Err  error
}

func (e TimeoutError) Error() string {
	if e.Node == "" {
		return "node_timeout"
	}
	return "node_timeout: " + e.Node
}

func (e TimeoutError) Unwrap() error {
	return e.Err
}

type SkipError struct {
	Reason string
}

func (e SkipError) Error() string {
	if e.Reason == "" {
		return "node_skipped"
	}
	return "node_skipped: " + e.Reason
}
