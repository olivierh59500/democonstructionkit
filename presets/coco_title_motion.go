package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// CocoTitleMotion shares Viva's horizontal cosine title path, starting at
// Coco's authored phase without the long entrance hold.
func CocoTitleMotion(width float64) motion.WaveClockConfig {
	config := VivaTitleMotion(width, 0)
	config.Start = .5
	return config
}
