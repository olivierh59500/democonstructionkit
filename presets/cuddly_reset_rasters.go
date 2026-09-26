package presets

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// CuddlyResetRasterOrbit preserves the two depth-ordered raster pairs and the
// three phase passes; all counts, thresholds, speed and materials remain data.
func CuddlyResetRasterOrbit(upLeft, upRight, downLeft, downRight *ebiten.Image) composite.PairedRasterOrbitConfig {
	return composite.PairedRasterOrbitConfig{
		Motion: motion.PairedPhaseConfig{
			Count: 16, Spacing: .25, Step: .05, Period: 2 * math.Pi,
			CenterY: 270, AmplitudeY: 200,
			Passes: []motion.PhasePass{
				{From: math.Pi, To: 2 * math.Pi, Material: 0, Alpha: .5},
				{From: 0, To: math.Pi / 2, IncludeFrom: true, IncludeTo: true, Material: 1, Alpha: 1},
				{From: math.Pi / 2, To: math.Pi, IncludeTo: true, Reverse: true, Material: 1, Alpha: 1},
			},
		},
		Pairs: []composite.RasterPair{
			{Left: downLeft, Right: downRight},
			{Left: upLeft, Right: upRight},
		},
		LeftX: 100, RightX: 668, ScaleX: 2, ScaleY: 1,
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver,
	}
}

// CuddlyResetRasterScroll is the vertical source-atop fill offset. The screen
// draws At(0) before Step so the initial raster starts at zero.
func CuddlyResetRasterScroll() motion.WrapBankConfig {
	return motion.WrapBankConfig{
		Start: []float64{0}, Velocity: []float64{-2},
		Lower: &motion.WrapLimit{Boundary: -480, Restart: 3, Inclusive: true},
	}
}
