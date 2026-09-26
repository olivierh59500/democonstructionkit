package motion

import (
	"fmt"
	"math"
)

// FormulaTrajectoryConfig describes an image, sprite or text position driven
// by two independently advancing clocks. Zero Period leaves a clock unwrapped.
// A clock only advances when Step is called, so scene stages can pause it.
type FormulaTrajectoryConfig struct {
	X, Y                FormulaExpr
	Start, Step, Period [2]float64
}

// FormulaTrajectory owns compiled coordinate formulas and their clock state.
// Sampling and advancing do not allocate and never depend on image dimensions.
type FormulaTrajectory struct {
	x, y                *FormulaProgram
	start, step, period [2]float64
	clock               [2]float64
}

func NewFormulaTrajectory(config FormulaTrajectoryConfig) (*FormulaTrajectory, error) {
	for _, values := range [...][2]float64{config.Start, config.Step, config.Period} {
		for _, value := range values {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, fmt.Errorf("motion: nonfinite formula trajectory clock")
			}
		}
	}
	if config.Period[0] < 0 || config.Period[1] < 0 {
		return nil, fmt.Errorf("motion: negative formula trajectory period")
	}
	x, err := CompileFormula(config.X)
	if err != nil {
		return nil, err
	}
	y, err := CompileFormula(config.Y)
	if err != nil {
		return nil, err
	}
	return &FormulaTrajectory{x: x, y: y, start: config.Start, step: config.Step, period: config.Period, clock: config.Start}, nil
}

// At returns the current position before the next Step. Index and dimensions
// remain editable inputs for formations that share the same trajectory.
func (trajectory *FormulaTrajectory) At(index, width, height, count float64) Point {
	a, b := trajectory.clock[0], trajectory.clock[1]
	return Point{
		X: trajectory.x.AtWithSecondaryTime(a, b, index, width, height, count),
		Y: trajectory.y.AtWithSecondaryTime(a, b, index, width, height, count),
	}
}

// Position samples one image without formation index or viewport inputs.
func (trajectory *FormulaTrajectory) Position() Point { return trajectory.At(0, 0, 0, 0) }

// Step advances each clock once, applying an optional exact modulo wrap.
func (trajectory *FormulaTrajectory) Step() {
	for axis := range trajectory.clock {
		value := trajectory.clock[axis] + trajectory.step[axis]
		if period := trajectory.period[axis]; period > 0 {
			value = math.Mod(value, period)
		}
		trajectory.clock[axis] = value
	}
}

// Clocks reports the phases used by the next At call.
func (trajectory *FormulaTrajectory) Clocks() [2]float64 { return trajectory.clock }

// Reset restores both initial phases without changing the configured speeds.
func (trajectory *FormulaTrajectory) Reset() { trajectory.clock = trajectory.start }

// SetStep changes the next clock increments, for example from a music cue.
func (trajectory *FormulaTrajectory) SetStep(step [2]float64) error {
	for _, value := range step {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("motion: nonfinite formula trajectory step")
		}
	}
	trajectory.step = step
	return nil
}
