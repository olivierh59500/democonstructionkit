package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/modulation"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// GroupSignals bind common visual properties to time, keyframes, beats or named
// audio inputs. Nil means identity (zero offset/angle, unit scale/opacity).
type GroupSignals struct {
	X, Y, ScaleX, ScaleY, Angle, Opacity *modulation.Signal
}

// GridFormation adds a centered row/column offset before shared motion.
// Rows defaults to the number required by Count when zero.
type GridFormation struct {
	Columns, Rows int
	StepX, StepY  float64
}

// GroupConfig describes a sprite/logo formation, independent of image metrics.
// Choose at most one Path, Points, Orbit, Weave, Circle, Harmonic, Formula, Coupled or Formation. With none, positions follow
// Velocity. Speed and PhaseSpacing use path pixels or the orbit's phase units;
// Delay is seconds per instance. Spacing is an additional screen-space offset.
// Points offers serializable path data; SplineSamples>0 selects a smooth spline.
type GroupConfig struct {
	Frames                                  []*ebiten.Image `json:"-"`
	Count                                   int
	FPS                                     float64
	FrameStride, FrameOffset                int
	Path                                    *motion.Path `json:"-"`
	Points                                  []motion.Point
	Closed                                  bool
	SplineSamples                           int
	Orbit                                   *motion.NestedOrbit
	Weave                                   *motion.Weave
	Circle                                  *motion.CircleFormation
	Harmonic                                *motion.HarmonicFormation `json:"-"`
	Formula                                 *motion.FormulaFormation  `json:"-"`
	FormulaWidth, FormulaHeight             float64
	Coupled                                 *motion.CoupledOrbitFormationConfig
	HarmonicClockStart, HarmonicClockStep   [2]float64
	HarmonicEnvelope                        *motion.BounceBankConfig
	Formation                               func(float64, int) motion.Point `json:"-"`
	Grid                                    *GridFormation
	Translation                             *motion.HarmonicTranslation
	RecurrentTranslation                    *motion.RecurrentTranslationConfig // Shared offset advanced once per Update.
	Origin, Velocity, Spacing               motion.Point
	Speed, Phase, PhaseSpacing, Delay       float64
	PhaseStep                               float64 // Optional cumulative phase per Update.
	Orient                                  bool
	ScaleX, ScaleY, Angle, AnchorX, AnchorY float64
	Opacity                                 float64 // Zero defaults to one; use a signal for a hidden starting value.
	AlphaOnly                               bool    // Multiply only the alpha channel, retaining bright RGB materials.
	Filter                                  ebiten.Filter
	Blend                                   ebiten.Blend
	Reverse                                 bool
	Signals                                 GroupSignals                       `json:"-"`
	PerInstance                             []GroupSignals                     `json:"-"`
	Context                                 func(kit.Frame) modulation.Context `json:"-"`
}

type GroupPose struct {
	X, Y, ScaleX, ScaleY, Angle, Opacity float64
	Frame                                int
}

// Group prepares reusable poses once in Update and reuses them for every Draw.
// Atlas frames are borrowed. This keeps motion and audio sampling independent of
// display refresh and allows the same formation to be drawn in several layers.
type Group struct {
	config           GroupConfig
	poses            []GroupPose
	phase            float64
	harmonicClocks   [2]float64
	harmonicEnvelope float64
	harmonicBounce   *motion.BounceBank
	recurrent        *motion.RecurrentTranslation
	coupled          *motion.CoupledOrbitFormation
	ready            bool
}

