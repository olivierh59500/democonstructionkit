package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// CuddlyStarwarsSpriteTrain binds eight reversed atlas strips to one authored
// XY path and staggered sinusoidal offsets. All counts, delays, waves and art
// regions can be changed before construction.
func CuddlyStarwarsSpriteTrain(image *ebiten.Image, pathX, pathY []float64) sprites.SampledSpriteTrainConfig {
	regions := make([]composite.Region, 8)
	for i := range regions {
		regions[i] = composite.Region{X: float64(112 - i*16), Width: 16, Height: 10}
	}
	return sprites.SampledSpriteTrainConfig{
		Motion: motion.SampledSpriteTrainConfig{
			PathX: pathX, PathY: pathY, Count: 8, Spacing: 5,
			ExtraAfter: 3, ExtraOffset: 5, WrapAfterLength: true,
			XAmplitude: 10, YAmplitude: 5, XRate: .07, YRate: .09,
			XIndexPhase: 1, YIndexPhase: 1, YCos: true,
		},
		Image: image, Regions: regions, ScaleX: 1, ScaleY: 1,
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver,
	}
}
