package motion

import "testing"

func TestCuedWaveClockMatchesEhhhLookaheadAndLanding(t *testing.T) {
	config := CuedWaveClockConfig{
		Clock:     WaveClockConfig{Wave: Wave{Amplitude: 158, Speed: 1, Rectify: true}},
		LandingAt: 3.1, HaltAtLanding: true,
		Controls: map[rune]int{'[': 1, '\\': 2, ']': 3, '{': 4},
		Modes: []CuedWaveMode{{},
			{SetStep: true, Step: .03, SetStop: true, ShowSecond: true},
			{SetStep: true, Step: .02, SetStop: true, ShowThird: true},
			{SetStep: true, Step: .07, SetStop: true},
			{SetStop: true, Stop: true}},
	}
	program, err := NewCuedWaveClock(config)
	if err != nil {
		t.Fatal(err)
	}
	reference, err := NewWaveClock(config.Clock)
	if err != nil {
		t.Fatal(err)
	}
	mode, step := 0, 0.0
	show2, show3, stop := false, false, false
	controls := map[int]rune{10: '[', 30: '\\', 180: ']', 260: '{',
		520: '[', 640: ']', 720: '{', 980: '\\', 1120: '{'}
	landings := 0
	for tick := 0; tick < 1400; tick++ {
		if got, want := program.At(), reference.At(0); got != want {
			t.Fatalf("tick %d bounce=%v, want %v", tick, got, want)
		}
		peek := controls[tick]
		switch peek {
		case '[':
			mode, show2 = 1, true
		case '\\':
			mode, show3 = 2, true
		case ']':
			mode = 3
		case '{':
			mode = 4
		}
		switch mode {
		case 1:
			step, stop = .03, false
		case 2:
			step, stop = .02, false
		case 3:
			step, stop = .07, false
		case 4:
			stop = true
		}
		if reference.Phase() >= 3.1 {
			reference.Reset()
			landings++
			if stop {
				step = 0
			}
		}
		if err := reference.SetStep(step); err != nil {
			t.Fatal(err)
		}
		reference.Step()
		if err := program.Step(peek); err != nil {
			t.Fatal(err)
		}
		want := CuedWaveState{Mode: mode, Step: step, Stop: stop,
			ShowSecond: show2, ShowThird: show3, Phase: reference.Phase()}
		if got := program.State(); got != want {
			t.Fatalf("tick %d state=%+v, want %+v", tick, got, want)
		}
	}
	if landings < 2 || !show2 || !show3 {
		t.Fatalf("incomplete cue coverage: landings=%d show2=%v show3=%v", landings, show2, show3)
	}
	if got := testing.AllocsPerRun(100, func() { _ = program.Step('A') }); got != 0 {
		t.Fatalf("cued wave step allocated %.2f objects", got)
	}
}
