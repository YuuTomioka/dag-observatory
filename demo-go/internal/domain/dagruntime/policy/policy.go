package policy

import "time"

type Policy struct {
	DefaultTimeout time.Duration
	DefaultRetry   RetryPolicy
}

