package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// BilizirWarpSampleOrigin cancels the native scrolling image's initial X crop.
const BilizirWarpSampleOrigin = 64

// BilizirWarpClockConfig compiles the six authored horizontal table sections
// and exposes independent vertical phase, spatial rate and amplitude settings.
func BilizirWarpClockConfig() (motion.WarpTableClockConfig, error) {
	table, err := motion.CompileWaveTable(BilizirWaveSections()...)
	if err != nil {
		return motion.WarpTableClockConfig{}, err
	}
	return motion.WarpTableClockConfig{
		Horizontal: table, SourceOrigin: BilizirWarpSampleOrigin,
		VerticalStep: .1, VerticalSpatial: .1, VerticalAmplitude: 35,
	}, nil
}