func NewGroup(c GroupConfig) (*Group, error) {
	if c.Count < 0 || c.Count > 1_000_000 || len(c.Frames) == 0 {
		return nil, fmt.Errorf("sprites: invalid group count or empty frame bank")
	}
	kinds := 0
	if c.Path != nil {
		kinds++
	}
	if len(c.Points) > 0 {
		kinds++
	}
	if c.Orbit != nil {
		kinds++
	}
	if c.Weave != nil {
		kinds++
	}
	if c.Circle != nil {
		kinds++
	}
	if c.Harmonic != nil {
		kinds++
	}
	if c.Formula != nil {
		kinds++
	}
	if c.Coupled != nil {
		kinds++
	}
	if c.Formation != nil {
		kinds++
	}
	if kinds > 1 {
		return nil, fmt.Errorf("sprites: choose one formation trajectory")
	}
	if c.Harmonic == nil && c.HarmonicEnvelope != nil {
		return nil, fmt.Errorf("sprites: harmonic envelope needs a harmonic formation")
	}
	if c.Harmonic != nil && c.Harmonic.IndexOffsetCount() > 0 && c.Count > c.Harmonic.IndexOffsetCount() {
		return nil, fmt.Errorf("sprites: harmonic index offsets are shorter than group count")
	}
	if c.Formula != nil && (c.Formula.X == nil || c.Formula.Y == nil) {
		return nil, fmt.Errorf("sprites: incomplete formula formation")
	}
	if c.Translation != nil && c.RecurrentTranslation != nil {
		return nil, fmt.Errorf("sprites: choose direct or recurrent shared translation")
	}
	for _, value := range [...]float64{c.HarmonicClockStart[0], c.HarmonicClockStart[1], c.HarmonicClockStep[0], c.HarmonicClockStep[1]} {
		if !finiteField(value) {
			return nil, fmt.Errorf("sprites: nonfinite harmonic clock")
		}
	}
	for _, v := range []float64{c.FPS, c.Origin.X, c.Origin.Y, c.Velocity.X, c.Velocity.Y, c.Spacing.X, c.Spacing.Y, c.Speed, c.Phase, c.PhaseSpacing, c.PhaseStep, c.Delay, c.ScaleX, c.ScaleY, c.Angle, c.AnchorX, c.AnchorY, c.Opacity, c.FormulaWidth, c.FormulaHeight} {
		if !finiteField(v) {
			return nil, fmt.Errorf("sprites: nonfinite group parameter")
		}
	}
	if c.SplineSamples < 0 {
		return nil, fmt.Errorf("sprites: invalid spline precision")
	}
	if len(c.Points) > 0 {
		var err error
		if c.SplineSamples > 0 {
			c.Path, err = motion.NewSpline(c.Points, c.Closed, c.SplineSamples)
		} else {
			c.Path, err = motion.NewPolyline(c.Points, c.Closed)
		}
		if err != nil {
			return nil, err
		}
	}
	if c.ScaleX == 0 {
		c.ScaleX = 1
	}
	if c.ScaleY == 0 {
		c.ScaleY = 1
	}
	if c.Opacity == 0 {
		c.Opacity = 1
	}
	if c.Grid != nil {
		grid := *c.Grid
		if grid.Columns < 1 || grid.Columns > 1_000_000 || grid.Rows < 0 || grid.Rows > 1_000_000 || !finiteField(grid.StepX) || !finiteField(grid.StepY) {
			return nil, fmt.Errorf("sprites: invalid grid formation")
		}
		if grid.Rows == 0 {
			grid.Rows = (c.Count + grid.Columns - 1) / grid.Columns
		}
		c.Grid = &grid
	}
	if c.Translation != nil {
		translation := *c.Translation
		if len(translation.X) > 64 || len(translation.Y) > 64 {
			return nil, fmt.Errorf("sprites: too many translation harmonics")
		}
		for _, term := range append(append([]motion.HarmonicTerm(nil), translation.X...), translation.Y...) {
			if !finiteField(term.Amplitude) || !finiteField(term.Rate) || !finiteField(term.Phase) {
				return nil, fmt.Errorf("sprites: nonfinite translation harmonic")
			}
		}
		translation.X = append([]motion.HarmonicTerm(nil), translation.X...)
		translation.Y = append([]motion.HarmonicTerm(nil), translation.Y...)
		c.Translation = &translation
	}
	c.Frames = append([]*ebiten.Image(nil), c.Frames...)
	c.Points = append([]motion.Point(nil), c.Points...)
	c.PerInstance = append([]GroupSignals(nil), c.PerInstance...)
	if c.Orbit != nil {
		copy := *c.Orbit
		c.Orbit = &copy
	}
	if c.Weave != nil {
		copy := *c.Weave
		c.Weave = &copy
	}
	if c.Circle != nil {
		copy := *c.Circle
		if copy.IndexCount < 0 || copy.IndexCount > 1_000_000 {
			return nil, fmt.Errorf("sprites: invalid circle index count")
		}
		for _, value := range []float64{copy.RadiusX, copy.RadiusY, copy.IndexAngle, copy.XAmplitude, copy.XRate, copy.XIndexPhase, copy.XPhase, copy.YAmplitude, copy.YRate, copy.YIndexPhase, copy.YPhase, copy.ScaleBase, copy.ScaleAmplitude, copy.ScaleRate, copy.ScaleIndexPhase, copy.ScalePhase} {
			if !finiteField(value) {
				return nil, fmt.Errorf("sprites: nonfinite circle parameter")
			}
		}
		if copy.ScaleBase == 0 && copy.ScaleAmplitude == 0 {
			copy.ScaleBase = 1
		}
		c.Circle = &copy
	}
	g := &Group{config: c, poses: make([]GroupPose, c.Count), phase: c.Phase, harmonicClocks: c.HarmonicClockStart, harmonicEnvelope: 1}
	if c.Coupled != nil {
		var err error
		g.coupled, err = motion.NewCoupledOrbitFormation(*c.Coupled)
		if err != nil {
			return nil, err
		}
		if g.coupled.Len() != c.Count {
			return nil, fmt.Errorf("sprites: coupled orbit ranges do not match group count")
		}
		g.config.Coupled = nil // The group owns its phase and range bank.
	}
	if c.RecurrentTranslation != nil {
		var err error
		g.recurrent, err = motion.NewRecurrentTranslation(*c.RecurrentTranslation)
		if err != nil {
			return nil, err
		}
		g.config.RecurrentTranslation = nil // The group owns an independent copy.
	}
	if c.HarmonicEnvelope != nil {
		var err error
		g.harmonicBounce, err = motion.NewBounceBank(*c.HarmonicEnvelope)
		if err != nil {
			return nil, err
		}
		if g.harmonicBounce.Len() != 1 {
			return nil, fmt.Errorf("sprites: harmonic envelope needs one bounce lane")
		}
		g.harmonicEnvelope = g.harmonicBounce.At(0)
	}
	if g.coupled == nil {
		g.sample(kit.Frame{})
	}
	return g, nil
}

