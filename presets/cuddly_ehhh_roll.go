package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// CuddlyEhhhRoller keeps the four authored lookahead characters, their sticky
// text lanes, per-mode phase speed and stop-at-landing behavior editable.
func CuddlyEhhhRoller() motion.CuedWaveClockConfig {
	return motion.CuedWaveClockConfig{
		Clock: RectifiedSine(0, 158, 0), InitialMode: 0,
		LandingAt: 3.1, HaltAtLanding: true,
		Controls: map[rune]int{'[': 1, '\\': 2, ']': 3, '{': 4},
		Modes: []motion.CuedWaveMode{
			{},
			{SetStep: true, Step: .03, SetStop: true, Stop: false, ShowSecond: true},
			{SetStep: true, Step: .02, SetStop: true, Stop: false, ShowThird: true},
			{SetStep: true, Step: .07, SetStop: true, Stop: false},
			{SetStop: true, Stop: true},
		},
	}
}
