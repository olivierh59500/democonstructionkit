package presets

import (
	"testing"

	"github.com/olivierh59500/democonstructionkit/timeline"
)

func TestCuddlyIntroStillBlankAndMusicBoundary(t *testing.T) {
	sequence, err := timeline.NewStillThenMain(CuddlyIntroStills())
	if err != nil {
		t.Fatal(err)
	}
	timer := 500
	for tick := 0; tick < 1000; tick++ {
		want := timeline.StillThenMainFrame{Tick: uint64(tick)}
		if timer > 0 {
			if timer > 200 {
				want.Phase = timeline.StillImage
			} else {
				want.Phase = timeline.StillBlank
			}
			timer -= 2
		} else {
			want.Phase = timeline.StillMain
			if timer == 0 {
				want.Cue = true
				timer = -1
			}
		}
		if got := sequence.Next(); got != want {
			t.Fatalf("tick %d = %+v, want %+v", tick, got, want)
		}
	}
	sequence.Reset()
	if got := sequence.Next(); got.Phase != timeline.StillImage || got.Tick != 0 {
		t.Fatalf("reset frame = %+v", got)
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_ = sequence.Next()
	}); allocations != 0 {
		t.Fatalf("intro stage allocated %v times per frame", allocations)
	}
}
