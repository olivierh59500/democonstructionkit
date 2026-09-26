package geometry

import (
	"fmt"
	"math"
)

// PointShapeBank supplies authored shapes without transferring ownership of
// its templates. Clone returns a fresh writable shape at a shape-change cue;
// Target returns a read-only template for a morph.
type PointShapeBank interface {
	Clone(name string) (PointWriter, error)
	Target(name string) (PointReader, error)
}

type PointAnimationKind uint8

const (
	PointAnimSinusGrid PointAnimationKind = iota
	PointAnimMorph
	PointAnimRotors
	PointAnimYOrbit
	PointAnimBounce
)

// PointAnimationSpec selects one reusable deformation or positional override.
// A morph with Frames zero inherits the scaled duration of its owning action.
type PointAnimationSpec struct {
	Kind      PointAnimationKind
	Target    string
	SinusGrid SinusGridConfig
	Morph     PointMorphConfig
	Rotors    RotorConfig
	YOrbit    YOrbitConfig
	Bounce    BounceCurveConfig
}

// PointAction changes only explicitly supplied fields. SetAnimations=false
// inherits the active programs; true with an empty slice clears them. DropLast
// and AppendAnimations run after that selection, before the optional OnEnter.
type PointAction struct {
	Frames           int
	Shape            string
	SetShape         bool
	Position         *Vec3
	InitialRotation  *Vec3
	RotationStep     *Vec3
	TranslationStep  *Vec3
	TextIndex        *int
	SetAnimations    bool
	Animations       []PointAnimationSpec
	DropLast         int
	AppendAnimations []PointAnimationSpec
	OnEnter          func(*PointSequence) error
}

// PointSequenceConfig keeps the stage program separate from production art and
// point storage. DurationScale converts authored stage frames to logical ticks.
type PointSequenceConfig struct {
	Shapes           PointShapeBank
	Actions          []PointAction
	DurationScale    float64
	Loop             bool
	InitialPosition  Vec3
	InitialRotation  Vec3
	InitialTextIndex int
}

// PointSceneState is the currently displayed model pose and caption cue.
// Points is borrowed from the shape bank; only Clone may replace it.
type PointSceneState struct {
	Points          PointWriter
	Position        Vec3
	Rotation        Vec3
	RotationStep    Vec3
	TranslationStep Vec3
	TextIndex       int
}

type pointProgram interface{ Step(*PointSceneState) }

type sinusGridProgram struct{ grid *SinusGrid }

func (p sinusGridProgram) Step(state *PointSceneState) { p.grid.Step(state.Points) }

type morphProgram struct{ morph *PointMorph }

func (p morphProgram) Step(state *PointSceneState) { p.morph.Step(state.Points) }

type rotorProgram struct{ rotors *Rotors }

func (p rotorProgram) Step(state *PointSceneState) { p.rotors.Step(state.Points) }

type orbitProgram struct{ orbit *YOrbit }

func (p orbitProgram) Step(state *PointSceneState) { state.Position = p.orbit.Step() }

type bounceProgram struct{ bounce *BounceCurve }

func (p bounceProgram) Step(state *PointSceneState) { state.Position.Y = p.bounce.Step() }

// PointSequence owns stage timing, inherited pose and ordered animation
// programs. It does not allocate or draw during ordinary Step calls.
type PointSequence struct {
	config     PointSequenceConfig
	state      PointSceneState
	animations []pointProgram
	action     int
	remaining  int
	finished   bool
}

