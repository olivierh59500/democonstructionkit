package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/effects"
)

// DOMRasterCopies is the three-image vertical raster cycle over the text mask.
// All source, phase, spacing, speed and wrap values remain editable.
func DOMRasterCopies(image *ebiten.Image) composite.RasterOverlayConfig {
	return composite.RasterOverlayConfig{
		Image: image, Y: 200, VelocityY: -1, ScaleX: 1, ScaleY: 1, Alpha: 1,
		WrapY:  &composite.RasterWrap{Boundary: 0, Restart: 200, Inclusive: true},
		Copies: []composite.RasterCopy{{Y: -200}, {}, {Y: 200}},
	}
}

// DOMScrollMask places raster color through the repeated text alpha, removes
// the top two native rows and positions the result inside the demo viewport.
func DOMScrollMask(raster, text kit.Effect) effects.MaskConfig {
	return effects.MaskConfig{Content: raster, Alpha: text, Width: 640, Height: 400,
		Blend: ebiten.BlendDestinationIn, MaskY: 2, ClearTop: 2,
		OutputX: 64, OutputY: 60}
}
