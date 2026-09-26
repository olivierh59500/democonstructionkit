package scrolling

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// DualProfiledRingConfig paints a back ring into an alpha-masked surface,
// composites it, then paints a front ring at the same sampled strip positions.
// Both rings can use different fonts, messages, speed and control syntax.
type DualProfiledRingConfig struct {
	Front, Back                   RingConfig
	Profile                       motion.SegmentedProfileConfig
	Raster                        composite.RasterOverlayConfig
	Width, RingHeight, MaskHeight int
	StripCount, StripWidth        int
	SourceHeight                  float64
	BaseY                         float64
	BackFirstFilter, BackFilter   ebiten.Filter
	FrontFilter, CompositeFilter  ebiten.Filter
	Unmanaged                     bool
}

// DualProfiledRing owns three bounded surfaces and two text transports. Every
// Update prepares the same pose for any number of Draw calls. Borrowed font
// atlases and raster art remain caller-owned.
type DualProfiledRing struct {
	config      DualProfiledRingConfig
	front, back *Ring
	profile     *motion.SegmentedProfile
	raster      *composite.RasterOverlay
	frontImage  *ebiten.Image
	backImage   *ebiten.Image
	maskImage   *ebiten.Image
	backFilter  ebiten.Filter
	firstUpdate bool
}

func NewDualProfiledRing(c DualProfiledRingConfig) (*DualProfiledRing, error) {
	if c.Width < 1 || c.RingHeight < 1 || c.MaskHeight < 1 ||
		c.StripCount < 1 || c.StripWidth < 1 || c.SourceHeight <= 0 ||
		c.StripCount*c.StripWidth > c.Width || c.Profile.VisibleSamples < c.StripCount {
		return nil, fmt.Errorf("scrolling: invalid dual profiled ring dimensions")
	}
	front, err := NewRing(c.Front)
	if err != nil {
		return nil, err
	}
	back, err := NewRing(c.Back)
	if err != nil {
		return nil, err
	}
	profile, err := motion.NewSegmentedProfile(c.Profile)
	if err != nil {
		return nil, err
	}
	raster, err := composite.NewRasterOverlay(c.Raster)
	if err != nil {
		return nil, err
	}
	newSurface := func(height int) *ebiten.Image {
		return ebiten.NewImageWithOptions(image.Rect(0, 0, c.Width, height),
			&ebiten.NewImageOptions{Unmanaged: c.Unmanaged})
	}
	return &DualProfiledRing{config: c, front: front, back: back,
		profile: profile, raster: raster, firstUpdate: true,
		frontImage: newSurface(c.RingHeight), backImage: newSurface(c.RingHeight),
		maskImage: newSurface(c.MaskHeight)}, nil
}

func (r *DualProfiledRing) Update(kit.Frame) error {
	r.raster.Step()
	r.profile.Step()
	r.back.Step()
	r.front.Step()
	r.backFilter = r.config.BackFilter
	if r.firstUpdate {
		r.backFilter = r.config.BackFirstFilter
		r.firstUpdate = false
	}
	return nil
}

func (r *DualProfiledRing) Draw(dst *ebiten.Image) {
	if r == nil || dst == nil {
		return
	}
	r.frontImage.Clear()
	r.backImage.Clear()
	r.maskImage.Clear()
	r.back.DrawAt(r.backImage, 0, 0)
	for i := 0; i < r.config.StripCount; i++ {
		r.drawStrip(r.maskImage, r.backImage, i, r.backFilter)
	}
	r.raster.Draw(r.maskImage)
	var options ebiten.DrawImageOptions
	options.Filter, options.Blend = r.config.CompositeFilter, ebiten.BlendSourceOver
	options.GeoM.Translate(0, 0)
	options.GeoM.Scale(1, 1)
	options.GeoM.Rotate(0)
	options.GeoM.Translate(0, 0)
	options.ColorScale.ScaleAlpha(1)
	dst.DrawImage(r.maskImage, &options)
	r.front.DrawAt(r.frontImage, 0, 0)
	for i := 0; i < r.config.StripCount; i++ {
		r.drawStrip(dst, r.frontImage, i, r.config.FrontFilter)
	}
}

func (r *DualProfiledRing) drawStrip(dst, source *ebiten.Image, index int, filter ebiten.Filter) {
	x := float64(index * r.config.StripWidth)
	region := composite.Region{X: x, Width: float64(r.config.StripWidth), Height: r.config.SourceHeight}
	var options ebiten.DrawImageOptions
	options.Filter = filter
	options.GeoM.Scale(1, 1)
	options.GeoM.Translate(x, r.profile.At(index)+r.config.BaseY)
	composite.DrawRegion(dst, source, region, &options)
}

func (r *DualProfiledRing) Profile() *motion.SegmentedProfile { return r.profile }
func (r *DualProfiledRing) Raster() *composite.RasterOverlay  { return r.raster }
func (r *DualProfiledRing) FrontRing() *Ring                  { return r.front }
func (r *DualProfiledRing) BackRing() *Ring                   { return r.back }

func (r *DualProfiledRing) Close() error {
	if r == nil {
		return nil
	}
	for _, image := range []*ebiten.Image{r.frontImage, r.backImage, r.maskImage} {
		if image != nil {
			image.Deallocate()
		}
	}
	r.frontImage, r.backImage, r.maskImage = nil, nil, nil
	return nil
}
