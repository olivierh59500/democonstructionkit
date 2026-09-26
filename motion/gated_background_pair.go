package motion

import (
	"fmt"
	"math"
)

// ThresholdAxis changes velocity when its current position is outside a
// configured interval, then applies exactly one movement step. It preserves
// overshoot and the historical check-before-move order.
type ThresholdAxis struct {
	Start, Velocity              float64
	Lower, Upper                 float64
	VelocityBelow, VelocityAbove float64
}

// GatedBackgroundPairConfig synchronizes one vertically bouncing background
// with a later-activated horizontal axis and a second coupled X/Y backdrop.
// The second X speed changes when the second Y crosses either boundary.
type GatedBackgroundPairConfig struct {
	GateStart, GateStep, GateOpen, GateReset float64
	FirstX, FirstY                           ThresholdAxis
	SecondY                                  ThresholdAxis
	SecondXStart, SecondXVelocity            float64
	SecondXMin, SecondXMax                   float64
	SecondXBelowVelocity                     float64
	SecondXAboveVelocity                     float64
}

// GatedBackgroundPair owns only the two positions and their animation clocks.
type GatedBackgroundPair struct {
	config                GatedBackgroundPairConfig
	gate                  float64
	firstX, firstY        ThresholdAxis
	secondY               ThresholdAxis
	secondX, secondXSpeed float64
}

func NewGatedBackgroundPair(c GatedBackgroundPairConfig) (*GatedBackgroundPair, error) {
	for _, value := range [...]float64{c.GateStart, c.GateStep, c.GateOpen, c.GateReset,
		c.SecondXStart, c.SecondXVelocity, c.SecondXMin, c.SecondXMax,
		c.SecondXBelowVelocity, c.SecondXAboveVelocity} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite background pair parameter")
		}
	}
	if c.GateStep < 0 || c.GateReset <= c.GateOpen || c.SecondXMin >= c.SecondXMax {
		return nil, fmt.Errorf("motion: invalid background pair gate or bounds")
	}
	for _, axis := range [...]ThresholdAxis{c.FirstX, c.FirstY, c.SecondY} {
		if axis.Lower >= axis.Upper {
			return nil, fmt.Errorf("motion: invalid background axis bounds")
		}
		for _, value := range [...]float64{axis.Start, axis.Velocity, axis.Lower, axis.Upper,
			axis.VelocityBelow, axis.VelocityAbove} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, fmt.Errorf("motion: nonfinite background axis")
			}
		}
	}
	pair := &GatedBackgroundPair{config: c}
	pair.Reset()
	return pair, nil
}

func (p *GatedBackgroundPair) Reset() {
	p.gate = p.config.GateStart
	p.firstX, p.firstY, p.secondY = p.config.FirstX, p.config.FirstY, p.config.SecondY
	p.secondX, p.secondXSpeed = p.config.SecondXStart, p.config.SecondXVelocity
}

func advanceThresholdAxis(axis ThresholdAxis) ThresholdAxis {
	if axis.Start < axis.Lower {
		axis.Velocity = axis.VelocityBelow
	}
	if axis.Start > axis.Upper {
		axis.Velocity = axis.VelocityAbove
	}
	axis.Start += axis.Velocity
	return axis
}

// Step reproduces gate, first backdrop, gate-reset and second-backdrop order.
func (p *GatedBackgroundPair) Step() error {
	gate := p.gate + p.config.GateStep
	firstY := advanceThresholdAxis(p.firstY)
	firstX := p.firstX
	if gate > p.config.GateOpen {
		firstX = advanceThresholdAxis(firstX)
	}
	if gate > p.config.GateReset {
		gate = 0
	}
	secondY := p.secondY
	secondXSpeed := p.secondXSpeed
	if secondY.Start < secondY.Lower {
		secondY.Velocity = secondY.VelocityBelow
		secondXSpeed = p.config.SecondXBelowVelocity
	}
	if secondY.Start > secondY.Upper {
		secondY.Velocity = secondY.VelocityAbove
		secondXSpeed = p.config.SecondXAboveVelocity
	}
	secondX := p.secondX + secondXSpeed
	secondX = math.Max(p.config.SecondXMin, math.Min(p.config.SecondXMax, secondX))
	secondY.Start += secondY.Velocity
	for _, value := range [...]float64{gate, firstX.Start, firstY.Start, secondX, secondY.Start} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("motion: background pair clock overflow")
		}
	}
	p.gate, p.firstX, p.firstY = gate, firstX, firstY
	p.secondX, p.secondXSpeed, p.secondY = secondX, secondXSpeed, secondY
	return nil
}

func (p *GatedBackgroundPair) Poses() [2]Point {
	return [2]Point{{X: p.firstX.Start, Y: p.firstY.Start},
		{X: p.secondX, Y: p.secondY.Start}}
}

func (p *GatedBackgroundPair) GatePhase() float64 { return p.gate }