func (g *Group) Update(f kit.Frame) error {
	if !finiteField(f.Time) {
		return fmt.Errorf("sprites: nonfinite group time")
	}
	if g.config.Harmonic != nil {
		nextClocks := [2]float64{g.harmonicClocks[0] + g.config.HarmonicClockStep[0], g.harmonicClocks[1] + g.config.HarmonicClockStep[1]}
		if !finiteField(nextClocks[0]) || !finiteField(nextClocks[1]) {
			return fmt.Errorf("sprites: harmonic clock overflows")
		}
		g.harmonicClocks = nextClocks
		if g.harmonicBounce != nil {
			g.harmonicBounce.Step()
			g.harmonicEnvelope = g.harmonicBounce.At(0)
		}
	}
	next := g.phase + g.config.PhaseStep
	if !finiteField(next) {
		return fmt.Errorf("sprites: group phase overflows")
	}
	g.phase = next
	if g.recurrent != nil {
		g.recurrent.Step()
	}
	if g.coupled != nil {
		g.coupled.Step()
	}
	return g.sample(f)
}

func (g *Group) sample(f kit.Frame) error {
	c := g.config
	translation := motion.Point{}
	if c.Translation != nil {
		translation = c.Translation.At(g.phase + f.Time*c.Speed)
	} else if g.recurrent != nil {
		translation = g.recurrent.At()
	}
	context := modulation.Context{Seconds: f.Time}
	if c.Context != nil {
		context = c.Context(f)
	}
	common := GroupPose{ScaleX: 1, ScaleY: 1, Opacity: 1}
	applyGroupSignals(&common, c.Signals, context)
	var coupledPoses []motion.Point
	if g.coupled != nil {
		coupledPoses = g.coupled.Poses()
	}
	for i := range g.poses {
		t := f.Time - float64(i)*c.Delay
		phase := g.phase + t*c.Speed + float64(i)*c.PhaseSpacing
		position, tangent := motion.Point{}, motion.Point{X: 1}
		circleScale := 1.0
		switch {
		case g.coupled != nil:
			position = coupledPoses[i]
		case c.Path != nil:
			position, tangent = c.Path.At(phase)
		case c.Orbit != nil:
			position = c.Orbit.At(phase)
			if c.Orient {
				a, b := c.Orbit.At(phase-1e-4), c.Orbit.At(phase+1e-4)
				tangent = motion.Point{X: b.X - a.X, Y: b.Y - a.Y}
			}
		case c.Weave != nil:
			position = c.Weave.At(phase, i)
			if c.Orient {
				a, b := c.Weave.At(phase-1e-4, i), c.Weave.At(phase+1e-4, i)
				tangent = motion.Point{X: b.X - a.X, Y: b.Y - a.Y}
			}
		case c.Circle != nil:
			position, circleScale = c.Circle.At(g.phase+t*c.Speed, i)
		case c.Harmonic != nil:
			position = c.Harmonic.At(i, g.harmonicClocks, g.harmonicEnvelope)
		case c.Formula != nil:
			position = c.Formula.At(phase, i, c.FormulaWidth, c.FormulaHeight, c.Count)
			if c.Orient {
				a := c.Formula.At(phase-1e-4, i, c.FormulaWidth, c.FormulaHeight, c.Count)
				b := c.Formula.At(phase+1e-4, i, c.FormulaWidth, c.FormulaHeight, c.Count)
				tangent = motion.Point{X: b.X - a.X, Y: b.Y - a.Y}
			}
		case c.Formation != nil:
			position = c.Formation(phase, i)
			if c.Orient {
				a, b := c.Formation(phase-1e-4, i), c.Formation(phase+1e-4, i)
				tangent = motion.Point{X: b.X - a.X, Y: b.Y - a.Y}
			}
		default:
			position = motion.Point{X: c.Velocity.X * t, Y: c.Velocity.Y * t}
			tangent = c.Velocity
		}
		p := GroupPose{X: c.Origin.X + float64(i)*c.Spacing.X + position.X, Y: c.Origin.Y + float64(i)*c.Spacing.Y + position.Y, ScaleX: c.ScaleX, ScaleY: c.ScaleY, Angle: c.Angle, Opacity: c.Opacity}
		if !finiteField(p.X) || !finiteField(p.Y) {
			return fmt.Errorf("sprites: nonfinite formation pose")
		}
		if c.Grid != nil {
			column := i % c.Grid.Columns
			row := i / c.Grid.Columns
			p.X += (float64(column) - float64(c.Grid.Columns-1)/2) * c.Grid.StepX
			p.Y += (float64(row) - float64(c.Grid.Rows-1)/2) * c.Grid.StepY
		}
		p.X += translation.X
		p.Y += translation.Y
		p.ScaleX *= circleScale
		p.ScaleY *= circleScale
		if c.Orient && (tangent.X != 0 || tangent.Y != 0) {
			p.Angle += math.Atan2(tangent.Y, tangent.X)
		}
		p.X += common.X
		p.Y += common.Y
		p.Angle += common.Angle
		p.ScaleX *= common.ScaleX
		p.ScaleY *= common.ScaleY
		p.Opacity *= common.Opacity
		if i < len(c.PerInstance) {
			applyGroupSignals(&p, c.PerInstance[i], context)
		}
		frame := math.Floor(t*c.FPS) + float64(c.FrameOffset) + float64(i)*float64(c.FrameStride)
		if !finiteField(frame) {
			frame = 0
		}
		frame = math.Mod(frame, float64(len(c.Frames)))
		if frame < 0 {
			frame += float64(len(c.Frames))
		}
		p.Frame = int(frame)
		g.poses[i] = p
	}
	g.ready = true
	return nil
}

