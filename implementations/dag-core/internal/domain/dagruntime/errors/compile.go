package errors

type CompileError struct {
	Kind    string
	Message string
}

func (e CompileError) Error() string {
	return e.Kind + ": " + e.Message
}

const (
	CompileErrInputProvidesCollision = "input_provides_collision"
	CompileErrDuplicateProvider      = "duplicate_provider"
	CompileErrDuplicateWriter        = "duplicate_writer"
	CompileErrUnresolvedRequire      = "unresolved_require"
	CompileErrCycleDetected          = "cycle_detected"
)
