package clock

import (
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
)

type RealClock struct{}

func NewRealClock() *RealClock {
	return &RealClock{}
}

func (c *RealClock) Now() time.Time {
	return time.Now()
}

var _ port.Clock = (*RealClock)(nil)

