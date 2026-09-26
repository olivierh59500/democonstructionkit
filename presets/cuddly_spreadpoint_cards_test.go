package presets

import (
	"image/color"
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/timeline"
)

func TestCuddlySpreadpointCardColorsAndMusicCuesMatchEveryTick(t *testing.T) {
	cycle, err := timeline.NewTintedCards(CuddlySpreadpointCards())
	if err != nil {
		t.Fatal(err)
	}
	card, iteration := 0, 0
	for tick := 0; tick < 4*181; tick++ {
		red, green, cue := 0, 0, -1
		handoff := false
		switch {
		case iteration < 20:
			red = int(math.Floor(float64(iteration) * 255 / 19))
			green = red
			iteration++
		case iteration < 160:
			red = 255
			green = int(math.Floor(34 + float64(159-iteration)*221/139))
			iteration++
		case iteration < 180:
			n := iteration - 160
			red = int(math.Floor(float64(19-n) * 255 / 19))
			green = int(math.Floor(float64(19-n) * 34 / 19))
			iteration++
		default:
			iteration = 0
			card++
			cue = card
			handoff = true
		}
		want := color.RGBA{R: uint8(red), G: uint8(green), B: uint8(green), A: 255}
		frame := cycle.Next()
		if frame.Card != card || frame.Tint != want || frame.Cue != cue || frame.Handoff != handoff || frame.Completed != (card == 4) {
			t.Fatalf("tick %d frame = %+v, want card=%d tint=%+v cue=%d handoff=%t completed=%t", tick,
				frame, card, want, cue, handoff, card == 4)
		}
	}
	if !cycle.Completed() || cycle.Next().Cue != -1 {
		t.Fatal("card sequence did not stop after the final music cue")
	}
	if allocations := testing.AllocsPerRun(100, func() {
		cycle.Reset()
		for range 181 {
			_ = cycle.Next()
		}
	}); allocations != 0 {
		t.Fatalf("card sequence allocated %v times per cycle", allocations)
	}
}
