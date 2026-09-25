package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// RectifiedSine is an editable absolute-sine bob for text, logos or sprites.
// Signed amplitude selects a downward or upward excursion from offset.
func RectifiedSine(offset, amplitude, step float64) motion.WaveClockConfig {
	return motion.WaveClockConfig{
		Wave: motion.Wave{Amplitude: amplitude, Speed: 1, Offset: offset, Rectify: true},
		Step: step,
	}
}
