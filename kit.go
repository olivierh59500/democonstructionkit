// Package democonstructionkit composes reusable Ebitengine demo effects.
package democonstructionkit

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/timeline"
)

type Frame = timeline.Frame

// Effect updates from an absolute local time and draws without advancing state.
// Instances own their working buffers; assets passed to them remain caller-owned.
type Effect interface {
	Update(Frame) error
	Draw(*ebiten.Image)
}

// Func makes production-specific choreography fit the same composition interface.
type Func struct {
	OnUpdate func(Frame) error
	OnDraw   func(*ebiten.Image)
}

func (f Func) Update(t Frame) error {
	if f.OnUpdate != nil {
		return f.OnUpdate(t)
	}
	return nil
}
func (f Func) Draw(dst *ebiten.Image) {
	if f.OnDraw != nil {
		f.OnDraw(dst)
	}
}

// Group layers effects in slice order. Do not share a mutable instance between groups.
type Group []Effect

func (g Group) Update(f Frame) error {
	for i, e := range g {
		if e == nil {
			return fmt.Errorf("kit: nil effect at %d", i)
		}
		if err := e.Update(f); err != nil {
			return fmt.Errorf("effect %d: %w", i, err)
		}
	}
	return nil
}
func (g Group) Draw(dst *ebiten.Image) {
	for _, e := range g {
		e.Draw(dst)
	}
}
func (g Group) Close() error {
	var err error
	for _, e := range g {
		err = errors.Join(err, Close(e))
	}
	return err
}

// Close releases an effect's owned resources when it implements Close.
func Close(e Effect) error {
	if c, ok := e.(interface{ Close() error }); ok {
		return c.Close()
	}
	return nil
}

// Viewport renders an effect at its own resolution, clips it, then fits it in Rect.
// This is the building block for multiscreen demos and transformed effect layers.
type Viewport struct {
	Effect     Effect
	Rect       image.Rectangle
	Opacity    float32
	Background color.Color
	Filter     ebiten.Filter
	canvas     *ebiten.Image
}

func NewViewport(effect Effect, width, height int, rect image.Rectangle) (*Viewport, error) {
	if effect == nil || width <= 0 || height <= 0 || rect.Empty() {
		return nil, fmt.Errorf("kit: invalid viewport")
	}
	return &Viewport{Effect: effect, Rect: rect, Opacity: 1, canvas: ebiten.NewImage(width, height)}, nil
}
func (v *Viewport) Update(f Frame) error { return v.Effect.Update(f) }
func (v *Viewport) Draw(dst *ebiten.Image) {
	v.canvas.Clear()
	if v.Background != nil {
		v.canvas.Fill(v.Background)
	}
	v.Effect.Draw(v.canvas)
	b := v.canvas.Bounds()
	scale := math.Min(float64(v.Rect.Dx())/float64(b.Dx()), float64(v.Rect.Dy())/float64(b.Dy()))
	op := ebiten.DrawImageOptions{Filter: v.Filter}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(v.Rect.Min.X)+(float64(v.Rect.Dx())-float64(b.Dx())*scale)/2, float64(v.Rect.Min.Y)+(float64(v.Rect.Dy())-float64(b.Dy())*scale)/2)
	op.ColorScale.ScaleAlpha(max(0, min(1, v.Opacity)))
	dst.DrawImage(v.canvas, &op)
}
func (v *Viewport) Close() error { v.canvas.Deallocate(); return Close(v.Effect) }

// Sequence selects scenes with local time. Its children must sample absolute time
// if they need to restart seamlessly on looping or seeking.
type Sequence struct {
	scenes []Effect
	timing *timeline.Sequence
	active int
}

func NewSequence(scenes []Effect, durations []time.Duration, loop bool) (*Sequence, error) {
	if len(scenes) != len(durations) {
		return nil, fmt.Errorf("kit: scene/duration count mismatch")
	}
	for _, s := range scenes {
		if s == nil {
			return nil, fmt.Errorf("kit: nil scene")
		}
	}
	t, err := timeline.NewSequence(durations, loop)
	if err != nil {
		return nil, err
	}
	return &Sequence{scenes: append([]Effect(nil), scenes...), timing: t, active: -1}, nil
}
func (s *Sequence) Update(f Frame) error {
	i, local, ok := s.timing.At(time.Duration(f.Time * float64(time.Second)))
	if !ok {
		s.active = -1
		return nil
	}
	s.active = i
	f.Time = local.Seconds()
	return s.scenes[i].Update(f)
}
func (s *Sequence) Draw(dst *ebiten.Image) {
	if s.active >= 0 {
		s.scenes[s.active].Draw(dst)
	}
}
func (s *Sequence) Close() error { return Group(s.scenes).Close() }

// Game adapts an effect tree to Ebitengine. Set Ebitengine TPS to Config.TPS in main.
type Game struct {
	root    Effect
	config  Config
	Clock   *timeline.Clock
	current Frame
}
type Config struct {
	Width, Height int
	TPS           float64
	Background    color.Color
}

func NewGame(effect Effect, c Config) (*Game, error) {
	if effect == nil || c.Width <= 0 || c.Height <= 0 {
		return nil, fmt.Errorf("kit: invalid game config")
	}
	if c.TPS == 0 {
		c.TPS = 60
	}
	clock, err := timeline.NewClock(c.TPS)
	if err != nil {
		return nil, err
	}
	return &Game{root: effect, config: c, Clock: clock}, nil
}
func (g *Game) Update() error { g.current = g.Clock.Step(); return g.root.Update(g.current) }
func (g *Game) Draw(screen *ebiten.Image) {
	if g.config.Background != nil {
		screen.Fill(g.config.Background)
	}
	g.root.Draw(screen)
}
func (g *Game) Layout(int, int) (int, int) { return g.config.Width, g.config.Height }
func (g *Game) Frame() Frame               { return g.current }
func (g *Game) Close() error               { return Close(g.root) }
