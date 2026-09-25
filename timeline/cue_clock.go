package timeline

import (
	"fmt"
	"math"
	"time"
)

// CueWindow accumulates seconds from a chosen logical tick. Duration is the
// completion threshold; zero keeps the window open. Tolerance is an optional
// positive comparison margin for historical floating-point clocks.
type CueWindow struct {
	StartTick           int
	Duration, Tolerance float64
}

// CueClockConfig describes one tick clock and independent overlapping windows.
// Rate can change during playback without resetting accumulated window time.
type CueClockConfig struct {
	Rate    int
	Windows []CueWindow
}

// CueClock owns a fixed-tick counter and per-window elapsed seconds. Step tests
// active windows before incrementing Tick, preserving end-of-frame boundaries.
type CueClock struct {
	rate    int
	tick    int
	windows []CueWindow
	elapsed []float64
}

func NewCueClock(config CueClockConfig) (*CueClock, error) {
	if config.Rate < 1 || config.Rate > 1_000_000 || len(config.Windows) > 1024 {
		return nil, fmt.Errorf("timeline: invalid cue clock rate or window count")
	}
	for _, window := range config.Windows {
		if math.IsNaN(window.Duration) || math.IsInf(window.Duration, 0) || math.IsNaN(window.Tolerance) || math.IsInf(window.Tolerance, 0) || window.Duration < 0 || window.Tolerance < 0 {
			return nil, fmt.Errorf("timeline: invalid cue window duration")
		}
	}
	return &CueClock{rate: config.Rate, windows: append([]CueWindow(nil), config.Windows...), elapsed: make([]float64, len(config.Windows))}, nil
}

func (clock *CueClock) Rate() int { return clock.rate }
func (clock *CueClock) Tick() int { return clock.tick }

func (clock *CueClock) SetRate(rate int) error {
	if rate < 1 || rate > 1_000_000 {
		return fmt.Errorf("timeline: invalid cue clock rate")
	}
	clock.rate = rate
	return nil
}

// Step accumulates one rate-dependent interval in windows already active at
// the current tick, then increments the tick. It allocates nothing.
func (clock *CueClock) Step() {
	delta := 1 / float64(clock.rate)
	for index, window := range clock.windows {
		if clock.tick >= window.StartTick {
			clock.elapsed[index] += delta
		}
	}
	clock.tick++
}

func (clock *CueClock) Active(index int) bool     { return clock.tick >= clock.windows[index].StartTick }
func (clock *CueClock) Elapsed(index int) float64 { return clock.elapsed[index] }
func (clock *CueClock) Done(index int) bool {
	window := clock.windows[index]
	return window.Duration > 0 && clock.Active(index) && clock.elapsed[index]+window.Tolerance >= window.Duration
}

// Reached compares a wall-clock duration with Tick/Rate using integer parts,
// retaining exact threshold behavior without floating-point accumulation.
func (clock *CueClock) Reached(duration time.Duration) bool {
	if duration <= 0 {
		return true
	}
	whole, remainder := clock.tick/clock.rate, clock.tick%clock.rate
	wantWhole, wantRemainder := duration/time.Second, duration%time.Second
	if int64(whole) != int64(wantWhole) {
		return int64(whole) > int64(wantWhole)
	}
	return int64(remainder)*int64(time.Second) >= int64(wantRemainder)*int64(clock.rate)
}

func (clock *CueClock) Reset() {
	clock.tick = 0
	clear(clock.elapsed)
}
