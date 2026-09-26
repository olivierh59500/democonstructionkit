package presets

import "github.com/olivierh59500/democonstructionkit/modulation"

// CuddlyStarwarsFlash is the reusable 301-tick backdrop opacity pulse.
func CuddlyStarwarsFlash() modulation.PeriodicDecayConfig {
	return modulation.PeriodicDecayConfig{
		Period: 301, TriggerAt: 280, Start: 0,
		Floor: 0, Peak: 20, Rate: .5, Gain: .05,
	}
}