func applyGroupSignals(p *GroupPose, s GroupSignals, c modulation.Context) {
	if s.X != nil {
		p.X += s.X.At(c)
	}
	if s.Y != nil {
		p.Y += s.Y.At(c)
	}
	if s.ScaleX != nil {
		p.ScaleX *= s.ScaleX.At(c)
	}
	if s.ScaleY != nil {
		p.ScaleY *= s.ScaleY.At(c)
	}
	if s.Angle != nil {
		p.Angle += s.Angle.At(c)
	}
	if s.Opacity != nil {
		p.Opacity *= s.Opacity.At(c)
	}
}

// Poses returns borrowed prepared transforms for inspection or another renderer.
func (g *Group) Poses() []GroupPose { return g.poses }
func (g *Group) Draw(dst *ebiten.Image) {
	if !g.ready {
		return
	}
	c := g.config
	for n := range g.poses {
		i := n
		if c.Reverse {
			i = len(g.poses) - 1 - n
		}
		p := g.poses[i]
		img := c.Frames[p.Frame]
		if img == nil || p.Opacity <= 0 || !finiteField(p.Opacity) || !finiteField(p.X) || !finiteField(p.Y) || !finiteField(p.ScaleX) || !finiteField(p.ScaleY) || !finiteField(p.Angle) {
			continue
		}
		op := ebiten.DrawImageOptions{Filter: c.Filter, Blend: c.Blend}
		op.GeoM.Translate(-float64(img.Bounds().Dx())*c.AnchorX, -float64(img.Bounds().Dy())*c.AnchorY)
		op.GeoM.Scale(p.ScaleX, p.ScaleY)
		op.GeoM.Rotate(p.Angle)
		op.GeoM.Translate(p.X, p.Y)
		if c.AlphaOnly {
			op.ColorScale.Scale(1, 1, 1, float32(math.Min(1, p.Opacity)))
		} else {
			op.ColorScale.ScaleAlpha(float32(math.Min(1, p.Opacity)))
		}
		dst.DrawImage(img, &op)
	}
}

