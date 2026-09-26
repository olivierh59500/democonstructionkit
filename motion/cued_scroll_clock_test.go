package motion

import "testing"

func TestCuedScrollClockMatchesPhenomenaPauseAndRotation(t *testing.T) {
	clock, err := NewCuedScrollClock(CuedScrollClockConfig{
		InitialTextStep: 1, InitialPauseTicks: 250, InitialRotationStep: .35,
		ResumeTextStep: 1, ResumeRotationStep: .35, RotationFrames: 30,
		Cues: map[string]ScrollCue{
			"^": {PauseTicks: 275, SetRotation: true, RotationStep: -1},
			"&": {PauseTicks: 275, SetRotation: true, RotationStep: 1},
			"#": {PauseTicks: 250, SetRotation: true, RotationStep: -1},
			"%": {PauseTicks: 225, SetRotation: true, RotationStep: -1},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	events := map[int]string{4: "^", 8: "&", 12: "#", 16: "%"}
	cursor, referenceCursor := 0, 0
	reference := CuedScrollState{TextStep: 1, PauseTicks: 250, RotationStep: .35}
	insert := func(count int) {
		for range count {
			cursor++
			if name := events[cursor]; name != "" {
				clock.OnControl(name)
			}
		}
	}
	for tick := 1; tick <= 1400; tick++ {
		if !reference.Paused {
			referenceCursor += reference.TextStep
			if name := events[referenceCursor]; name != "" {
				reference.Paused = true
				switch name {
				case "^":
					reference.PauseTicks, reference.RotationStep = 275, -1
				case "&":
					reference.PauseTicks, reference.RotationStep = 275, 1
				case "#":
					reference.PauseTicks, reference.RotationStep = 250, -1
				case "%":
					reference.PauseTicks, reference.RotationStep = 225, -1
				}
			}
		} else {
			reference.PauseTicks--
			if reference.PauseTicks == 0 {
				reference.Paused = false
				reference.RotationStep = .35
				reference.TextStep = 1
			}
		}
		reference.Rotation += reference.RotationStep
		if reference.Rotation >= 30 {
			reference.Rotation -= 30
		}
		if reference.Rotation < 0 {
			reference.Rotation += 30
		}
		if err := clock.Step(insert); err != nil {
			t.Fatal(err)
		}
		if clock.State() != reference || cursor != referenceCursor {
			t.Fatalf("tick %d: state=%+v cursor=%d, want %+v cursor=%d", tick,
				clock.State(), cursor, reference, referenceCursor)
		}
	}
	if got := testing.AllocsPerRun(100, func() { _ = clock.Step(insert) }); got != 0 {
		t.Fatalf("cued scroll clock allocated %.2f objects", got)
	}
}

func TestCuedScrollClockStopsBudgetAndRejectsInvalidCue(t *testing.T) {
	clock, err := NewCuedScrollClock(CuedScrollClockConfig{RotationFrames: 30,
		Cues: map[string]ScrollCue{"stop": {StopBudget: true}}})
	if err != nil {
		t.Fatal(err)
	}
	if !clock.OnControl("stop") || clock.OnControl("unknown") {
		t.Fatal("control budget policy changed")
	}
	if _, err := NewCuedScrollClock(CuedScrollClockConfig{RotationFrames: 30,
		Cues: map[string]ScrollCue{"bad": {PauseTicks: -1}}}); err == nil {
		t.Fatal("accepted a negative pause")
	}
}
