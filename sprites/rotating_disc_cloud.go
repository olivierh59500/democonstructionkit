package sprites

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

// RotatingDiscCloudConfig combines independent point projection with one
// batched disc material. Tint nil uses white. Antialias is explicit so screens
// can choose their historical edge coverage and mobile submission cost.
type RotatingDiscCloudConfig struct {
	Motion    geometry.RotatingDiscCloudConfig
	Tint      color.Color
	Antialias bool
}

// RotatingDiscCloud owns its projected pose banks and one Discs renderer.
// Point data is copied; the drawing shader and buffers are closed with it.
type RotatingDiscCloud struct {
	motion   *geometry.RotatingDiscCloud
	renderer *Discs
	discs    []Disc
	tint     ebiten.ColorScale
}

func NewRotatingDiscCloud(c RotatingDiscCloudConfig) (*RotatingDiscCloud, error) {
	motion, err := geometry.NewRotatingDiscCloud(c.Motion)
	if err != nil {
		return nil, err
	}
	renderer, err := NewDiscs(len(c.Motion.Points))
	if err != nil {
		return nil, err
	}
	renderer.Antialias = c.Antialias
	cloud := &RotatingDiscCloud{motion: motion, renderer: renderer, discs: make([]Disc, len(c.Motion.Points))}
	if c.Tint != nil {
		cloud.tint.ScaleWithColor(c.Tint)
	}
	return cloud, nil
}

func (c *RotatingDiscCloud) Update(kit.Frame) error {
	if c == nil {
		return fmt.Errorf("sprites: nil rotating disc cloud")
	}
	if err := c.motion.Step(); err != nil {
		return err
	}
	for i, pose := range c.motion.Poses() {
		c.discs[i] = Disc{X: pose.X, Y: pose.Y, Radius: pose.Radius, ColorScale: c.tint}
	}
	return nil
}

func (c *RotatingDiscCloud) Draw(dst *ebiten.Image) { c.DrawAt(dst, 0, 0) }
func (c *RotatingDiscCloud) DrawAt(dst *ebiten.Image, x, y float64) {
	if c == nil || dst == nil || !c.motion.Ready() {
		return
	}
	c.renderer.DrawAt(dst, c.discs, x, y)
}

// Motion exposes the owned projection clock for cue-driven rate changes.
func (c *RotatingDiscCloud) Motion() *geometry.RotatingDiscCloud { return c.motion }

func (c *RotatingDiscCloud) Close() error {
	if c == nil || c.renderer == nil {
		return nil
	}
	err := c.renderer.Close()
	c.renderer = nil
	return err
}

var _ kit.Effect = (*RotatingDiscCloud)(nil)
