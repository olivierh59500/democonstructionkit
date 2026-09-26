package motion

import (
	"fmt"
	"math"
)

// CoupledLogoConfig describes two depth-linked image paths. The secondary
// logo inherits the primary Y and depth, then adds its own harmonic motion.
// Phase steps are radians per simulation update.
type CoupledLogoConfig struct {
	PhasePrimary, PhaseSecondary float64
	StepPrimary, StepSecondary   float64
	SpeedMultiplier              float64
	DepthBase, DepthSin          float64
	DepthDivisor                 float64
	YBase, YSumAmplitude         float64
	YSin, YCos                   float64
	SecondaryDepthCos            float64
	SecondaryYSin                float64
}

type CoupledLogoPose struct{ Y, Depth float64 }

// CoupledLogoMotion owns phases but no image resources. DrawOrder paints the
// smaller depth first, with the secondary logo in front on equal depths.
type CoupledLogoMotion struct {
	config             CoupledLogoConfig
	primary, secondary float64
	speed              float64
}

func NewCoupledLogoMotion(c CoupledLogoConfig) (*CoupledLogoMotion, error) {
	for _, v := range [...]float64{c.PhasePrimary, c.PhaseSecondary, c.StepPrimary, c.StepSecondary,
		c.SpeedMultiplier, c.DepthBase, c.DepthSin, c.DepthDivisor,
		c.YBase, c.YSumAmplitude, c.YSin, c.YCos,
		c.SecondaryDepthCos, c.SecondaryYSin} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("motion: nonfinite coupled logo parameter")
		}
	}
	if c.SpeedMultiplier < 0 || c.DepthDivisor <= 0 {
		return nil, fmt.Errorf("motion: invalid coupled logo speed or depth divisor")
	}
	return &CoupledLogoMotion{config: c, primary: c.PhasePrimary,
		secondary: c.PhaseSecondary, speed: c.SpeedMultiplier}, nil
}

func (m *CoupledLogoMotion) Step() error {
	a := m.primary + m.config.StepPrimary*m.speed
	b := m.secondary + m.config.StepSecondary*m.speed
	if math.IsNaN(a) || math.IsInf(a, 0) || math.IsNaN(b) || math.IsInf(b, 0) {
		return fmt.Errorf("motion: coupled logo phase overflow")
	}
	m.primary, m.secondary = a, b
	return nil
}

func (m *CoupledLogoMotion) SetSpeedMultiplier(speed float64) error {
	if math.IsNaN(speed) || math.IsInf(speed, 0) || speed < 0 {
		return fmt.Errorf("motion: invalid coupled logo speed")
	}
	m.speed = speed
	return nil
}

func (m *CoupledLogoMotion) Phases() (primary, secondary float64) {
	return m.primary, m.secondary
}

func (m *CoupledLogoMotion) Poses() [2]CoupledLogoPose {
	c := m.config
	sinPrimary, cosPrimary := math.Sin(m.primary), math.Cos(m.primary)
	sinSecondary, cosSecondary := math.Sin(m.secondary), math.Cos(m.secondary)
	depth := (c.DepthBase + c.DepthSin*sinPrimary) / c.DepthDivisor
	y := c.YBase + (cosPrimary+sinPrimary)*c.YSumAmplitude +
		c.YCos*cosPrimary + c.YSin*sinPrimary
	return [2]CoupledLogoPose{{Y: y, Depth: depth},
		{Y: y + sinSecondary*(c.SecondaryYSin*depth), Depth: depth + c.SecondaryDepthCos*cosSecondary}}
}

func (m *CoupledLogoMotion) DrawOrder() [2]int {
	poses := m.Poses()
	if poses[1].Depth >= poses[0].Depth {
		return [2]int{0, 1}
	}
	return [2]int{1, 0}
}
