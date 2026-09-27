package authoring

import (
	"fmt"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func formationEase(name string) (motion.Ease, error) {
	switch name {
	case "", "linear":
		return motion.Linear, nil
	case "smooth":
		return motion.Smooth, nil
	default:
		return nil, fmt.Errorf("unknown formation ease %q", name)
	}
}

func formationPointKeys(keys []FormationPointKey) ([]motion.Key[motion.Point], error) {
	result := make([]motion.Key[motion.Point], len(keys))
	for i, key := range keys {
		ease, err := formationEase(key.Ease)
		if err != nil {
			return nil, fmt.Errorf("formation key %d: %w", i, err)
		}
		result[i] = motion.Key[motion.Point]{Time: key.Time, Value: point(key.Value), Ease: ease}
	}
	return result, nil
}

// compileFormation maps editor data to the same formation engine used by Go
// scenes. Construction validates bounds, cue delays and loop continuity.
func compileFormation(spec CuedFormation, count int) (*motion.CuedFormation, error) {
	if len(spec.PoseKeys) > 4096 || len(spec.OriginKeys) > 4096 ||
		len(spec.SpacingKeys) > 4096 || len(spec.ArcKeys) > 4096 || len(spec.Cues) > 64 {
		return nil, fmt.Errorf("too many formation keys or cues")
	}
	config := motion.CuedFormationConfig{
		Origin: point(spec.Origin), Spacing: point(spec.Spacing), Arc: point(spec.Arc),
		Count: count, Loop: spec.Loop,
	}
	for i, pose := range spec.PoseKeys {
		ease, err := formationEase(pose.Ease)
		if err != nil {
			return nil, fmt.Errorf("formation pose %d: %w", i, err)
		}
		config.PoseKeys = append(config.PoseKeys, motion.FormationPoseKey{
			Time: pose.Time, Origin: point(pose.Origin), Spacing: point(pose.Spacing),
			Arc: point(pose.Arc), Ease: ease,
		})
	}
	var err error
	if config.OriginKeys, err = formationPointKeys(spec.OriginKeys); err != nil {
		return nil, err
	}
	if config.SpacingKeys, err = formationPointKeys(spec.SpacingKeys); err != nil {
		return nil, err
	}
	if config.ArcKeys, err = formationPointKeys(spec.ArcKeys); err != nil {
		return nil, err
	}
	for _, cue := range spec.Cues {
		if len(cue.X) > 16 || len(cue.Y) > 16 {
			return nil, fmt.Errorf("too many formation harmonics")
		}
		compiled := motion.FormationCue{
			Start: cue.Start, Duration: cue.Duration, Stagger: cue.Stagger,
			Fade: cue.Fade, LeadIndex: cue.LeadIndex,
		}
		for _, harmonic := range cue.X {
			compiled.X = append(compiled.X, motion.FormationHarmonic{
				FirstAmplitude: harmonic.FirstAmplitude, LastAmplitude: harmonic.LastAmplitude,
				Cycles: harmonic.Cycles, Phase: harmonic.Phase, IndexPhase: harmonic.IndexPhase,
			})
		}
		for _, harmonic := range cue.Y {
			compiled.Y = append(compiled.Y, motion.FormationHarmonic{
				FirstAmplitude: harmonic.FirstAmplitude, LastAmplitude: harmonic.LastAmplitude,
				Cycles: harmonic.Cycles, Phase: harmonic.Phase, IndexPhase: harmonic.IndexPhase,
			})
		}
		config.Cues = append(config.Cues, compiled)
	}
	return motion.NewCuedFormation(config)
}

func compileSampledFormation(spec SampledFormation, count int) (*motion.KeyframedFormation, error) {
	if len(spec.Frames) > 4096 || count < 1 || count > 10000 ||
		len(spec.Frames) > 0 && count > 1_048_576/len(spec.Frames) {
		return nil, fmt.Errorf("too many sampled formation points")
	}
	config := motion.KeyframedFormationConfig{Count: count, Loop: spec.Loop}
	for i, frame := range spec.Frames {
		ease, err := formationEase(frame.Ease)
		if err != nil {
			return nil, fmt.Errorf("sampled formation frame %d: %w", i, err)
		}
		points := make([]motion.Point, len(frame.Points))
		for j, v := range frame.Points {
			points[j] = point(v)
		}
		config.Frames = append(config.Frames, motion.FormationFrame{Time: frame.Time, Points: points, Ease: ease})
	}
	return motion.NewKeyframedFormation(config)
}
