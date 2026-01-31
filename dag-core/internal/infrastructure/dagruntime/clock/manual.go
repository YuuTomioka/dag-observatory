package clock

import (
	"sync"
	"time"

	"dag-observatory/dag-core/internal/application/dagruntime/port"
)

type ManualClock struct {
	mu  sync.RWMutex
	now time.Time
}

func NewManualClock(initial time.Time) *ManualClock {
	return &ManualClock{now: initial}
}

func (c *ManualClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.now
}

func (c *ManualClock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = t
}

func (c *ManualClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

var _ port.Clock = (*ManualClock)(nil)

