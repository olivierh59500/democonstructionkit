package recipes

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestStarwarsProfileKeepsEverySampleAndTailWrap(t *testing.T) {
	profile, err := motion.NewSegmentedProfile(CuddlyStarwarsProfile())
	if err != nil {
		t.Fatal(err)
	}
	var want []float64
	absolute := motion.Wave{Amplitude: 50, Speed: 1, Rectify: true}
	for index, segment := range []int{100, 100, 50, 60, 30, 50, 30, 72, 60, 50, 30, 100} {
		phase := 0.0
		step := 2 * math.Pi / float64(segment)
		for range segment {
			value := 50*math.Sin(phase)/2 + .5
			if index == 3 {
				value = absolute.At(0, phase)
			}
			want = append(want, math.Floor(value+.5))
			phase += step
		}
	}
	want = append(want, 0)
	if len(profile.Table()) != len(want) {
		t.Fatalf("profile length %d, want %d", len(profile.Table()), len(want))
	}
	for i, value := range want {
		if profile.Table()[i] != value {
			t.Fatalf("profile sample %d = %v, want %v", i, profile.Table()[i], value)
		}
	}
	index, wraps := 20, 0
	for tick := 0; tick < 5000; tick++ {
		index++
		if index > len(want)-80 {
			index = 20
			wraps++
		}
		profile.Step()
		if profile.Index() != index {
			t.Fatalf("tick %d index %d, want %d", tick, profile.Index(), index)
		}
		for strip := 0; strip < 20; strip++ {
			if got := profile.At(strip); got != want[index+strip] {
				t.Fatalf("tick %d strip %d = %v, want %v", tick, strip, got, want[index+strip])
			}
		}
	}
	if wraps < 2 {
		t.Fatalf("only %d profile wraps", wraps)
	}
	if got := testing.AllocsPerRun(100, profile.Step); got != 0 {
		t.Fatalf("profile step allocated %.2f objects", got)
	}
}
