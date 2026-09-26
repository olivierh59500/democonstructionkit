package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// DelayedTrailConfig maps ordered sprite images to independent history delays.
// Capacity zero uses max(Delays)+1; Reverse changes drawing order, not which
// image receives each delay. Assets remain borrowed by the effect.
type DelayedTrailConfig struct {
	Images                           []*ebiten.Image
	Delays                           []int
	Capacity                         int
	Initial                          geometry.Vec2
	ScaleX, ScaleY, AnchorX, AnchorY float64
	Opacity                          float64
	Reverse                          bool
	Filter                           ebiten.Filter
	Blend                            ebiten.Blend
}

// DelayedTrail advances one point history and prepares a cached sprite group.
// Draw never pushes history, so the same trail may be layered several times.
type DelayedTrail struct {
	group    *Group
	history  *motion.PointHistory[geometry.Vec2]
	delays   []int
	initial  geometry.Vec2
	position geometry.Vec2
}

func NewDelayedTrail(c DelayedTrailConfig) (*DelayedTrail, error) {
	if len(c.Images) == 0 || len(c.Images) > 1_000_000 || len(c.Delays) != len(c.Images) ||
		math.IsNaN(c.Initial.X) || math.IsInf(c.Initial.X, 0) || math.IsNaN(c.Initial.Y) || math.IsInf(c.Initial.Y, 0) {
		return nil, fmt.Errorf("sprites: invalid delayed trail images or position")
	}
	maximum := 0
	for _, delay := range c.Delays {
		if delay < 0 || delay > 1_000_000 {
			return nil, fmt.Errorf("sprites: invalid delayed trail delay")
		}
		maximum = max(maximum, delay)
	}
	if c.Capacity == 0 {
		c.Capacity = maximum + 1
	}
	if c.Capacity < maximum+1 || c.Capacity > 1_000_001 {
		return nil, fmt.Errorf("sprites: delayed trail capacity cannot hold the delays")
	}
	history, err := motion.NewPointHistory(c.Capacity, c.Initial)
	if err != nil {
		return nil, err
	}
	trail := &DelayedTrail{history: history, delays: append([]int(nil), c.Delays...),
		initial: c.Initial, position: c.Initial}
	group, err := NewGroup(GroupConfig{
		Frames: c.Images, Count: len(c.Images), FrameStride: 1,
		ScaleX: c.ScaleX, ScaleY: c.ScaleY, AnchorX: c.AnchorX, AnchorY: c.AnchorY,
		Opacity: c.Opacity, Reverse: c.Reverse, Filter: c.Filter, Blend: c.Blend,
		Formation: func(_ float64, index int) motion.Point {
			position, _ := trail.history.At(trail.delays[index])
			return motion.Point{X: position.X, Y: position.Y}
		},
	})
	if err != nil {
		return nil, err
	}
	trail.group = group
	return trail, nil
}

// SetPosition supplies the next input or path sample without moving history.
func (t *DelayedTrail) SetPosition(point geometry.Vec2) error {
	if t == nil || math.IsNaN(point.X) || math.IsInf(point.X, 0) || math.IsNaN(point.Y) || math.IsInf(point.Y, 0) {
		return fmt.Errorf("sprites: invalid delayed trail point")
	}
	t.position = point
	return nil
}

func (t *DelayedTrail) Position() geometry.Vec2 { return t.position }

func (t *DelayedTrail) Update(frame kit.Frame) error {
	if t == nil {
		return fmt.Errorf("sprites: nil delayed trail")
	}
	t.history.Push(t.position)
	return t.group.Update(frame)
}

func (t *DelayedTrail) Draw(dst *ebiten.Image) {
	if t != nil {
		t.group.Draw(dst)
	}
}

func (t *DelayedTrail) Poses() []GroupPose { return t.group.Poses() }

// Reset restores the initial point to every delay without allocating.
func (t *DelayedTrail) Reset() error {
	if t == nil {
		return fmt.Errorf("sprites: nil delayed trail")
	}
	t.position = t.initial
	t.history.Reset(t.initial)
	return t.group.Advance(0)
}
