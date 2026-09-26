package motion

import (
	"math"
	"slices"
	"testing"
)

func TestLatchedTriggersMatchKnucklebusterRandomHits(t *testing.T) {
	random := func() func() float64 {
		seed := uint32(42)
		return func() float64 {
			seed = seed*1664525 + 1013904223
			return float64(seed) / 4294967296
		}
	}
	actualRandom, referenceRandom := random(), random()
	actualRandom() // Scene initialization consumes one seeded value.
	referenceRandom()
	hits, err := NewLatchedTriggers(LatchedTriggersConfig{Count: 3, Period: 5,
		ReleaseAt: 1, Source: TriggerRandom, RandomFloat: actualRandom,
		RandomRange: 750, RandomBias: 1, Threshold: 725})
	if err != nil {
		t.Fatal(err)
	}
	var on [3]bool
	var triggers [3]float64
	timer, visibleCount := 0, 0
	for tick := 0; tick < 5000; tick++ {
		if timer == 0 {
			for i := range on {
				triggers[i] = math.Floor(referenceRandom()*750) + 1
			}
			timer = 5
		}
		for i, value := range triggers {
			if value >= 725 {
				on[i] = true
			}
		}
		want := on
		for _, active := range want {
			if active {
				visibleCount++
			}
		}
		timer--
		if timer == 1 {
			on = [3]bool{}
		}
		got := hits.Step()
		if !slices.Equal(got, want[:]) || hits.Timer() != timer || hits.Tick() != tick+1 {
			t.Fatalf("tick %d visibility=%v timer=%d, want %v timer=%d", tick,
				got, hits.Timer(), want, timer)
		}
	}
	if visibleCount < 20 {
		t.Fatalf("only %d active hit frames", visibleCount)
	}
	if got := testing.AllocsPerRun(100, func() { hits.Step() }); got != 0 {
		t.Fatalf("latched trigger step allocated %.2f objects", got)
	}
}

func TestLatchedTriggersAcceptMusicSignalMode(t *testing.T) {
	hits, err := NewLatchedTriggers(LatchedTriggersConfig{Count: 2, Period: 4,
		ReleaseAt: 1, Source: TriggerSignal, Threshold: 1,
		SignalEveryTick: true, Signal: func(index, tick int) bool { return index == 0 && tick == 0 }})
	if err != nil {
		t.Fatal(err)
	}
	for tick, want := range [][2]bool{{true, false}, {true, false}, {true, false}, {false, false}} {
		got := hits.Step()
		if got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("tick %d signal visibility=%v, want %v", tick, got, want)
		}
	}
}
