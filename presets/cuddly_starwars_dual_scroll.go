package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// CuddlyStarwarsDualScroll keeps independent green/red font banks, a sampled
// profile and source-in raster material. Replace either Ring.Font for another
// atlas or edit the profile/strip/filter settings before construction.
func CuddlyStarwarsDualScroll(green, red, raster *ebiten.Image, text string) (scrolling.DualProfiledRingConfig, error) {
	front, err := BitmapFont("cuddly-starwars-scroll", green, ebiten.FilterNearest)
	if err != nil {
		return scrolling.DualProfiledRingConfig{}, err
	}
	back, err := BitmapFont("cuddly-starwars-scroll", red, ebiten.FilterNearest)
	if err != nil {
		return scrolling.DualProfiledRingConfig{}, err
	}
	return scrolling.DualProfiledRingConfig{
		Front:   scrolling.RingConfig{Text: text, Font: front, Viewport: 320, Speed: 8, Controls: true},
		Back:    scrolling.RingConfig{Text: text, Font: back, Viewport: 320, Speed: 8, Controls: true},
		Profile: CuddlyStarwarsProfile(),
		Raster: composite.RasterOverlayConfig{
			Image: raster, ScaleX: 320, ScaleY: 1, Alpha: 1,
			VelocityY: -.5, WrapY: &composite.RasterWrap{Boundary: -72, Restart: 0},
			Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceIn,
		},
		Width: 320, RingHeight: 25, MaskHeight: 200,
		StripCount: 20, StripWidth: 16, SourceHeight: 26, BaseY: 26,
		BackFirstFilter: ebiten.FilterNearest, BackFilter: ebiten.FilterLinear,
		FrontFilter: ebiten.FilterNearest, CompositeFilter: ebiten.FilterNearest,
		Unmanaged: true,
	}, nil
}
