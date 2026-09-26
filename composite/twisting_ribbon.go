package composite

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// TwistingRibbonConfig borrows two live source images and preserves the back-
// then-front draw order for each strip. FirstUpdateHolds keeps a source screen
// whose first visible frame draws the starting pose before any phase advance.
type TwistingRibbonConfig struct {
	Front, Back      *ebiten.Image
	Motion           motion.TwistingRibbonConfig
	SourceHeight     float64
	Filter           ebiten.Filter
	Blend            ebiten.Blend
	FirstUpdateHolds bool
}

// TwistingRibbon renders a two-sided strip deformation without intermediate
// GPU surfaces or per-frame point-slice allocation. The artwork stays borrowed.
type TwistingRibbon struct {
	config  TwistingRibbonConfig
	motion  *motion.TwistingRibbon
	updated bool
}

func NewTwistingRibbon(c TwistingRibbonConfig) (*TwistingRibbon, error) {
	if c.Front == nil || c.Back == nil || c.SourceHeight <= 0 || math.IsNaN(c.SourceHeight) || math.IsInf(c.SourceHeight, 0) ||
		c.Front.Bounds().Dx() < c.Motion.Width || c.Back.Bounds().Dx() < c.Motion.Width ||
		float64(c.Front.Bounds().Dy()) < c.SourceHeight || float64(c.Back.Bounds().Dy()) < c.SourceHeight {
		return nil, fmt.Errorf("composite: invalid twisting ribbon sources")
	}
	controller, err := motion.NewTwistingRibbon(c.Motion)
	if err != nil {
		return nil, err
	}
	return &TwistingRibbon{config: c, motion: controller}, nil
}

func (r *TwistingRibbon) Update(kit.Frame) error {
	if r == nil {
		return fmt.Errorf("composite: nil twisting ribbon")
	}
	if r.config.FirstUpdateHolds && !r.updated {
		r.updated = true
		return nil
	}
	r.Advance()
	return nil
}

// Advance selects the next prepared pose when a caller draws before stepping.
func (r *TwistingRibbon) Advance() {
	if r == nil {
		return
	}
	r.updated = true
	r.motion.Advance()
}

func (r *TwistingRibbon) Draw(dst *ebiten.Image) { r.DrawAt(dst, 0, 0) }

func (r *TwistingRibbon) DrawAt(dst *ebiten.Image, x, y float64) {
	if r == nil || dst == nil {
		return
	}
	c := r.config
	for _, pose := range r.motion.Poses() {
		region := Region{X: float64(pose.SourceX), Width: float64(c.Motion.StripWidth), Height: c.SourceHeight}
		if pose.BackVisible {
			var options ebiten.DrawImageOptions
			options.Filter, options.Blend = c.Filter, c.Blend
			options.GeoM.Scale(1, pose.BackScaleY)
			options.GeoM.Translate(x+float64(pose.SourceX), y+pose.BackY)
			DrawRegion(dst, c.Back, region, &options)
		}
		if pose.FrontVisible {
			var options ebiten.DrawImageOptions
			options.Filter, options.Blend = c.Filter, c.Blend
			options.GeoM.Scale(1, pose.FrontScaleY)
			options.GeoM.Translate(x+float64(pose.SourceX), y+pose.FrontY)
			DrawRegion(dst, c.Front, region, &options)
		}
	}
}

func (r *TwistingRibbon) Phase() int                          { return r.motion.Phase() }
func (r *TwistingRibbon) Poses() []motion.TwistingRibbonSlice { return r.motion.Poses() }

var _ kit.Effect = (*TwistingRibbon)(nil)
