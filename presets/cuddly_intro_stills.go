package presets

import "github.com/olivierh59500/democonstructionkit/timeline"

// CuddlyIntroStills replaces the 500-to-zero counter advancing by two:
// Calvin is visible for 150 ticks, followed by 100 blank ticks.
func CuddlyIntroStills() timeline.StillThenMainConfig {
	return timeline.StillThenMainConfig{StillTicks: 150, BlankTicks: 100}
}
