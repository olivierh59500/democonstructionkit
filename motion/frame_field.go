package motion

import (
	"fmt"
	"math"
)

// FrameParticle is one sprite position, velocity, material and fractional
// atlas-frame clock. Velocity and Rate are per simulation update.
type FrameParticle struct {
	X, Y, VelocityX, VelocityY, Phase, Rate float64
	Image                                   int
}

// FrameAxisWrap subtracts Shift after crossing Boundary. Positive Shift tests
// the upper boundary; negative Shift tests the lower boundary. OnWrap can
// change the other coordinate or material while preserving overshoot.
type FrameAxisWrap struct {
	Boundary, Shift float64
	Inclusive       bool
	OnWrap          func(index int, particle *FrameParticle)
}

// FrameFieldConfig describes a bounded population of independently animated
// sprites. Spawn receives reset=false at construction and true after a sprite
// reaches EndPhase. It may use a seeded random source or authored positions.
type FrameFieldConfig struct {
	Count        int
	EndPhase     float64 // Zero disables atlas-frame completion and respawn.
	Spawn        func(index int, reset bool) FrameParticle
	WrapX, WrapY *FrameAxisWrap
}

// FrameField owns animation state without depending on a font or image atlas.
// A renderer may draw the same particles with pixels, images or vector shapes.
type FrameField struct {
	config FrameFieldConfig
	items  []FrameParticle
	speed  float64
}

func NewFrameField(c FrameFieldConfig) (*FrameField, error) {
	if c.Count < 1 || c.Count > 1<<16 || !finiteFrameField(c.EndPhase) || c.EndPhase < 0 || c.Spawn == nil {
		return nil, fmt.Errorf("motion: invalid frame field")
	}
	for _, wrap := range [...]*FrameAxisWrap{c.WrapX, c.WrapY} {
		if wrap != nil && (!finiteFrameField(wrap.Boundary) || !finiteFrameField(wrap.Shift) || wrap.Shift == 0) {
			return nil, fmt.Errorf("motion: invalid frame field wrap")
		}
	}
	if c.WrapX != nil {
		copy := *c.WrapX
		c.WrapX = &copy
	}
	if c.WrapY != nil {
		copy := *c.WrapY
		c.WrapY = &copy
	}
	f := &FrameField{config: c, items: make([]FrameParticle, c.Count), speed: 1}
	if err := f.Reset(); err != nil {
		return nil, err
	}
	return f, nil
}

func finiteFrameField(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func validFrameParticle(p FrameParticle, animated bool) bool {
	return finiteFrameField(p.X) && finiteFrameField(p.Y) &&
		finiteFrameField(p.VelocityX) && finiteFrameField(p.VelocityY) &&
		finiteFrameField(p.Phase) && finiteFrameField(p.Rate) &&
		p.Rate >= 0 && (!animated || p.Rate > 0)
}

// Reset reinitializes the population in index order. A source can deliberately
// start beyond EndPhase; the first Step then respawns that particle exactly once.
func (f *FrameField) Reset() error {
	for i := range f.items {
		p := f.config.Spawn(i, false)
		if !validFrameParticle(p, f.config.EndPhase > 0) {
			return fmt.Errorf("motion: invalid initial frame particle %d", i)
		}
		f.items[i] = p
	}
	return nil
}

// Step advances one simulation tick and respawns completed particles in index
// order. It neither allocates nor changes the frame clock while drawing.
func (f *FrameField) Step() error {
	for i := range f.items {
		p := &f.items[i]
		p.X += p.VelocityX * f.speed
		p.Y += p.VelocityY * f.speed
		if f.config.WrapX != nil {
			if err := f.wrap(p, i, f.config.WrapX, true); err != nil {
				return err
			}
		}
		if f.config.WrapY != nil {
			if err := f.wrap(p, i, f.config.WrapY, false); err != nil {
				return err
			}
		}
		p.Phase += p.Rate * f.speed
		if f.config.EndPhase > 0 && p.Phase >= f.config.EndPhase {
			next := f.config.Spawn(i, true)
			if !validFrameParticle(next, true) {
				return fmt.Errorf("motion: invalid respawned frame particle %d", i)
			}
			*p = next
		}
	}
	return nil
}

func (f *FrameField) wrap(p *FrameParticle, index int, rule *FrameAxisWrap, horizontal bool) error {
	if rule == nil {
		return nil
	}
	coordinate := &p.X
	if !horizontal {
		coordinate = &p.Y
	}
	value := *coordinate
	if rule.Shift > 0 && (value > rule.Boundary || rule.Inclusive && value == rule.Boundary) ||
		rule.Shift < 0 && (value < rule.Boundary || rule.Inclusive && value == rule.Boundary) {
		*coordinate -= rule.Shift
		if rule.OnWrap != nil {
			rule.OnWrap(index, p)
		}
		if !validFrameParticle(*p, f.config.EndPhase > 0) {
			return fmt.Errorf("motion: invalid wrapped frame particle %d", index)
		}
	}
	return nil
}

// SetSpeedMultiplier allows timeline or music cues to change every particle's
// playback speed without resetting its position or individual rate.
func (f *FrameField) SetSpeedMultiplier(speed float64) error {
	if !finiteFrameField(speed) || speed < 0 {
		return fmt.Errorf("motion: invalid frame field speed")
	}
	f.speed = speed
	return nil
}

func (f *FrameField) Count() int { return len(f.items) }

// Samples returns borrowed state until the next Step or Reset. Do not mutate it.
func (f *FrameField) Samples() []FrameParticle { return f.items }
