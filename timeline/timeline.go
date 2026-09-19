// Package timeline provides deterministic clocks and scene selection.
package timeline

import (
	"fmt"
	"math"
	"time"
)

// Frame is an absolute sample, not a request to advance inside Draw.
type Frame struct {
	Tick        uint64
	Time, Delta float64
}

// Clock uses a fixed update rate and supports pause and playback speed.
type Clock struct {
	rate, speed, seconds float64
	tick                 uint64
	paused               bool
}

func NewClock(rate float64) (*Clock, error) {
	if rate <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) {
		return nil, fmt.Errorf("timeline: invalid tick rate")
	}
	return &Clock{rate: rate, speed: 1}, nil
}
func (c *Clock) SetSpeed(speed float64) error {
	if speed < 0 || math.IsNaN(speed) || math.IsInf(speed, 0) {
		return fmt.Errorf("timeline: invalid playback speed")
	}
	c.speed = speed
	return nil
}
func (c *Clock) Pause(paused bool) { c.paused = paused }
func (c *Clock) Reset()            { c.seconds = 0; c.tick = 0 }

// Step emits the initial frame at time zero, then advances for the next call.
func (c *Clock) Step() Frame {
	delta := c.speed / c.rate
	if c.paused {
		delta = 0
	}
	f := Frame{Tick: c.tick, Time: c.seconds, Delta: delta}
	c.seconds += delta
	c.tick++
	return f
}

type Sequence struct {
	ends  []time.Duration
	total time.Duration
	loop  bool
}

// NewSequence rejects zero/negative durations and overflowing totals.
func NewSequence(durations []time.Duration, loop bool) (*Sequence, error) {
	if len(durations) == 0 {
		return nil, fmt.Errorf("timeline: empty sequence")
	}
	s := &Sequence{loop: loop}
	for _, d := range durations {
		if d <= 0 || s.total > time.Duration(math.MaxInt64)-d {
			return nil, fmt.Errorf("timeline: invalid duration")
		}
		s.total += d
		s.ends = append(s.ends, s.total)
	}
	return s, nil
}

// At selects a half-open scene interval. A finished sequence returns ok=false.
func (s *Sequence) At(elapsed time.Duration) (index int, local time.Duration, ok bool) {
	if elapsed < 0 {
		return 0, 0, false
	}
	if s.loop {
		elapsed %= s.total
	} else if elapsed >= s.total {
		return 0, 0, false
	}
	start := time.Duration(0)
	for i, end := range s.ends {
		if elapsed < end {
			return i, elapsed - start, true
		}
		start = end
	}
	return 0, 0, false
}
func (s *Sequence) Duration() time.Duration { return s.total }
