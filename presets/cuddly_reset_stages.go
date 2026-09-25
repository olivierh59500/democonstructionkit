package presets

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/timeline"
)

// CuddlyResetStages preserves the four named text-cursor handoffs. Trigger is
// immediate so the next part may draw later in the same simulation frame.
func CuddlyResetStages() timeline.EventStagesConfig {
	return timeline.EventStagesConfig{
		Stages: []string{"plain-a", "plain-b", "masked-a", "showcase", "final"}, Initial: "plain-a",
		Transitions: []timeline.StageTransition{
			{From: "plain-a", Event: "first-end", To: "plain-b"},
			{From: "plain-b", Event: "second-end", To: "masked-a"},
			{From: "masked-a", Event: "third-end", To: "showcase"},
			{From: "showcase", Event: "fourth-end", To: "final"},
		},
	}
}

// CuddlyResetShowcaseWindows selects back fade, raster fade and final hold.
func CuddlyResetShowcaseWindows() []timeline.CueRange {
	return []timeline.CueRange{{Start: 0, End: 41}, {Start: 41, End: 81}, {Start: 81, End: math.Inf(1)}}
}

// CuddlyResetFadeRamp is sampled before each Step to retain its first zero
// alpha frame and forty one displayed samples from zero through full opacity.
func CuddlyResetFadeRamp() timeline.HoldRampConfig { return timeline.HoldRampConfig{Frames: 40} }

// CuddlyResetDirector combines the editable stage, timing and fade recipes in
// one DCK controller while leaving images, fonts and draw order to the screen.
func CuddlyResetDirector() timeline.StageSequenceConfig {
	return timeline.StageSequenceConfig{
		Stages: CuddlyResetStages(), TimedStage: "showcase", Rate: 1,
		Windows: CuddlyResetShowcaseWindows(),
		Fades:   []timeline.HoldRampConfig{CuddlyResetFadeRamp(), CuddlyResetFadeRamp(), {}},
	}
}
