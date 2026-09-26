package timeline

import (
	"fmt"
	"math"
)

// CueRampConfig interpolates one scalar over an existing CueClock window.
// Duration zero uses that window's duration. From and To may describe audio
// gain, opacity, position or another caller-owned value.
type CueRampConfig struct {
	Window   int
	From, To float64
	Duration float64
}

// CueRamp samples an existing clock without advancing it. Rate changes and
// resets on the clock immediately affect this envelope; Value allocates nothing.
type CueRamp struct {
	clock    *CueClock
	window   int
	from, to float64
	duration float64
}

func NewCueRamp(clock *CueClock, config CueRampConfig) (*CueRamp, error) {
	if clock == nil || config.Window < 0 || config.Window >= len(clock.windows) ||
		math.IsNaN(config.From) || math.IsInf(config.From, 0) ||
		math.IsNaN(config.To) || math.IsInf(config.To, 0) ||
		math.IsNaN(config.Duration) || math.IsInf(config.Duration, 0) || config.Duration < 0 {
		return nil, fmt.Errorf("timeline: invalid cue ramp")
	}
	duration := config.Duration
	if duration == 0 {
		duration = clock.windows[config.Window].Duration
	}
	if duration <= 0 {
		return nil, fmt.Errorf("timeline: cue ramp needs a finite duration")
	}
	return &CueRamp{clock: clock, window: config.Window, from: config.From, to: config.To, duration: duration}, nil
}

func (ramp *CueRamp) Value() float64 {
	progress := min(1, max(0, ramp.clock.Elapsed(ramp.window)/ramp.duration))
	return ramp.from + (ramp.to-ramp.from)*progress
}