func NewPointSequence(c PointSequenceConfig) (*PointSequence, error) {
	if c.Shapes == nil || len(c.Actions) == 0 || len(c.Actions) > 100_000 ||
		math.IsNaN(c.DurationScale) || math.IsInf(c.DurationScale, 0) || c.DurationScale <= 0 ||
		!finiteSceneVector(c.InitialPosition) || !finiteSceneVector(c.InitialRotation) {
		return nil, fmt.Errorf("geometry: invalid point sequence configuration")
	}
	actions := make([]PointAction, len(c.Actions))
	for i, action := range c.Actions {
		scaledFrames := float64(action.Frames) * c.DurationScale
		if action.Frames < 1 || action.Frames > 1_000_000 || math.IsNaN(scaledFrames) || math.IsInf(scaledFrames, 0) ||
			scaledFrames < 1 || scaledFrames >= float64(int(^uint(0)>>1)) ||
			action.DropLast < 0 || action.DropLast > 1_000_000 || action.SetShape && action.Shape == "" {
			return nil, fmt.Errorf("geometry: invalid point action %d", i)
		}
		for _, vector := range [...]*Vec3{action.Position, action.InitialRotation, action.RotationStep, action.TranslationStep} {
			if vector != nil && !finiteSceneVector(*vector) {
				return nil, fmt.Errorf("geometry: nonfinite point action %d", i)
			}
		}
		if action.SetShape {
			if _, err := c.Shapes.Target(action.Shape); err != nil {
				return nil, fmt.Errorf("geometry: point action %d shape: %w", i, err)
			}
		}
		for _, spec := range append(append([]PointAnimationSpec(nil), action.Animations...), action.AppendAnimations...) {
			if err := validatePointAnimation(c.Shapes, spec, int(scaledFrames)); err != nil {
				return nil, fmt.Errorf("geometry: point action %d animation: %w", i, err)
			}
		}
		actions[i] = clonePointAction(action)
	}
	c.Actions = actions
	s := &PointSequence{config: c}
	if err := s.Reset(); err != nil {
		return nil, err
	}
	return s, nil
}

func finiteSceneVector(v Vec3) bool {
	return !math.IsNaN(v.X) && !math.IsInf(v.X, 0) && !math.IsNaN(v.Y) && !math.IsInf(v.Y, 0) &&
		!math.IsNaN(v.Z) && !math.IsInf(v.Z, 0)
}

func clonePointAction(a PointAction) PointAction {
	copyVector := func(v *Vec3) *Vec3 {
		if v == nil {
			return nil
		}
		copied := *v
		return &copied
	}
	a.Position = copyVector(a.Position)
	a.InitialRotation = copyVector(a.InitialRotation)
	a.RotationStep = copyVector(a.RotationStep)
	a.TranslationStep = copyVector(a.TranslationStep)
	if a.TextIndex != nil {
		index := *a.TextIndex
		a.TextIndex = &index
	}
	copySpecs := func(specs []PointAnimationSpec) []PointAnimationSpec {
		result := append([]PointAnimationSpec(nil), specs...)
		for i := range result {
			result[i].Rotors.Radii = append([]float64(nil), result[i].Rotors.Radii...)
			result[i].Bounce.Samples = append([]float64(nil), result[i].Bounce.Samples...)
		}
		return result
	}
	a.Animations = copySpecs(a.Animations)
	a.AppendAnimations = copySpecs(a.AppendAnimations)
	return a
}

func validatePointAnimation(bank PointShapeBank, spec PointAnimationSpec, stageFrames int) error {
	switch spec.Kind {
	case PointAnimSinusGrid:
		_, err := NewSinusGrid(spec.SinusGrid)
		return err
	case PointAnimMorph:
		if spec.Target == "" || spec.Morph.Tail > MorphRepeat || spec.Morph.Frames < 0 || spec.Morph.Frames > 1_000_000 {
			return fmt.Errorf("geometry: invalid morph target or policy")
		}
		if _, err := bank.Target(spec.Target); err != nil {
			return err
		}
		if spec.Morph.Frames == 0 && stageFrames < 1 {
			return fmt.Errorf("geometry: invalid morph duration")
		}
		return nil
	case PointAnimRotors:
		_, err := NewRotors(spec.Rotors)
		return err
	case PointAnimYOrbit:
		_, err := NewYOrbit(spec.YOrbit)
		return err
	case PointAnimBounce:
		_, err := NewBounceCurve(spec.Bounce)
		return err
	default:
		return fmt.Errorf("geometry: unknown point animation %d", spec.Kind)
	}
}

// Reset selects the first stage from the configured initial state. Shape
// templates are cloned again, so a replay does not retain old point mutations.
func (s *PointSequence) Reset() error {
	if s == nil {
		return fmt.Errorf("geometry: nil point sequence")
	}
	s.state = PointSceneState{Position: s.config.InitialPosition, Rotation: s.config.InitialRotation,
		TextIndex: s.config.InitialTextIndex}
	s.animations = s.animations[:0]
	s.action, s.remaining, s.finished = -1, 0, false
	return s.enterNext()
}

