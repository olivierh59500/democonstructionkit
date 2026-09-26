package presets

import (
	"image/color"

	"github.com/olivierh59500/democonstructionkit/timeline"
)

// CuddlySpreadpointCards keeps the three exact channel ramps and one black
// handoff tick per card. Cue 4 starts the main screen after the final handoff.
func CuddlySpreadpointCards() timeline.TintedCardsConfig {
	black := color.RGBA{A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	amber := color.RGBA{R: 255, G: 34, B: 34, A: 255}
	return timeline.TintedCardsConfig{
		Count: 4, HandoffTicks: 1,
		Ramps: []timeline.CardTintRamp{
			{Frames: 20, From: black, To: white},
			{Frames: 140, From: white, To: amber, Reverse: true},
			{Frames: 20, From: amber, To: black, Reverse: true},
		},
	}
}
