package modulation

import (
	"fmt"
	"math"
)

// PeakBankConfig defines independent bounded peaks with linear decay. Decay
// uses units per supplied delta; delta=1 retains an exact per-update clock.
type PeakBankConfig struct {
	Count                 int
	Floor, Ceiling, Decay float64
}

// PeakCharge raises one peak to at least Level. A level outside the configured
// range is clamped; duplicate indices combine by their maximum level.
type PeakCharge struct {
	Index int
	Level float64
}

// PeakBank retains reusable storage for independent, variably charged peaks.
// Unlike a fixed retriggered envelope, charging does not lower an existing peak.
type PeakBank struct {
	config PeakBankConfig
	levels []float64
}

func NewPeakBank(c PeakBankConfig) (*PeakBank, error) {
	if c.Count < 1 || c.Count > 65536 || !finite(c.Floor) || !finite(c.Ceiling) ||
		!finite(c.Decay) || c.Ceiling < c.Floor || c.Decay < 0 || !finite(c.Ceiling-c.Floor) {
		return nil, fmt.Errorf("modulation: invalid peak bank")
	}
	p := &PeakBank{config: c, levels: make([]float64, c.Count)}
	p.Reset()
	return p, nil
}

// Charge immediately raises one peak without applying decay. Invalid indices
// and nonfinite levels return false and leave the bank unchanged.
func (p *PeakBank) Charge(index int, level float64) bool {
	if p == nil || index < 0 || index >= len(p.levels) || !finite(level) {
		return false
	}
	p.levels[index] = math.Max(p.levels[index], math.Max(p.config.Floor, math.Min(p.config.Ceiling, level)))
	return true
}

// Step charges every supplied peak, then decays all peaks during the same step.
// Validation completes before changing any level. No storage is allocated.
func (p *PeakBank) Step(charges []PeakCharge, delta float64) error {
	if p == nil || !finite(delta) || delta < 0 {
		return fmt.Errorf("modulation: invalid peak bank step")
	}
	for _, charge := range charges {
		if charge.Index < 0 || charge.Index >= len(p.levels) || !finite(charge.Level) {
			return fmt.Errorf("modulation: invalid peak charge")
		}
	}
	for _, charge := range charges {
		p.Charge(charge.Index, charge.Level)
	}
	span := p.config.Ceiling - p.config.Floor
	decay := 0.0
	if p.config.Decay > 0 {
		if delta > span/p.config.Decay {
			decay = span
		} else {
			decay = p.config.Decay * delta
		}
	}
	for index, level := range p.levels {
		p.levels[index] = math.Max(p.config.Floor, level-decay)
	}
	return nil
}

// Level returns zero for an unavailable peak. Valid peaks retain Floor.
func (p *PeakBank) Level(index int) float64 {
	if p == nil || index < 0 || index >= len(p.levels) {
		return 0
	}
	return p.levels[index]
}

func (p *PeakBank) Len() int {
	if p == nil {
		return 0
	}
	return len(p.levels)
}

func (p *PeakBank) Reset() {
	if p != nil {
		for index := range p.levels {
			p.levels[index] = p.config.Floor
		}
	}
}
