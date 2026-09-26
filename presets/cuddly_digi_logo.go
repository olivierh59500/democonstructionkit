package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// CuddlyDigiLogo compiles the authored overlapping curve once and binds it to
// 170 independently sampled image rows and the shared rectified bounce.
func CuddlyDigiLogo(image *ebiten.Image) (composite.TableWarpLogoConfig, error) {
	curve, err := motion.CompileWaveProgram(CuddlyDigiWaveProgram()...)
	if err != nil {
		return composite.TableWarpLogoConfig{}, err
	}
	return composite.TableWarpLogoConfig{
		Image: image, SourceX: 0, Width: 335,
		Motion: motion.TableWarpRowsConfig{
			Curve: curve, Rows: 170, CounterStart: 0, Step: 1,
			XBase: 320, HalfWidth: 167.5,
			YBase: 160, RowOffset: -.5, SourceHeight: 1,
			BounceDivisor: 2, Bounce: RectifiedSine(200, -160, .03),
		},
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver,
	}, nil
}
