package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// CuddlyFullscreenLogoLayer composes the static logo and source-atop raster
// into one bounded banner surface with independent image/filter parameters.
func CuddlyFullscreenLogoLayer(logo, raster *ebiten.Image) composite.SurfaceLayerConfig {
	return composite.SurfaceLayerConfig{
		Width: 768, Height: 52,
		Passes: []composite.SurfaceImagePass{
			{Image: logo, X: 95, Y: 5, ScaleX: 1.3, ScaleY: 1.3,
				Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver},
			{Image: raster, ScaleX: 1, ScaleY: 1.2,
				Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceAtop},
		},
		Outputs: []composite.SurfaceOutput{{}},
	}
}

// CuddlySpreadpointLogoLayer layers two independently filterable logo images
// around a source-in raster. Passes zero and two receive the same live pose.
func CuddlySpreadpointLogoLayer(inner, raster, outer *ebiten.Image) composite.SurfaceLayerConfig {
	return composite.SurfaceLayerConfig{
		Width: 128, Height: 128,
		Passes: []composite.SurfaceImagePass{
			{Image: inner, Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver},
			{Image: raster, ScaleX: 128, ScaleY: 1,
				Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceIn},
			{Image: outer, Filter: ebiten.FilterLinear, Blend: ebiten.BlendSourceOver},
		},
		Outputs: []composite.SurfaceOutput{{X: 148, Y: 29}},
	}
}
