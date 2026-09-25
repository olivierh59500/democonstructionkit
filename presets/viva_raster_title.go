package presets

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// VivaRasterTitleCanvas preserves the standalone 528×36 raster material.
// Change the returned scales, source crops or phases for another title.
func VivaRasterTitleCanvas(title, raster *ebiten.Image) composite.RasterTitleConfig {
	return composite.RasterTitleConfig{
		Title: title, Raster: raster, Motion: VivaRasterWrapConfig(),
		Mode: composite.RasterTitleCanvas, CanvasSize: title.Bounds().Size(),
		RasterScaleX: 24, RasterScaleY: 1, TitleScaleX: 1, TitleScaleY: 1,
		ThirdOffsetY: 72,
	}
}

// VivaRasterTitleDirect keeps the embedded panel's clipped 2× material and
// draws directly into the destination instead of allocating a title canvas.
func VivaRasterTitleDirect(title, raster *ebiten.Image, width int) composite.RasterTitleConfig {
	w := float64(width)
	return composite.RasterTitleConfig{
		Title: title, Raster: raster, Motion: VivaRasterWrapConfig(),
		Mode: composite.RasterTitleDirect, Clip: image.Rect(0, 14, width, 86),
		UnderlayWidth: w, UnderlayHeight: 72,
		RasterScaleX: w / float64(raster.Bounds().Dx()), RasterScaleY: 2,
		TitleScaleX: w / float64(title.Bounds().Dx()), TitleScaleY: 2,
		ThirdOffsetY: 72,
	}
}
