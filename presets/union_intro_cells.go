package presets

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// UnionIntroTextCells deforms the lettering with editable row and column
// waves. The remaining two cell-wave channels stay available to variations.
func UnionIntroTextCells() composite.HarmonicCellWarpConfig {
	return composite.HarmonicCellWarpConfig{
		Cell: image.Pt(32, 16), Filter: ebiten.FilterLinear,
		Waves: composite.CellWaveBank{
			XRows:    motion.Waves{{Amplitude: 32, Spatial: .3, Speed: .08}},
			YColumns: motion.Waves{{Amplitude: 16, Spatial: .3, Speed: .08}},
		},
	}
}
