package motion

import (
	"fmt"
	"math"
)

// BounceToggleConfig oscillates one scalar and optionally advances a material
// index when either bound is crossed. Strict and inclusive bounds, overshoot
// and clamping are independent so the same clock can drive a logo or sprite.
type BounceToggleConfig struct {
	Start, Velocity, Min, Max  float64
	Inclusive, Clamp           bool
	ToggleLower, ToggleUpper   bool
	Materials, InitialMaterial int
}

type BounceTogglePose struct {
	Value, Velocity float64
	Material        int
}

// BounceToggle stores the material separately from the sign of its scale.
// This keeps a selected face through the return trip until the next authored
// boundary, including an overshoot frame with a negative scale.
type BounceToggle struct {
	config BounceToggleConfig
	pose   BounceTogglePose
}

func NewBounceToggle(c BounceToggleConfig) (*BounceToggle, error) {
	for _, value := range [...]float64{c.Start, c.Velocity, c.Min, c.Max} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite bounce toggle parameter")
		}
	}
	if c.Min >= c.Max || c.Materials < 1 || c.Materials > 1<<16 ||
		c.InitialMaterial < 0 || c.InitialMaterial >= c.Materials ||
		c.Start < c.Min || c.Start > c.Max || c.Velocity == 0 {
		return nil, fmt.Errorf("motion: invalid bounce toggle")
	}
	clock := &BounceToggle{config: c}
	clock.Reset()
	return clock, nil
}

func (b *BounceToggle) Reset() {
	b.pose = BounceTogglePose{Value: b.config.Start, Velocity: b.config.Velocity,
		Material: b.config.InitialMaterial}
}

// Step advances after drawing the current pose. Lower and upper checks run in
// that order, matching authored effects with two independent boundary tests.
func (b *BounceToggle) Step() {
	b.pose.Value += b.pose.Velocity
	low := b.pose.Value < b.config.Min || b.config.Inclusive && b.pose.Value <= b.config.Min
	if low {
		b.pose.Velocity = -b.pose.Velocity
		if b.config.ToggleLower {
			b.pose.Material = (b.pose.Material + 1) % b.config.Materials
		}
		if b.config.Clamp {
			b.pose.Value = b.config.Min
		}
	}
	high := b.pose.Value > b.config.Max || b.config.Inclusive && b.pose.Value >= b.config.Max
	if high {
		b.pose.Velocity = -b.pose.Velocity
		if b.config.ToggleUpper {
			b.pose.Material = (b.pose.Material + 1) % b.config.Materials
		}
		if b.config.Clamp {
			b.pose.Value = b.config.Max
		}
	}
}

func (b *BounceToggle) Pose() BounceTogglePose { return b.pose }
