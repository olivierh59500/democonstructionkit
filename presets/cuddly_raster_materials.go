package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// CuddlyMegaScrollerMask colors the live text surface with a borrowed bar mask.
func CuddlyMegaScrollerMask(mask *ebiten.Image) composite.RasterOverlayConfig {
	return composite.RasterOverlayConfig{Image: mask,
		ScaleX: 1, ScaleY: 1, Alpha: 1,
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceAtop}
}

// CuddlyDOCInnerMaterials returns the ordered black and raster passes for the
// inner text. The white pixel and raster image remain caller-owned.
func CuddlyDOCInnerMaterials(white, raster *ebiten.Image) (black, stripes composite.RasterOverlayConfig) {
	var shade ebiten.ColorScale
	shade.Scale(0, 0, 0, 1)
	black = composite.RasterOverlayConfig{Image: white,
		ScaleX: 640, ScaleY: 120, Alpha: 1,
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceAtop,
		ColorScale: shade}
	stripes = composite.RasterOverlayConfig{Image: raster,
		Y: 20, ScaleX: 75, ScaleY: 1, Alpha: 1,
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceAtop}
	return black, stripes
}
