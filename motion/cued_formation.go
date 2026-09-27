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

// FormationPoseKey sets the origin, inter-item spacing and arch at an absolute
// time. Arc offsets the middle item most and leaves the endpoints in place.
// Ease controls interpolation to the next pose; keys may describe a complete
// formation even while independent harmonic cues move individual items.
type FormationPoseKey struct {
	Time                 float64
	Origin, Spacing, Arc Point
	Ease                 Ease
}

// CuedFormationConfig arranges any number of images along a resting row or
// column, then layers reusable, staggered motion cues over that arrangement.
// A positive Loop repeats the entire choreography in seconds.
type CuedFormationConfig struct {
	Origin, Spacing, Arc Point
	Count                int
	Loop                 float64
	Cues                 []FormationCue
	// OriginKeys, SpacingKeys and ArcKeys replace the resting pose at chosen times.
	// Their absolute Point values interpolate before harmonic cues are added.
	OriginKeys, SpacingKeys, ArcKeys []Key[Point]
	// PoseKeys is a concise alternative for changing all three properties together.
	// It cannot be combined with the independent property keys.
	PoseKeys []FormationPoseKey
}

// CuedFormation contains validated, caller-independent choreography data.
type CuedFormation struct {
	config       CuedFormationConfig
	originTrack  *Track[Point]
	spacingTrack *Track[Point]
	arcTrack     *Track[Point]
}

func NewCuedFormation(config CuedFormationConfig) (*CuedFormation, error) {
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	if config.Count < 1 || config.Count > 1_000_000 || len(config.Cues) > 64 {
		return nil, fmt.Errorf("motion: invalid formation size or cue count")
	}
	for _, value := range []float64{config.Origin.X, config.Origin.Y, config.Spacing.X, config.Spacing.Y, config.Arc.X, config.Arc.Y, config.Loop} {
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
	if len(config.PoseKeys) > 0 {
		if len(config.OriginKeys) > 0 || len(config.SpacingKeys) > 0 || len(config.ArcKeys) > 0 {
			return nil, fmt.Errorf("motion: combined and independent formation keys cannot be mixed")
		}
		if len(config.PoseKeys) > 4096 {
			return nil, fmt.Errorf("motion: too many formation pose keys")
		}
		copyConfig.OriginKeys = make([]Key[Point], len(config.PoseKeys))
		copyConfig.SpacingKeys = make([]Key[Point], len(config.PoseKeys))
		copyConfig.ArcKeys = make([]Key[Point], len(config.PoseKeys))
		for i, pose := range config.PoseKeys {
			copyConfig.OriginKeys[i] = Key[Point]{Time: pose.Time, Value: pose.Origin, Ease: pose.Ease}
			copyConfig.SpacingKeys[i] = Key[Point]{Time: pose.Time, Value: pose.Spacing, Ease: pose.Ease}
			copyConfig.ArcKeys[i] = Key[Point]{Time: pose.Time, Value: pose.Arc, Ease: pose.Ease}
		}
	}
	originTrack, err := formationTrack(copyConfig.OriginKeys, config.Loop)
	if err != nil {
		return nil, fmt.Errorf("motion: invalid formation origin track: %w", err)
	}
	spacingTrack, err := formationTrack(copyConfig.SpacingKeys, config.Loop)
	if err != nil {
		return nil, fmt.Errorf("motion: invalid formation spacing track: %w", err)
	}
	arcTrack, err := formationTrack(copyConfig.ArcKeys, config.Loop)
	if err != nil {
		return nil, fmt.Errorf("motion: invalid formation arc track: %w", err)
	}
	copyConfig.OriginKeys, copyConfig.SpacingKeys, copyConfig.ArcKeys, copyConfig.PoseKeys = nil, nil, nil, nil
	return &CuedFormation{config: copyConfig, originTrack: originTrack, spacingTrack: spacingTrack, arcTrack: arcTrack}, nil
}

func formationTrack(keys []Key[Point], loop float64) (*Track[Point], error) {
	if len(keys) == 0 {
		return nil, nil
	}
	if len(keys) > 4096 {
		return nil, fmt.Errorf("too many formation keys")
	}
	for _, key := range keys {
		if key.Time < 0 || loop > 0 && key.Time > loop ||
			math.IsNaN(key.Value.X) || math.IsInf(key.Value.X, 0) ||
			math.IsNaN(key.Value.Y) || math.IsInf(key.Value.Y, 0) {
			return nil, fmt.Errorf("nonfinite or out-of-range formation key")
		}
	}
	track, err := NewTrack(keys, func(a, b Point, fraction float64) Point {
		return Point{X: Lerp(a.X, b.X, fraction), Y: Lerp(a.Y, b.Y, fraction)}
	})
	if err != nil {
		return nil, err
	}
	if loop > 0 && track.At(0) != track.At(loop) {
		return nil, fmt.Errorf("formation key track jumps at loop boundary")
	}
	return track, nil
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
	origin, spacing, arc := c.Origin, c.Spacing, c.Arc
	if formation.originTrack != nil {
		origin = formation.originTrack.At(seconds)
	}
	if formation.spacingTrack != nil {
		spacing = formation.spacingTrack.At(seconds)
	}
	if formation.arcTrack != nil {
		arc = formation.arcTrack.At(seconds)
	}
	arcWeight := 0.0
	if c.Count > 1 && (arc.X != 0 || arc.Y != 0) {
		arcWeight = math.Sin(math.Pi * float64(index) / float64(c.Count-1))
	}
	position := Point{X: origin.X + float64(index)*spacing.X + arc.X*arcWeight, Y: origin.Y + float64(index)*spacing.Y + arc.Y*arcWeight}
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
