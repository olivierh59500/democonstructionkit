package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// BilizirCopperBars supplies the shared 10-bit two-clock raster program used
// by Bilizir, Coco and the Coco panel in Multiscreen. The caller can change
// height, rendering primitive, clock policy, phase and source image afterward.
func BilizirCopperBars(image *ebiten.Image, height int, mode composite.CopperDraw, clock composite.CopperClock) composite.CopperBarsConfig {
	return composite.CopperBarsConfig{
		Image: image, Offsets: BilizirCopperOffsets(), Height: height, Count: height / 2,
		RowStep: 2, SourceStep: 2, SourcePeriod: 20,
		BaseX: 60, XShift: 1, VelocityA: 3, VelocityB: -5,
		IndexStepA: 7, IndexStepB: 10, Clock: clock, DrawMode: mode,
	}
}
