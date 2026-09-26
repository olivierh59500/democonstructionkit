package motion

import (
	"fmt"
	"math"
)

// RecurrentTranslationConfig applies a fixed phase step to X/Y harmonic terms.
// ReanchorEvery recomputes exact phases periodically to bound recurrence drift;
// zero keeps stepping without reanchoring. It is suitable for synchronized
// sprite formations whose common translation advances once per logical tick.
type RecurrentTranslationConfig struct {
	Harmonics     HarmonicTranslation
	PhaseStep     float64
	ReanchorEvery uint64
	XStepDeltas   []float64 // Optional exact per-term steps for authored clocks.
	YStepDeltas   []float64
}

type recurrentTranslationTerm struct {
	spec                    HarmonicTerm
	delta, sinStep, cosStep float64
	sin, cos                float64
}

// RecurrentTranslation computes shared harmonic motion with no per-tick
// trigonometric calls except at the configured reanchor boundary.
type RecurrentTranslation struct {
	x, y          []recurrentTranslationTerm
	reanchorEvery uint64
	tick          uint64
	point         Point
}

func NewRecurrentTranslation(c RecurrentTranslationConfig) (*RecurrentTranslation, error) {
	if math.IsNaN(c.PhaseStep) || math.IsInf(c.PhaseStep, 0) || len(c.Harmonics.X) > 64 || len(c.Harmonics.Y) > 64 ||
		len(c.XStepDeltas) != 0 && len(c.XStepDeltas) != len(c.Harmonics.X) ||
		len(c.YStepDeltas) != 0 && len(c.YStepDeltas) != len(c.Harmonics.Y) {
		return nil, fmt.Errorf("motion: invalid recurrent translation configuration")
	}
	makeTerms := func(specs []HarmonicTerm, exact []float64) ([]recurrentTranslationTerm, error) {
		terms := make([]recurrentTranslationTerm, len(specs))
		for i, spec := range specs {
			for _, value := range [...]float64{spec.Amplitude, spec.Rate, spec.Phase} {
				if math.IsNaN(value) || math.IsInf(value, 0) {
					return nil, fmt.Errorf("motion: nonfinite recurrent harmonic")
				}
			}
			delta := c.PhaseStep * spec.Rate
			if len(exact) != 0 {
				delta = exact[i]
			}
			if math.IsNaN(delta) || math.IsInf(delta, 0) {
				return nil, fmt.Errorf("motion: recurrent phase step overflows")
			}
			terms[i] = recurrentTranslationTerm{spec: spec, delta: delta}
			terms[i].sinStep, terms[i].cosStep = math.Sincos(delta)
			terms[i].sin, terms[i].cos = math.Sincos(spec.Phase)
		}
		return terms, nil
	}
	x, err := makeTerms(c.Harmonics.X, c.XStepDeltas)
	if err != nil {
		return nil, err
	}
	y, err := makeTerms(c.Harmonics.Y, c.YStepDeltas)
	if err != nil {
		return nil, err
	}
	r := &RecurrentTranslation{x: x, y: y, reanchorEvery: c.ReanchorEvery}
	r.sample()
	return r, nil
}

// Step advances every term together. The resulting point may be reused by any
// number of sprites or layers before the next simulation update.
func (r *RecurrentTranslation) Step() Point {
	if r == nil {
		return Point{}
	}
	r.tick++
	reanchor := r.reanchorEvery > 0 && r.tick%r.reanchorEvery == 0
	advance := func(terms []recurrentTranslationTerm) {
		for i := range terms {
			term := &terms[i]
			if reanchor {
				phase := term.spec.Phase + float64(r.tick)*term.delta
				term.sin, term.cos = math.Sincos(math.Mod(phase, 2*math.Pi))
			} else {
				term.sin, term.cos = term.sin*term.cosStep+term.cos*term.sinStep,
					term.cos*term.cosStep-term.sin*term.sinStep
			}
		}
	}
	advance(r.x)
	advance(r.y)
	r.sample()
	return r.point
}

// At returns the current shared translation without advancing its clocks.
func (r *RecurrentTranslation) At() Point {
	if r == nil {
		return Point{}
	}
	return r.point
}

// Tick reports the number of logical advances since construction or Reset.
func (r *RecurrentTranslation) Tick() uint64 { return r.tick }

// Reset restores the authored initial phases and first visible translation.
func (r *RecurrentTranslation) Reset() {
	if r == nil {
		return
	}
	r.tick = 0
	for _, terms := range [...][]recurrentTranslationTerm{r.x, r.y} {
		for i := range terms {
			terms[i].sin, terms[i].cos = math.Sincos(terms[i].spec.Phase)
		}
	}
	r.sample()
}

func (r *RecurrentTranslation) sample() {
	r.point = Point{}
	for _, term := range r.x {
		wave := term.sin
		if term.spec.Cos {
			wave = term.cos
		}
		r.point.X += term.spec.Amplitude * wave
	}
	for _, term := range r.y {
		wave := term.sin
		if term.spec.Cos {
			wave = term.cos
		}
		r.point.Y += term.spec.Amplitude * wave
	}
}
