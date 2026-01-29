package policy

import "time"

type RetryPolicy struct {
	MaxAttempts int
	Backoff     time.Duration
}

