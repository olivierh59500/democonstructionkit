package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// VivaTitleMotion is an editable horizontal cosine path for any title image.
// Standalone Viva holds for 970 ticks; Multiscreen starts moving immediately.
func VivaTitleMotion(width float64, holdTicks int) motion.WaveClockConfig {
	return motion.WaveClockConfig{
		Wave:  motion.Wave{Amplitude: width, Offset: 64, Speed: 1, Cos: true},
		Start: 1.5, Step: .0125, HoldTicks: holdTicks,
	}
}
