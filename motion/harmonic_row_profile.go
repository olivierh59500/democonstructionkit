package motion

import (
	"fmt"
	"math"
)

// HarmonicRowStage adds one existing harmonic formation to a row's position.
// Nil Index samples the current row. A nonnil Index selects a fixed item,
// allowing a global offset to be sampled once and shared by every row.
// Harmonics within a stage are summed first; stages are added in caller order.
type HarmonicRowStage struct {
	Motion HarmonicFormationConfig
	Index  *int
}

// HarmonicRowProfileConfig defines a bounded, cached row recipe. ClockScale
// converts supplied seconds into the formation's two clocks; zero keeps that
// clock fixed. Separate stages preserve arithmetic grouping between global
// harmonics, row spacing and local harmonics without changing the sine engine.
type HarmonicRowProfileConfig struct {
	Count      int
	ClockScale [2]float64
	Stages     []HarmonicRowStage
}

type harmonicRowStage struct {
	motion *HarmonicFormation
	values []Point
	index  int
	fixed  bool
}

// HarmonicRowProfile owns copied formation recipes and cached stage samples.
// Apply can be shared by several materials at the same time without resampling
// their waves. Like other mutable clocks, it belongs to one rendering goroutine.
type HarmonicRowProfile struct {
	count      int
	clockScale [2]float64
	stages     []harmonicRowStage
	timeBits   uint64
	valid      bool
}

func NewHarmonicRowProfile(c HarmonicRowProfileConfig) (*HarmonicRowProfile, error) {
	if c.Count < 1 || c.Count > 65536 || len(c.Stages) < 1 || len(c.Stages) > 64 ||
		c.Count > (1<<20)/len(c.Stages) || !pathFinite(c.ClockScale[0]) || !pathFinite(c.ClockScale[1]) {
		return nil, fmt.Errorf("motion: invalid harmonic row profile")
	}
	p := &HarmonicRowProfile{count: c.Count, clockScale: c.ClockScale, stages: make([]harmonicRowStage, len(c.Stages))}
	for index, stage := range c.Stages {
		formation, err := NewHarmonicFormation(stage.Motion)
		if err != nil {
			return nil, err
		}
		cached := harmonicRowStage{motion: formation}
		count := c.Count
		if stage.Index != nil {
			if *stage.Index < -1_000_000 || *stage.Index > 1_000_000 {
				return nil, fmt.Errorf("motion: invalid fixed harmonic row index")
			}
			cached.index, cached.fixed = *stage.Index, true
			count = 1
		}
		cached.values = make([]Point, count)
		p.stages[index] = cached
	}
	return p, nil
}

// Sample fills the stage cache once per distinct time. Fixed-index stages
// evaluate once, regardless of row count. The harmonic envelope is one.
func (p *HarmonicRowProfile) Sample(seconds float64) error {
	if p == nil || !pathFinite(seconds) {
		return fmt.Errorf("motion: invalid harmonic row time")
	}
	bits := math.Float64bits(seconds)
	if p.valid && bits == p.timeBits {
		return nil
	}
	clocks := [2]float64{seconds * p.clockScale[0], seconds * p.clockScale[1]}
	if !pathFinite(clocks[0]) || !pathFinite(clocks[1]) {
		return fmt.Errorf("motion: harmonic row clock overflows")
	}
	p.valid = false
	for _, stage := range p.stages {
		for row := range stage.values {
			index := row
			if stage.fixed {
				index = stage.index
			}
			value := stage.motion.At(index, clocks, 1)
			if !pathFinite(value.X) || !pathFinite(value.Y) {
				return fmt.Errorf("motion: harmonic row sample overflows")
			}
			stage.values[row] = value
		}
	}
	p.timeBits, p.valid = bits, true
	return nil
}

// Apply adds the cached stages to base in their original order. Supplying an
// image vertex in base preserves placement before deformation. A false result
// rejects unavailable rows, nonfinite inputs or overflowing output coordinates.
func (p *HarmonicRowProfile) Apply(row int, seconds float64, base Point) (Point, bool) {
	if p == nil || row < 0 || row >= p.count || !pathFinite(base.X) || !pathFinite(base.Y) || p.Sample(seconds) != nil {
		return Point{}, false
	}
	for _, stage := range p.stages {
		index := row
		if stage.fixed {
			index = 0
		}
		base.X += stage.values[index].X
		base.Y += stage.values[index].Y
		if !pathFinite(base.X) || !pathFinite(base.Y) {
			return Point{}, false
		}
	}
	return base, true
}

func (p *HarmonicRowProfile) Len() int {
	if p == nil {
		return 0
	}
	return p.count
}
