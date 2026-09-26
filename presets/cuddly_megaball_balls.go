package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// CuddlyMegaballFormation preserves two ordered nineteen-ball trains with a
// gap in their source indices. The returned orbit, ranges, count and material
// can be edited before construction; live controls use CoupledOrbitController.
func CuddlyMegaballFormation(image *ebiten.Image) sprites.GroupConfig {
	return sprites.GroupConfig{
		Frames: []*ebiten.Image{image}, Count: 38,
		Coupled: &motion.CoupledOrbitFormationConfig{
			Orbit: motion.CoupledOrbit{
				CenterX: 192, CenterY: 135, Radius: 60, DepthRadius: 30, PhaseStep: .00025,
				XIncrement: 1, YIncrement: -2, ZIncrement: -1, QIncrement: -10,
				XOffset: 5, YOffset: 4, ZOffset: 1, QOffset: 246, QScale: 251,
			},
			Ranges: []motion.OrbitRange{{Start: 0, Count: 19, Step: 1}, {Start: 40, Count: 19, Step: 1}},
		},
		AnchorX: .5, AnchorY: .5, Filter: ebiten.FilterNearest,
	}
}
