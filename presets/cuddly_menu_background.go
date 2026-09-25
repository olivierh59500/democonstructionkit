package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// CuddlyMenuTileParallax keeps one tile of overscan around the menu viewport.
// The camera moves at half speed with integer pixels and a strict tile wrap.
func CuddlyMenuTileParallax(tile *ebiten.Image, width, height, period int) composite.CachedTileParallaxConfig {
	return composite.CachedTileParallaxConfig{
		Tile: tile, Width: width, Height: height,
		PeriodX: period, PeriodY: period, OverscanX: period, OverscanY: period,
		DivisorX: 2, DivisorY: 2, WrapX: float64(period), WrapY: float64(period),
		ClampNegative: true, PixelQuantize: true, Unmanaged: true,
		Blend: ebiten.BlendCopy,
	}
}
