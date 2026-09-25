package motion

import (
	"fmt"
	"math"
)

// IndexedHarmonic contributes a sine or cosine to one coordinate. The phase
// is (clock + index*IndexPhase)*Rate + index*IndexRate + Phase. SecondaryClock
// selects clocks[1]; otherwise clocks[0] is used. Envelope multiplies the
// amplitude by the caller's current envelope value.
type IndexedHarmonic struct {
	Amplitude, Rate, IndexPhase, IndexRate, Phase float64
	Cos                                           bool
	SecondaryClock                                bool
	Envelope                                      bool
}

// FormationBounds optionally restricts the final sprite position. Bounds are
// applied after all harmonics and the per-instance screen-space spacing.
type FormationBounds struct{ Min, Max Point }

// HarmonicFormationConfig describes a reusable sprite or glyph trajectory.
// Each axis can combine any number of independently phased harmonics. The two
// clocks and envelope are supplied per update, so they can follow audio or an
// authored motion controller without changing this serializable geometry.
type HarmonicFormationConfig struct {
	Origin, Spacing Point
	X, Y            []IndexedHarmonic
	Bounds          *FormationBounds
}

// HarmonicFormation owns a validated copy of its terms and samples without
// allocation. It can drive sprites.Group or another renderer.
type HarmonicFormation struct{ config HarmonicFormationConfig }

func NewHarmonicFormation(config HarmonicFormationConfig) (*HarmonicFormation, error) {
	finite := func(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
	if len(config.X) > 64 || len(config.Y) > 64 {
		return nil, fmt.Errorf("motion: too many formation harmonics")
	}
	for _, value := range []float64{config.Origin.X, config.Origin.Y, config.Spacing.X, config.Spacing.Y} {
		if !finite(value) {
			return nil, fmt.Errorf("motion: nonfinite formation origin or spacing")
		}
	}
	for _, term := range append(append([]IndexedHarmonic(nil), config.X...), config.Y...) {
		for _, value := range []float64{term.Amplitude, term.Rate, term.IndexPhase, term.IndexRate, term.Phase} {
			if !finite(value) {
				return nil, fmt.Errorf("motion: nonfinite formation harmonic")
			}
		}
	}
	if config.Bounds != nil {
		b := *config.Bounds
		if !finite(b.Min.X) || !finite(b.Min.Y) || !finite(b.Max.X) || !finite(b.Max.Y) || b.Min.X > b.Max.X || b.Min.Y > b.Max.Y {
			return nil, fmt.Errorf("motion: invalid formation bounds")
		}
		config.Bounds = &b
	}
	config.X = append([]IndexedHarmonic(nil), config.X...)
	config.Y = append([]IndexedHarmonic(nil), config.Y...)
	return &HarmonicFormation{config: config}, nil
}

// At returns one position. Clocks and envelope are caller-owned simulation
// values; they need not use seconds as their unit.
func (formation *HarmonicFormation) At(index int, clocks [2]float64, envelope float64) Point {
	config := formation.config
	p := Point{X: config.Origin.X + float64(index)*config.Spacing.X, Y: config.Origin.Y + float64(index)*config.Spacing.Y}
	p.X = sampleFormationAxis(p.X, config.X, index, clocks, envelope)
	p.Y = sampleFormationAxis(p.Y, config.Y, index, clocks, envelope)
	if config.Bounds != nil {
		p.X = min(max(p.X, config.Bounds.Min.X), config.Bounds.Max.X)
		p.Y = min(max(p.Y, config.Bounds.Min.Y), config.Bounds.Max.Y)
	}
	return p
}

func sampleFormationAxis(value float64, terms []IndexedHarmonic, index int, clocks [2]float64, envelope float64) float64 {
	for _, term := range terms {
		clock := clocks[0]
		if term.SecondaryClock {
			clock = clocks[1]
		}
		phase := (clock+float64(index)*term.IndexPhase)*term.Rate + float64(index)*term.IndexRate + term.Phase
		wave := math.Sin(phase)
		if term.Cos {
			wave = math.Cos(phase)
		}
		amplitude := term.Amplitude
		if term.Envelope {
			amplitude *= envelope
		}
		value += amplitude * wave
	}
	return value
}
