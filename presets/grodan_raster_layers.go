package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// GrodanRasterLayers composes the big, vertical and paired small ribbons into
// their three original bounded surfaces. Every pass and output remains editable.
func GrodanRasterLayers(scrolls [4]*scrolling.Scrolling,
	big, vertical, smallTop, smallBottom *ebiten.Image) [3]composite.SurfaceLayerConfig {
	pass := func(image *ebiten.Image, x, y, sx, sy float64) composite.SurfaceImagePass {
		return composite.SurfaceImagePass{Image: image, X: x, Y: y,
			ScaleX: sx, ScaleY: sy, Blend: ebiten.BlendSourceAtop}
	}
	verticalOutputs := make([]composite.SurfaceOutput, 0, 6)
	for _, x := range [...]float64{0, 64, 128, 480, 544, 608} {
		verticalOutputs = append(verticalOutputs, composite.SurfaceOutput{X: x})
	}
	return [3]composite.SurfaceLayerConfig{
		{Width: 640, Height: 200, Sources: []kit.Effect{scrolls[0]},
			Passes:  []composite.SurfaceImagePass{pass(big, 0, 0, 4, 2)},
			Outputs: []composite.SurfaceOutput{{Y: 200}}, OwnSources: true},
		{Width: 32, Height: 400, Sources: []kit.Effect{scrolls[1]},
			Passes:  []composite.SurfaceImagePass{pass(vertical, 0, 0, 2, 2)},
			Outputs: verticalOutputs, OwnSources: true},
		{Width: 320, Height: 32, Sources: []kit.Effect{scrolls[2], scrolls[3]},
			Passes: []composite.SurfaceImagePass{
				pass(smallTop, 0, 0, 2, 2), pass(smallBottom, 0, 24, 2, 2)},
			Outputs: []composite.SurfaceOutput{{Y: 16, ScaleX: 2, ScaleY: 2}}, OwnSources: true},
	}
}
