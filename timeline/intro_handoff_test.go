package timeline

import (
	"math"
	"testing"
)

func TestIntroHandoffPreservesFadeAndStrictMusicBoundary(t *testing.T) {
	handoff, err := NewIntroHandoff(IntroHandoffConfig{
		FadeStart: 0, FadeStep: .03, FadeMax: 1,
		Cue: IntroCueAboveFade, CueThreshold: .1,
	})
	if err != nil {
		t.Fatal(err)
	}
	for tick := 0; tick < 200; tick++ {
		handoff.Step(false)
		if handoff.Main() || handoff.CueReady() {
			t.Fatal("intro advanced into main without completion")
		}
	}
	handoff.Step(true)
	if !handoff.Main() || !handoff.JustEntered() || handoff.Fade() != 0 || handoff.CueReady() {
		t.Fatalf("entry tick changed fade or cue: %+v", handoff)
	}
	for tick := 1; tick <= 4; tick++ {
		handoff.Step(false)
		if handoff.CueReady() != (tick == 4) {
			t.Fatalf("tick %d music threshold = %v", tick, handoff.CueReady())
		}
	}
	handoff.MarkCue()
	for tick := 0; tick < 100; tick++ {
		handoff.Step(false)
	}
	if handoff.CueReady() || handoff.Fade() != 1 {
		t.Fatalf("fade or one-shot cue after completion: %+v", handoff)
	}
	handoff.Reset()
	if handoff.Main() || handoff.Fade() != 0 || handoff.CueReady() {
		t.Fatalf("reset state = %+v", handoff)
	}
	if allocations := testing.AllocsPerRun(100, func() { handoff.Step(false) }); allocations != 0 {
		t.Fatalf("handoff step allocates %v times", allocations)
	}
}

func TestIntroHandoffCanTriggerImmediatelyOnEntry(t *testing.T) {
	handoff, err := NewIntroHandoff(IntroHandoffConfig{FadeStart: 1, FadeMax: 1, Cue: IntroCueOnEntry})
	if err != nil {
		t.Fatal(err)
	}
	handoff.Step(true)
	if !handoff.CueReady() || handoff.Fade() != 1 {
		t.Fatalf("entry cue was not available: %+v", handoff)
	}
	handoff.Step(false)
	if handoff.CueReady() || handoff.JustEntered() {
		t.Fatal("entry cue repeated on next tick")
	}
}

func TestIntroHandoffRejectsInvalidSettings(t *testing.T) {
	for _, config := range []IntroHandoffConfig{
		{FadeStep: -1},
		{FadeStart: 2, FadeMax: 1},
		{FadeMax: 1, Cue: IntroCue(99)},
		{FadeMax: math.NaN()},
	} {
		if _, err := NewIntroHandoff(config); err == nil {
			t.Fatalf("accepted invalid handoff: %+v", config)
		}
	}
}
