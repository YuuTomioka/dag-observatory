package node

import (
	"time"

	"dag-observatory/demo-go/internal/domain/dagruntime/policy"
)

type ExecutionSpec struct {
	Deterministic bool
	Idempotent    bool
	SideEffect    bool
	Timeout       time.Duration
	RetryPolicy   *policy.RetryPolicy
}