func (g *Group) SetCount(count int) error {
	if count < 0 || count > 1_000_000 {
		return fmt.Errorf("sprites: invalid group count")
	}
	if g.config.Harmonic != nil && g.config.Harmonic.IndexOffsetCount() > 0 && count > g.config.Harmonic.IndexOffsetCount() {
		return fmt.Errorf("sprites: harmonic index offsets are shorter than group count")
	}
	if g.coupled != nil && count != g.coupled.Len() {
		return fmt.Errorf("sprites: coupled orbit count needs matching ranges")
	}
	if count > cap(g.poses) {
		g.poses = make([]GroupPose, count)
	} else {
		g.poses = g.poses[:count]
	}
	return nil
}

// SetPhase accepts an authored phase accumulator without converting it through
// seconds. The new phase is sampled on the next Update, alongside all bindings.
func (g *Group) SetPhase(phase float64) error {
	if g.recurrent != nil || g.coupled != nil {
		return fmt.Errorf("sprites: use Update for a stateful formation")
	}
	if !finiteField(phase) {
		return fmt.Errorf("sprites: nonfinite group phase")
	}
	g.config.Phase = phase
	g.phase = phase
	return nil
}

// Phase returns the group's cumulative authored phase.
func (g *Group) Phase() float64 { return g.phase }

// SetHarmonicState samples a configured harmonic formation with two caller-owned
// phase clocks and an amplitude envelope. It prepares poses immediately, so
// callers can draw the group more than once without advancing it again.
func (g *Group) SetHarmonicState(clocks [2]float64, envelope float64) error {
	if g.config.Harmonic == nil || !finiteField(clocks[0]) || !finiteField(clocks[1]) || !finiteField(envelope) {
		return fmt.Errorf("sprites: invalid harmonic state")
	}
	g.harmonicClocks = clocks
	g.harmonicEnvelope = envelope
	return g.sample(kit.Frame{})
}

// ResetHarmonics restores configured phase clocks and the optional bouncing
// envelope, then prepares the initial poses without advancing the group.
func (g *Group) ResetHarmonics() error {
	if g.config.Harmonic == nil {
		return fmt.Errorf("sprites: group has no harmonic formation")
	}
	g.harmonicClocks = g.config.HarmonicClockStart
	g.harmonicEnvelope = 1
	if g.harmonicBounce != nil {
		g.harmonicBounce.Reset()
		g.harmonicEnvelope = g.harmonicBounce.At(0)
	}
	return g.sample(kit.Frame{})
}

// Advance adds a caller-chosen phase increment and samples new poses once.
// It is useful for variable-speed user controls without exposing a local
// per-sprite controller. Do not also call Update for the same simulation tick.
func (g *Group) Advance(delta float64) error {
	if g.recurrent != nil || g.coupled != nil {
		return fmt.Errorf("sprites: use Update for a stateful formation")
	}
	if !finiteField(delta) || !finiteField(g.phase+delta) {
		return fmt.Errorf("sprites: invalid group phase increment")
	}
	g.phase += delta
	return g.sample(kit.Frame{})
}

// RecurrentTranslationController exposes the shared cached offset for cues or
// inspection. Group.Update owns its Step call; do not step it separately.
func (g *Group) RecurrentTranslationController() *motion.RecurrentTranslation {
	if g == nil {
		return nil
	}
	return g.recurrent
}

// ResetRecurrentTranslation restores the authored clocks and prepared poses.
func (g *Group) ResetRecurrentTranslation() error {
	if g == nil || g.recurrent == nil {
		return fmt.Errorf("sprites: group has no recurrent translation")
	}
	g.recurrent.Reset()
	return g.sample(kit.Frame{})
}

// CoupledOrbitController exposes the owned sequential orbit for input or music
// cues. Change parameters before Group.Update; Draw never advances its phase.
func (g *Group) CoupledOrbitController() *motion.CoupledOrbitFormation {
	if g == nil {
		return nil
	}
	return g.coupled
}

// ResetCoupledOrbit restores authored parameters and waits for the next Update.
func (g *Group) ResetCoupledOrbit() error {
	if g == nil || g.coupled == nil {
		return fmt.Errorf("sprites: group has no coupled orbit")
	}
	g.coupled.Reset()
	clear(g.poses)
	g.ready = false
	return nil
}
