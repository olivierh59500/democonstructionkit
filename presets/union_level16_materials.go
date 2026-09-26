package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// UnionLevel16Water is an ordinary source-over image with a positive-Y wrap.
// The borrowed image, placement, cadence and boundary remain editable.
func UnionLevel16Water(image *ebiten.Image) composite.RasterOverlayConfig {
	return composite.RasterOverlayConfig{
		Image: image, X: 20, Y: 0, VelocityY: 2,
		ScaleX: 1, ScaleY: 1, Alpha: 1,
		WrapY:  &composite.RasterWrap{Boundary: 220, Restart: 0, Inclusive: true},
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver,
	}
}

// UnionLevel16Raster moves independently of the water with a negative-Y wrap.
func UnionLevel16Raster(image *ebiten.Image) composite.RasterOverlayConfig {
	return composite.RasterOverlayConfig{
		Image: image, X: 300, Y: 120, VelocityY: -2,
		ScaleX: 1, ScaleY: 1, Alpha: 1,
		WrapY:  &composite.RasterWrap{Boundary: -25, Restart: 120, Inclusive: true},
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver,
	}
}
