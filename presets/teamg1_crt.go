package presets

import "github.com/olivierh59500/democonstructionkit/effects"

// TeamG1TimedCRT preserves its animated scanline and phosphor profile while
// exposing every shader parameter and its simulation clock for reuse.
func TeamG1TimedCRT() effects.TimedCRTConfig {
	return effects.TimedCRTConfig{
		Curvature: .25, ScanlineFrequency: 800, ScanlineAmplitude: .04,
		ScanlineTimeRate: 2, ChromaticShift: .003, Vignette: .7,
		GlowShift: .001, GlowGain: .1,
		FlickerBase: .95, FlickerAmplitude: .05, FlickerRate: 120,
		TimeStep: .016,
	}
}
