package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// CuddlyMegaScrollerMaskTiles keeps the 48 ordered, eight-pixel-spaced copies
// of the oversized white-bar image. Background draws only visible copies.
func CuddlyMegaScrollerMaskTiles() composite.BackgroundConfig {
	return composite.BackgroundConfig{
		PeriodX: 8, CopiesX: 48,
		Filter: ebiten.FilterLinear, Blend: ebiten.BlendSourceOver,
	}
}
