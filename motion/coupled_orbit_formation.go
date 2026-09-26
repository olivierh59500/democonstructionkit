package motion

import (
	"fmt"
	"math"
)

// OrbitRange visits Count consecutive slot indices with an editable step.
// Multiple ranges retain authored interleaving and gaps between sprite trains.
type OrbitRange struct {
	Start, Count, Step int
}

// CoupledOrbitFormationConfig binds a stateful CoupledOrbit to an ordered slot
// program. The orbit advances once for every sampled slot, not once per frame.
type CoupledOrbitFormationConfig struct {
	Orbit  CoupledOrbit
	Ranges []OrbitRange
}

// CoupledOrbitFormation keeps a bounded pose bank for repeated sprite drawing.
type CoupledOrbitFormation struct {
	orbit   CoupledOrbit
	initial CoupledOrbit
	indices []float64
	poses   []Point
}

func NewCoupledOrbitFormation(c CoupledOrbitFormationConfig) (*CoupledOrbitFormation, error) {
	for _, value := range [...]float64{c.Orbit.XIncrement, c.Orbit.YIncrement, c.Orbit.ZIncrement,
		c.Orbit.QIncrement, c.Orbit.XOffset, c.Orbit.YOffset, c.Orbit.ZOffset,
		c.Orbit.QOffset, c.Orbit.QScale, c.Orbit.CenterX, c.Orbit.CenterY,
		c.Orbit.Radius, c.Orbit.DepthRadius, c.Orbit.PhaseStep} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite coupled orbit setting")
		}
	}
	if len(c.Ranges) == 0 || len(c.Ranges) > 256 {
		return nil, fmt.Errorf("motion: invalid coupled orbit ranges")
	}
	count := 0
	for _, span := range c.Ranges {
		if span.Count < 1 || span.Step == 0 || span.Count > 1_000_000-count {
			return nil, fmt.Errorf("motion: invalid coupled orbit range")
		}
		last := float64(span.Start) + float64(span.Count-1)*float64(span.Step)
		if math.IsInf(last, 0) || math.Abs(last) > 1<<52 {
			return nil, fmt.Errorf("motion: coupled orbit index exceeds exact range")
		}
		count += span.Count
	}
	formation := &CoupledOrbitFormation{
		orbit: c.Orbit, initial: c.Orbit,
		indices: make([]float64, 0, count), poses: make([]Point, count),
	}
	for _, span := range c.Ranges {
		for i := 0; i < span.Count; i++ {
			formation.indices = append(formation.indices, float64(span.Start)+float64(i)*float64(span.Step))
		}
	}
	return formation, nil
}

// Step prepares every pose in range order. Drawing or inspecting poses never
// changes the orbit phase, so the same formation may appear on several layers.
func (f *CoupledOrbitFormation) Step() []Point {
	if f == nil {
		return nil
	}
	for i, index := range f.indices {
		f.poses[i].X, f.poses[i].Y = f.orbit.Next(index)
	}
	return f.poses
}

func (f *CoupledOrbitFormation) Poses() []Point { return f.poses }
func (f *CoupledOrbitFormation) Len() int       { return len(f.indices) }

// Orbit exposes the owned controller so input or audio cues can edit its
// increments and offsets without restarting the accumulated phases.
func (f *CoupledOrbitFormation) Orbit() *CoupledOrbit { return &f.orbit }

// Reset restores the initial phase and authored control values.
func (f *CoupledOrbitFormation) Reset() {
	if f == nil {
		return
	}
	f.orbit = f.initial
	clear(f.poses)
}
