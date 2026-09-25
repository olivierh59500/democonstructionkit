package motion

import (
	"fmt"
	"math"
)

// FormationHarmonic moves a formation from its first to its last item with
// independently chosen amplitudes. Cycles is the number of sine cycles during
// one cue; IndexPhase adds a phase offset per item, in radians.
type FormationHarmonic struct {
	FirstAmplitude, LastAmplitude float64
	Cycles, Phase, IndexPhase     float64
}

// FormationCue applies a finite motion to a formation. LeadIndex moves first;
// each adjacent item starts Stagger seconds later. Multiple cues may overlap.
type FormationCue struct {
	Start, Duration, Stagger, Fade float64
	LeadIndex                      int
	X, Y                           []FormationHarmonic
}

// CuedFormationConfig arranges any number of images along a resting row or
// column, then layers reusable, staggered motion cues over that arrangement.
// A positive Loop repeats the entire choreography in seconds.
type CuedFormationConfig struct {
	Origin, Spacing Point
	Count           int
	Loop            float64
	Cues            []FormationCue
}

// CuedFormation contains validated, caller-independent choreography data.
type CuedFormation struct{ config CuedFormationConfig }

func NewCuedFormation(config CuedFormationConfig) (*CuedFormation, error) {
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	if config.Count < 1 || config.Count > 1_000_000 || len(config.Cues) > 64 {
		return nil, fmt.Errorf("motion: invalid formation size or cue count")
	}
	for _, value := range []float64{config.Origin.X, config.Origin.Y, config.Spacing.X, config.Spacing.Y, config.Loop} {
		if !finite(value) {
			return nil, fmt.Errorf("motion: nonfinite formation parameter")
		}
	}
	if config.Loop < 0 {
		return nil, fmt.Errorf("motion: negative formation loop")
	}
	copyConfig := config
	copyConfig.Cues = make([]FormationCue, len(config.Cues))
	for index, cue := range config.Cues {
		if cue.Start < 0 || cue.Duration <= 0 || cue.Stagger < 0 || cue.Fade < 0 || cue.Fade*2 > cue.Duration || cue.LeadIndex < 0 || cue.LeadIndex >= config.Count || len(cue.X) > 16 || len(cue.Y) > 16 {
			return nil, fmt.Errorf("motion: invalid formation cue %d", index)
		}
		for _, value := range []float64{cue.Start, cue.Duration, cue.Stagger, cue.Fade} {
			if !finite(value) {
				return nil, fmt.Errorf("motion: nonfinite formation cue %d", index)
			}
		}
		for _, term := range append(append([]FormationHarmonic(nil), cue.X...), cue.Y...) {
			for _, value := range []float64{term.FirstAmplitude, term.LastAmplitude, term.Cycles, term.Phase, term.IndexPhase} {
				if !finite(value) {
					return nil, fmt.Errorf("motion: nonfinite formation harmonic")
				}
			}
		}
		maximumDelay := math.Max(float64(cue.LeadIndex), float64(config.Count-1-cue.LeadIndex)) * cue.Stagger
		if config.Loop > 0 && cue.Start+cue.Duration+maximumDelay > config.Loop {
			return nil, fmt.Errorf("motion: formation cue %d crosses loop boundary", index)
		}
		cue.X = append([]FormationHarmonic(nil), cue.X...)
		cue.Y = append([]FormationHarmonic(nil), cue.Y...)
		copyConfig.Cues[index] = cue
	}
	return &CuedFormation{config: copyConfig}, nil
}

// At returns the top-left or anchor position for one item at an absolute time.
// It performs no allocation and returns the resting position between cues.
func (formation *CuedFormation) At(seconds float64, index int) Point {
	if formation == nil || index < 0 || index >= formation.config.Count || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return Point{}
	}
	c := formation.config
	if c.Loop > 0 {
		seconds = Wrap(seconds, c.Loop)
	}
	position := Point{X: c.Origin.X + float64(index)*c.Spacing.X, Y: c.Origin.Y + float64(index)*c.Spacing.Y}
	for _, cue := range c.Cues {
		local := seconds - cue.Start - math.Abs(float64(index-cue.LeadIndex))*cue.Stagger
		if local < 0 || local >= cue.Duration {
			continue
		}
		progress := local / cue.Duration
		weight := 1.0
		if cue.Fade > 0 {
			weight = Smooth(local/cue.Fade) * Smooth((cue.Duration-local)/cue.Fade)
		}
		position.X += weight * sampleFormationHarmonics(cue.X, progress, index, c.Count)
		position.Y += weight * sampleFormationHarmonics(cue.Y, progress, index, c.Count)
	}
	return position
}

func sampleFormationHarmonics(terms []FormationHarmonic, progress float64, index, count int) float64 {
	amount := 0.0
	fraction := 0.0
	if count > 1 {
		fraction = float64(index) / float64(count-1)
	}
	for _, term := range terms {
		amplitude := Lerp(term.FirstAmplitude, term.LastAmplitude, fraction)
		phase := term.Phase + float64(index)*term.IndexPhase
		amount += amplitude * (math.Sin(2*math.Pi*term.Cycles*progress+phase) - math.Sin(phase))
	}
	return amount
}
