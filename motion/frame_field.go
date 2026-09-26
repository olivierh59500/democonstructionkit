package motion

import (
	"fmt"
	"math"
)

// FrameParticle is one sprite position and fractional atlas-frame clock.
// Rate is frames per simulation update, independent for each particle.
type FrameParticle struct {
	X, Y, Phase, Rate float64
}

// FrameFieldConfig describes a bounded population of independently animated
// sprites. Spawn receives reset=false at construction and true after a sprite
// reaches EndPhase. It may use a seeded random source or authored positions.
type FrameFieldConfig struct {
	Count    int
	EndPhase float64
	Spawn    func(index int, reset bool) FrameParticle
}

// FrameField owns animation state without depending on a font or image atlas.
// A renderer may draw the same particles with pixels, images or vector shapes.
type FrameField struct {
	config FrameFieldConfig
	items  []FrameParticle
	speed  float64
}

func NewFrameField(c FrameFieldConfig) (*FrameField, error) {
	if c.Count < 1 || c.Count > 1<<16 || !finiteFrameField(c.EndPhase) || c.EndPhase <= 0 || c.Spawn == nil {
		return nil, fmt.Errorf("motion: invalid frame field")
	}
	f := &FrameField{config: c, items: make([]FrameParticle, c.Count), speed: 1}
	if err := f.Reset(); err != nil {
		return nil, err
	}
	return f, nil
}

func finiteFrameField(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func validFrameParticle(p FrameParticle) bool {
	return finiteFrameField(p.X) && finiteFrameField(p.Y) && finiteFrameField(p.Phase) &&
		finiteFrameField(p.Rate) && p.Rate > 0
}

// Reset reinitializes the population in index order. A source can deliberately
// start beyond EndPhase; the first Step then respawns that particle exactly once.
func (f *FrameField) Reset() error {
	for i := range f.items {
		p := f.config.Spawn(i, false)
		if !validFrameParticle(p) {
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
		p.Phase += p.Rate * f.speed
		if p.Phase >= f.config.EndPhase {
			next := f.config.Spawn(i, true)
			if !validFrameParticle(next) {
				return fmt.Errorf("motion: invalid respawned frame particle %d", i)
			}
			*p = next
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
