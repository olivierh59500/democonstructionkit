package motion

import (
	"math"
	"testing"
)

func TestWaveClockPreservesDrawBeforeStepBounce(t *testing.T) {
	clock, err := NewWaveClock(WaveClockConfig{Wave: Wave{Amplitude: -80, Speed: 1, Offset: 95, Rectify: true}, Step: .045})
	if err != nil {
		t.Fatal(err)
	}
	phase := 0.0
	for tick := 0; tick < 1000; tick++ {
		want := 95 - math.Abs(math.Sin(phase)*80)
		if got := clock.At(0); math.Abs(got-want) > 1e-12 {
			t.Fatalf("tick %d bounce = %v, want %v", tick, got, want)
		}
		phase += .045
		clock.Step()
	}
	clock.Reset()
	if clock.Phase() != 0 {
		t.Fatal("reset retained wave phase")
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = clock.At(0); clock.Step() }); allocations != 0 {
		t.Fatalf("wave clock sample/step allocates %v times", allocations)
	}
}

func TestWaveClockKeepsPhaseAcrossTempoChangesAndCycleReset(t *testing.T) {
	clock, err := NewWaveClock(WaveClockConfig{Wave: Wave{Amplitude: 158, Speed: 1, Rectify: true}})
	if err != nil {
		t.Fatal(err)
	}
	phase, step := 0.0, 0.0
	for tick := 0; tick < 400; tick++ {
		if got, want := clock.At(0), math.Abs(math.Sin(phase)*158); got != want {
			t.Fatalf("tick %d: %v != %v", tick, got, want)
		}
		switch tick {
		case 20:
			step = .03
		case 150:
			step = .07
		case 300:
			step = .02
		}
		if phase >= 3.1 {
			phase = 0
			clock.Reset()
		}
		if err := clock.SetStep(step); err != nil {
			t.Fatal(err)
		}
		phase += step
		clock.Step()
	}
	if err := clock.SetPhase(1.25); err != nil {
		t.Fatal(err)
	}
	if clock.Phase() != 1.25 {
		t.Fatal("SetPhase did not set the next sample phase")
	}
}

func TestWaveClockHoldPreservesFirstMovingTick(t *testing.T) {
	clock, err := NewWaveClock(WaveClockConfig{Wave: Wave{Amplitude: 768, Offset: 64, Speed: 1, Cos: true}, Start: 1.5, Step: .0125, HoldTicks: 970})
	if err != nil {
		t.Fatal(err)
	}
	phase, hold := 1.5, 970
	for tick := 0; tick < 1200; tick++ {
		if hold > 0 {
			hold--
		} else {
			phase += .0125
		}
		clock.Step()
		if clock.Phase() != phase || clock.HoldRemaining() != hold {
			t.Fatalf("tick %d: phase %v, hold %d; want %v,%d", tick, clock.Phase(), clock.HoldRemaining(), phase, hold)
		}
	}
	clock.Reset()
	if clock.Phase() != 1.5 || clock.HoldRemaining() != 970 {
		t.Fatal("reset lost initial title hold")
	}
	if err := clock.SetHold(3); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		clock.Step()
	}
	if clock.Phase() != 1.5 || clock.HoldRemaining() != 0 {
		t.Fatal("restarted hold advanced phase early")
	}
}
