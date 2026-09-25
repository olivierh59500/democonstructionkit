package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// GrodanSpriteFormationConfig recreates the twelve-sprite oscillating train.
// The caller supplies its two phase clocks and bouncing vertical envelope;
// all image frames, count and drawing scale remain choices of sprites.Group.
func GrodanSpriteFormationConfig() motion.HarmonicFormationConfig {
	return motion.HarmonicFormationConfig{
		Origin: motion.Point{X: 304, Y: 100},
		X:      []motion.IndexedHarmonic{{Amplitude: 290, Rate: 1, IndexPhase: -.2, Cos: true}},
		Y: []motion.IndexedHarmonic{
			{Amplitude: 1, Rate: 1, IndexPhase: -.2, SecondaryClock: true, Envelope: true},
			{Amplitude: 1, Rate: 1, SecondaryClock: true, Envelope: true},
		},
	}
}

// MegaTwistSpriteFormationConfig recreates the four independent harmonic
// curves and per-index ripples of the glowing sprite ensemble. Dimensions and
// sprite size define the editable center and clipping bounds.
func MegaTwistSpriteFormationConfig(width, height, spriteSize float64) motion.HarmonicFormationConfig {
	half := spriteSize / 2
	return motion.HarmonicFormationConfig{
		Origin: motion.Point{X: width / 2, Y: height / 2},
		Bounds: &motion.FormationBounds{
			Min: motion.Point{X: half, Y: half},
			Max: motion.Point{X: width - half, Y: height - half},
		},
		X: []motion.IndexedHarmonic{
			{Amplitude: 100, Rate: 1.35, IndexPhase: .155, Phase: 1.25},
			{Amplitude: 100, Rate: 1.86, IndexPhase: .155, Phase: .54},
			{Amplitude: 20, IndexRate: .289, Phase: 1.15},
		},
		Y: []motion.IndexedHarmonic{
			{Amplitude: 60, Rate: 1.72, IndexPhase: .155, Phase: .23, Cos: true},
			{Amplitude: 60, Rate: 1.63, IndexPhase: .155, Phase: .98, Cos: true},
			{Amplitude: 20, IndexRate: .456, Phase: .85, Cos: true},
		},
	}
}

// MegaTwistGlowPainterConfig returns the three-pass halo and final sprite
// material. The caller can change any layer count, alpha, scale or filter.
func MegaTwistGlowPainterConfig(spriteSize, zoom float64) sprites.GlowPainterConfig {
	return sprites.GlowPainterConfig{
		PositionScale: zoom, BaseScale: zoom, LayerScaleStep: .1,
		AnchorX: spriteSize / 2, AnchorY: spriteSize / 2,
		Layers: 3, GlowAlpha: .3, GlowFilter: ebiten.FilterLinear,
	}
}
