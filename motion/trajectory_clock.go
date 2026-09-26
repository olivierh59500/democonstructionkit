package motion

import (
	"fmt"
	"math"
)

// TrajectoryClockConfig combines a phase-parameterized 2D path with its update
// cadence. Sample may be a built-in path's At method or an authored function.
// Start is sampled immediately; Step advances before the next sample is read.
type TrajectoryClockConfig struct {
	Sample func(float64) Point
	Start  float64
	Step   float64
}

// TrajectoryClock samples a path once per update and retains its last point.
// Reading At repeatedly, for example to draw a sprite into several layers,
// does not advance the phase or recalculate the trajectory.
type TrajectoryClock struct {
	config TrajectoryClockConfig
	phase  float64
	point  Point
}

func NewTrajectoryClock(config TrajectoryClockConfig) (*TrajectoryClock, error) {
	if config.Sample == nil || !trajectoryFinite(config.Start) || !trajectoryFinite(config.Step) {
		return nil, fmt.Errorf("motion: invalid trajectory clock setting")
	}
	point := config.Sample(config.Start)
	if !trajectoryFinite(point.X) || !trajectoryFinite(point.Y) {
		return nil, fmt.Errorf("motion: nonfinite trajectory point")
	}
	return &TrajectoryClock{config: config, phase: config.Start, point: point}, nil
}

func trajectoryFinite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

// At returns the cached point. Call Step before drawing when the original scene
// advances its phase before sampling; otherwise draw first and Step afterwards.
func (clock *TrajectoryClock) At() Point { return clock.point }

func (clock *TrajectoryClock) Phase() float64 { return clock.phase }

// Step preserves the old pose if the next phase or sampled point is invalid.
func (clock *TrajectoryClock) Step() error {
	next := clock.phase + clock.config.Step
	if !trajectoryFinite(next) {
		return fmt.Errorf("motion: nonfinite trajectory phase")
	}
	point := clock.config.Sample(next)
	if !trajectoryFinite(point.X) || !trajectoryFinite(point.Y) {
		return fmt.Errorf("motion: nonfinite trajectory point")
	}
	clock.phase, clock.point = next, point
	return nil
}

// SetStep changes tempo without changing the current phase or pose.
func (clock *TrajectoryClock) SetStep(step float64) error {
	if !trajectoryFinite(step) {
		return fmt.Errorf("motion: nonfinite trajectory step")
	}
	clock.config.Step = step
	return nil
}

// SetSample changes the trajectory at the current phase. A scene can use it to
// change a logo's path without restarting the animation clock.
func (clock *TrajectoryClock) SetSample(sample func(float64) Point) error {
	if sample == nil {
		return fmt.Errorf("motion: nil trajectory sample")
	}
	point := sample(clock.phase)
	if !trajectoryFinite(point.X) || !trajectoryFinite(point.Y) {
		return fmt.Errorf("motion: nonfinite trajectory point")
	}
	clock.config.Sample, clock.point = sample, point
	return nil
}

// SetPhase seeks to another phase without altering the step or path.
func (clock *TrajectoryClock) SetPhase(phase float64) error {
	if !trajectoryFinite(phase) {
		return fmt.Errorf("motion: nonfinite trajectory phase")
	}
	point := clock.config.Sample(phase)
	if !trajectoryFinite(point.X) || !trajectoryFinite(point.Y) {
		return fmt.Errorf("motion: nonfinite trajectory point")
	}
	clock.phase, clock.point = phase, point
	return nil
}

// Reset returns to the authored starting phase while retaining tempo and path.
func (clock *TrajectoryClock) Reset() error { return clock.SetPhase(clock.config.Start) }
