package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// TeamG1LogoFormation configures twelve independently scaled image instances.
// Counts, radii, harmonics, image frames, tint signals and timing can be changed
// through the returned GroupConfig before constructing sprites.Group.
func TeamG1LogoFormation(image *ebiten.Image, width, height float64) sprites.GroupConfig {
	return sprites.GroupConfig{
		Frames: []*ebiten.Image{image}, Count: 12,
		Origin:    motion.Point{X: width / 2, Y: height / 2},
		PhaseStep: .02, AnchorX: .5, AnchorY: .5,
		Circle: &motion.CircleFormation{
			RadiusX: 150, RadiusY: 150, IndexCount: 12,
			XAmplitude: 20, XRate: 2, XIndexPhase: 1,
			YAmplitude: 20, YRate: 2, YIndexPhase: 1,
			ScaleBase: .5, ScaleAmplitude: .5,
			ScaleRate: 1, ScaleIndexPhase: .5,
		},
	}
}