func (s *PointSequence) enterNext() error {
	next := s.action + 1
	if next >= len(s.config.Actions) {
		if !s.config.Loop {
			s.finished = true
			return nil
		}
		next = 0
	}
	a := s.config.Actions[next]
	if a.SetShape {
		points, err := s.config.Shapes.Clone(a.Shape)
		if err != nil {
			return err
		}
		s.state.Points = points
	}
	if a.Position != nil {
		s.state.Position = *a.Position
	}
	if a.InitialRotation != nil {
		s.state.Rotation = *a.InitialRotation
	}
	if a.RotationStep != nil {
		s.state.RotationStep = *a.RotationStep
	}
	if a.TranslationStep != nil {
		s.state.TranslationStep = *a.TranslationStep
	}
	if a.TextIndex != nil {
		s.state.TextIndex = *a.TextIndex
	}
	frames := int(float64(a.Frames) * s.config.DurationScale)
	if a.SetAnimations {
		s.animations = s.animations[:0]
		for _, spec := range a.Animations {
			program, err := s.newProgram(spec, frames)
			if err != nil {
				return err
			}
			s.animations = append(s.animations, program)
		}
	}
	s.animations = s.animations[:len(s.animations)-min(a.DropLast, len(s.animations))]
	for _, spec := range a.AppendAnimations {
		program, err := s.newProgram(spec, frames)
		if err != nil {
			return err
		}
		s.animations = append(s.animations, program)
	}
	s.action, s.remaining = next, frames
	if a.OnEnter != nil {
		return a.OnEnter(s)
	}
	return nil
}

func (s *PointSequence) newProgram(spec PointAnimationSpec, stageFrames int) (pointProgram, error) {
	switch spec.Kind {
	case PointAnimSinusGrid:
		program, err := NewSinusGrid(spec.SinusGrid)
		return sinusGridProgram{program}, err
	case PointAnimMorph:
		if s.state.Points == nil {
			return nil, fmt.Errorf("geometry: morph needs current points")
		}
		target, err := s.config.Shapes.Target(spec.Target)
		if err != nil {
			return nil, err
		}
		config := spec.Morph
		if config.Frames == 0 {
			config.Frames = stageFrames
		}
		program, err := NewPointMorph(s.state.Points, target, config)
		return morphProgram{program}, err
	case PointAnimRotors:
		program, err := NewRotors(spec.Rotors)
		return rotorProgram{program}, err
	case PointAnimYOrbit:
		program, err := NewYOrbit(spec.YOrbit)
		return orbitProgram{program}, err
	case PointAnimBounce:
		program, err := NewBounceCurve(spec.Bounce)
		return bounceProgram{program}, err
	default:
		return nil, fmt.Errorf("geometry: unknown point animation %d", spec.Kind)
	}
}

// Step preserves the authored boundary rule: a completed stage enters the
// next one without also moving or deforming the displayed points on that tick.
func (s *PointSequence) Step() error {
	if s == nil {
		return fmt.Errorf("geometry: nil point sequence")
	}
	if s.finished {
		return nil
	}
	s.remaining--
	if s.remaining <= 0 {
		return s.enterNext()
	}
	s.state.Rotation = s.state.Rotation.Add(s.state.RotationStep)
	s.state.Position = s.state.Position.Add(s.state.TranslationStep)
	for _, animation := range s.animations {
		animation.Step(&s.state)
	}
	return nil
}

// State returns the current pose and borrowed point collection without stepping.
func (s *PointSequence) State() PointSceneState { return s.state }

// ActionIndex and Remaining expose the current stage and its pending ticks.
func (s *PointSequence) ActionIndex() int { return s.action }
func (s *PointSequence) Remaining() int   { return s.remaining }

// AnimationCount reports the number of currently inherited or selected programs.
func (s *PointSequence) AnimationCount() int { return len(s.animations) }

// Finished reports a non-looping sequence after its final boundary.
func (s *PointSequence) Finished() bool { return s.finished }
