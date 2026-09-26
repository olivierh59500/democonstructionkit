package scrolling

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// RibbonConfig composes an atlas-backed horizontal or vertical scroll with a
// fixed-tick transport. CullAdvance is independent of actual font metrics;
// historical scrollers sometimes used a different width for the visible-window
// estimate. Scale and baseline are destination-space drawing parameters.
type RibbonConfig struct {
	Text                  string
	Font                  *Atlas
	Vertical, SkipMissing bool
	Clock                 motion.RibbonClockConfig
	ScaleX, ScaleY        float64
	BaselineY             float64
	CullAdvance           float64
}

// Ribbon owns a deterministic clock and cached generic Scrolling renderer.
// It borrows the atlas and draws into a caller-owned surface without clearing.
type Ribbon struct {
	config     RibbonConfig
	clock      *motion.RibbonClock
	renderer   *Scrolling
	visibleMap Mapper
	drawExtent float64
}

func NewRibbon(c RibbonConfig) (*Ribbon, error) {
	if c.Text == "" || c.Font == nil || c.Font.Metrics() == nil ||
		math.IsNaN(c.ScaleX) || math.IsInf(c.ScaleX, 0) ||
		math.IsNaN(c.ScaleY) || math.IsInf(c.ScaleY, 0) ||
		math.IsNaN(c.BaselineY) || math.IsInf(c.BaselineY, 0) ||
		math.IsNaN(c.CullAdvance) || math.IsInf(c.CullAdvance, 0) || c.CullAdvance < 0 {
		return nil, fmt.Errorf("scrolling: invalid ribbon text, font or scale")
	}
	if c.ScaleX == 0 {
		c.ScaleX = 1
	}
	if c.ScaleY == 0 {
		c.ScaleY = 1
	}
	if c.CullAdvance == 0 {
		c.CullAdvance = c.Font.Metrics().LineHeight()
	}
	if c.ScaleX <= 0 || c.ScaleY <= 0 || c.CullAdvance <= 0 {
		return nil, fmt.Errorf("scrolling: nonpositive ribbon scale or cull width")
	}
	glyphs := c.Font.Layout(c.Text, AtlasText{Vertical: c.Vertical, SkipMissing: c.SkipMissing})
	renderer, err := New(Config{Glyphs: glyphs, Vertical: c.Vertical})
	if err != nil {
		return nil, err
	}
	if c.Clock.Length == 0 {
		c.Clock.Length = renderer.Length()
	}
	clock, err := motion.NewRibbonClock(c.Clock)
	if err != nil {
		_ = renderer.Close()
		return nil, err
	}
	ribbon := &Ribbon{config: c, clock: clock, renderer: renderer}
	ribbon.visibleMap = func(g Sample, _ *ebiten.DrawImageOptions) bool {
		if ribbon.config.Vertical {
			return g.Y+g.Glyph.Advance*ribbon.config.ScaleY > 0 && g.Y < ribbon.drawExtent
		}
		return g.X+g.Glyph.Advance*ribbon.config.ScaleX > 0 && g.X < ribbon.drawExtent
	}
	return ribbon, nil
}

func (r *Ribbon) Update(kit.Frame) error { return r.clock.Step() }

func (r *Ribbon) Draw(dst *ebiten.Image) {
	if r == nil || dst == nil {
		return
	}
	c := r.config
	state := IdentityState()
	state.ScaleX, state.ScaleY = c.ScaleX, c.ScaleY
	if c.Vertical {
		r.drawExtent = float64(dst.Bounds().Dy())
		state.Y = r.drawExtent - r.clock.Offset()*c.ScaleY
	} else {
		r.drawExtent = float64(dst.Bounds().Dx())
		state.X = r.clock.Offset() * c.ScaleX
		state.Y = c.BaselineY * c.ScaleY
		if r.clock.Offset() < 0 {
			state.First = int(math.Floor(-r.clock.Offset() / (c.CullAdvance * c.ScaleX)))
		}
	}
	state.Map = r.visibleMap
	r.renderer.DrawAt(dst, state)
}

func (r *Ribbon) Offset() float64                    { return r.clock.Offset() }
func (r *Ribbon) Length() float64                    { return r.renderer.Length() }
func (r *Ribbon) Clock() *motion.RibbonClock         { return r.clock }
func (r *Ribbon) SetSpeedMultiplier(v float64) error { return r.clock.SetMultiplier(v) }

func (r *Ribbon) Close() error {
	if r == nil || r.renderer == nil {
		return nil
	}
	err := r.renderer.Close()
	r.renderer = nil
	return err
}
