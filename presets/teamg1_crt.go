package presets

import "github.com/olivierh59500/democonstructionkit/effects"

// TeamG1TimedCRT keeps animated scanlines, color fringe, glow and flicker.
// A flat projection leaves all rows of the tightly fitted intro glyphs visible.
func TeamG1TimedCRT() effects.TimedCRTConfig {
	return effects.TimedCRTConfig{
		Curvature: 0, ScanlineFrequency: 800, ScanlineAmplitude: .04,
		ScanlineTimeRate: 2, ChromaticShift: .003, Vignette: .7,
		GlowShift: .001, GlowGain: .1,
		FlickerBase: .95, FlickerAmplitude: .05, FlickerRate: 120,
		TimeStep: .016,
	}
}
