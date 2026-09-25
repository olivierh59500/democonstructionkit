package presets

import (
	"image"

	"github.com/olivierh59500/democonstructionkit/composite"
)

// TCBMountainBands preserves the standalone and Multiscreen strip order,
// half-pixel velocities, strict source crops and integer phase truncation.
// Images and output placement remain caller-owned.
func TCBMountainBands() composite.BandsConfig {
	transport := TCBMountainWrapConfig()
	config := composite.BandsConfig{CopyOffsets: [][2]float64{{0, 0}, {640, 0}}}
	for index, velocity := range transport.Velocity {
		y := index * 10
		if index >= 16 {
			y += 84
		}
		config.Bands = append(config.Bands, composite.MovingBand{
			Source: image.Rect(0, index*10, 1024, index*10+10),
			Y:      float64(y), VelocityX: velocity, WrapX: 256,
			MotionScaleX: 2, TruncatePhaseX: true,
		})
	}
	return config
}
