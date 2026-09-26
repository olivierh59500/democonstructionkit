package modulation

import "testing"

func TestPeriodicDecayMatchesStarwarsBackdropFlash(t *testing.T) {
	pulse, err := NewPeriodicDecay(PeriodicDecayConfig{
		Period: 301, TriggerAt: 280, Peak: 20, Rate: .5, Gain: .05,
	})
	if err != nil {
		t.Fatal(err)
	}
	counter, fade := 0, 0.0
	triggers := 0
	for tick := 0; tick < 5000; tick++ {
		counter++
		if counter > 300 {
			counter = 0
		}
		if counter == 280 {
			fade = 20
			triggers++
		}
		if fade > 0 {
			fade -= .5
		} else {
			fade = 0
		}
		want := fade * .05
		if got := pulse.Step(); got != want || pulse.Counter() != counter || pulse.Tick() != tick+1 {
			t.Fatalf("tick %d opacity=%v counter=%d, want %v counter=%d", tick,
				got, pulse.Counter(), want, counter)
		}
	}
	if triggers < 2 {
		t.Fatalf("only %d periodic flashes", triggers)
	}
	if got := testing.AllocsPerRun(100, func() { pulse.Step() }); got != 0 {
		t.Fatalf("periodic decay step allocated %.2f objects", got)
	}
}

func TestPeriodicDecayCanUseLiveSignal(t *testing.T) {
	pulse, err := NewPeriodicDecay(PeriodicDecayConfig{Period: 10, TriggerAt: 9,
		Peak: 4, Rate: 1, Gain: 1, Trigger: func(tick int) bool { return tick == 1 }})
	if err != nil {
		t.Fatal(err)
	}
	if a, b := pulse.Step(), pulse.Step(); a != 0 || b != 3 {
		t.Fatalf("live trigger returned %v, %v; want 0, 3", a, b)
	}
}
