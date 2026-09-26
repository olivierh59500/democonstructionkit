package motion

import (
	"fmt"
	"math"
)

// CaptionCycleConfig moves one line between two vertical limits. InitialWait
// can hold the first hidden line longer than the repeating HoldWait period.
type CaptionCycleConfig struct {
	Count                 int
	Top, Bottom, StartY   float64
	Speed                 float64
	InitialWait, HoldWait int
}

type CaptionPose struct {
	Index    int
	Y        float64
	Wait     int
	Velocity float64
}

// CaptionCycle prepares the next pose after a line has been drawn. Drawing
// before Step preserves the initial hidden line and exact arrival/exit holds.
type CaptionCycle struct {
	config CaptionCycleConfig
	pose   CaptionPose
}

func NewCaptionCycle(c CaptionCycleConfig) (*CaptionCycle, error) {
	if c.Count < 1 || c.Count > 1<<16 || c.Top >= c.Bottom ||
		math.IsNaN(c.Top) || math.IsInf(c.Top, 0) ||
		math.IsNaN(c.Bottom) || math.IsInf(c.Bottom, 0) ||
		math.IsNaN(c.StartY) || math.IsInf(c.StartY, 0) ||
		math.IsNaN(c.Speed) || math.IsInf(c.Speed, 0) || c.Speed <= 0 ||
		c.InitialWait < 1 || c.HoldWait < 1 || c.InitialWait < c.HoldWait {
		return nil, fmt.Errorf("motion: invalid caption cycle")
	}
	m := &CaptionCycle{config: c}
	m.Reset()
	return m, nil
}

func (m *CaptionCycle) Reset() {
	m.pose = CaptionPose{Y: m.config.StartY, Wait: m.config.InitialWait,
		Velocity: m.config.Speed}
}

func (m *CaptionCycle) Step() {
	p := &m.pose
	if p.Wait == m.config.HoldWait {
		p.Y += p.Velocity
	}
	if p.Y >= m.config.Bottom {
		p.Y = m.config.Bottom
		p.Wait--
		if p.Wait <= 0 {
			p.Velocity = -m.config.Speed
			p.Wait = m.config.HoldWait
		}
	}
	if p.Y <= m.config.Top {
		p.Y = m.config.Top
		p.Wait--
		if p.Wait <= 0 {
			p.Velocity = m.config.Speed
			p.Wait = m.config.HoldWait
			p.Index = (p.Index + 1) % m.config.Count
		}
	}
}

func (m *CaptionCycle) Pose() CaptionPose { return m.pose }
