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

// GroupConfig describes a sprite/logo formation, independent of image metrics.
// Choose at most one Path, Points, Orbit or Weave. With none, positions follow
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
	Origin, Velocity, Spacing               motion.Point
	Speed, Phase, PhaseSpacing, Delay       float64
	Orient                                  bool
	ScaleX, ScaleY, Angle, AnchorX, AnchorY float64
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
	config GroupConfig
	poses  []GroupPose
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
	if kinds > 1 {
		return nil, fmt.Errorf("sprites: choose one formation trajectory")
	}
	for _, v := range []float64{c.FPS, c.Origin.X, c.Origin.Y, c.Velocity.X, c.Velocity.Y, c.Spacing.X, c.Spacing.Y, c.Speed, c.Phase, c.PhaseSpacing, c.Delay, c.ScaleX, c.ScaleY, c.Angle, c.AnchorX, c.AnchorY} {
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
	g := &Group{config: c, poses: make([]GroupPose, c.Count)}
	g.Update(kit.Frame{})
	return g, nil
}

func (g *Group) Update(f kit.Frame) error {
	if !finiteField(f.Time) {
		return fmt.Errorf("sprites: nonfinite group time")
	}
	c := g.config
	context := modulation.Context{Seconds: f.Time}
	if c.Context != nil {
		context = c.Context(f)
	}
	common := GroupPose{ScaleX: 1, ScaleY: 1, Opacity: 1}
	applyGroupSignals(&common, c.Signals, context)
	for i := range g.poses {
		t := f.Time - float64(i)*c.Delay
		phase := c.Phase + t*c.Speed + float64(i)*c.PhaseSpacing
		position, tangent := motion.Point{}, motion.Point{X: 1}
		switch {
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
		default:
			position = motion.Point{X: c.Velocity.X * t, Y: c.Velocity.Y * t}
			tangent = c.Velocity
		}
		p := GroupPose{X: c.Origin.X + float64(i)*c.Spacing.X + position.X, Y: c.Origin.Y + float64(i)*c.Spacing.Y + position.Y, ScaleX: c.ScaleX, ScaleY: c.ScaleY, Angle: c.Angle, Opacity: 1}
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
		op.ColorScale.ScaleAlpha(float32(math.Min(1, p.Opacity)))
		dst.DrawImage(img, &op)
	}
}

func (g *Group) SetCount(count int) error {
	if count < 0 || count > 1_000_000 {
		return fmt.Errorf("sprites: invalid group count")
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
	if !finiteField(phase) {
		return fmt.Errorf("sprites: nonfinite group phase")
	}
	g.config.Phase = phase
	return nil
}
