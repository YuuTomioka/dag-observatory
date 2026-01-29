package errors

type RuntimeError struct {
	Kind    string
	Message string
}

func (e RuntimeError) Error() string {
	return e.Kind + ": " + e.Message
}

const (
	RuntimeErrNodeFailed       = "node_failed"
	RuntimeErrNodeTimeout      = "node_timeout"
	RuntimeErrRetryExhausted   = "retry_exhausted"
	RuntimeErrTxnCommitFailure = "txn_commit_failed"
)

