package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// PhenomenaDNARows is shared by the standalone intro and its Multiscreen panel.
// The renderer chooses its own slice scale and origin for each viewport.
func PhenomenaDNARows() motion.RecurrentRowWaveConfig {
	return motion.RecurrentRowWaveConfig{
		Base: 67, Flat: 80, Amplitude: 80,
		RevealTime: 5 * 50, IndexDelay: .0033, SampleTimeStep: 1.0 / 6.0,
		StartAngle: 5 * 10.50, TimeDivisor: 6, AngleStep: 1.0 / 36.0,
	}
}
