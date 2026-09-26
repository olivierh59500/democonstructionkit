package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// CuddlyEhhhLogoRows uses the same sampled-row effect as Digi with an
// independent curve and no vertical bounce. Transparent rows beyond the art's
// source height remain part of the original 170-step draw order.
func CuddlyEhhhLogoRows(image *ebiten.Image) (composite.TableWarpLogoConfig, error) {
	curve, err := CuddlyEhhhProfile()
	if err != nil {
		return composite.TableWarpLogoConfig{}, err
	}
	return composite.TableWarpLogoConfig{
		Image: image, SourceX: 0, Width: 767,
		Motion: motion.TableWarpRowsConfig{
			Curve: curve, Rows: 170, CounterStart: 0, Step: 1,
			XBase: 384, HalfWidth: 383.5,
			YBase: 0, RowOffset: -.5, SourceHeight: 1,
			BounceDivisor: 1, Bounce: motion.WaveClockConfig{},
		},
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver,
	}, nil
}
