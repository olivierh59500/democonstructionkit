package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// TCBLogoRowProfile applies an editable displacement table to a cropped logo.
// Native left/top coordinates are 8,96; callers may change placement, scale,
// phase, wrap or filter for another image or viewport.
func TCBLogoRowProfile(offsets []float64, sourceWidth float64) composite.ProfileImageConfig {
	return composite.ProfileImageConfig{
		Offsets: offsets, PhaseStep: 1, RowStep: 1, PhaseWrap: len(offsets) - 80,
		Gain: 1, BaseX: 8 + sourceWidth/2, BaseY: 96,
		ScaleX: 1, ScaleY: 1, Filter: ebiten.FilterNearest, Batch: true,
	}
}
