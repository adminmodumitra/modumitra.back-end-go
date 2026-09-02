package sso

import "time"

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

type FrozenClock struct {
	T time.Time
}

func (c *FrozenClock) Now() time.Time {
	if c == nil || c.T.IsZero() {
		return time.Now().UTC()
	}
	return c.T.UTC()
}

func (c *FrozenClock) Advance(d time.Duration) {
	c.T = c.T.Add(d)
}
