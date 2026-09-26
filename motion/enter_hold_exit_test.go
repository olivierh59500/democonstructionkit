package motion

import (
	"slices"
	"testing"
)

func TestEnterHoldExitKeepsDeltaSameTickDoubleDraws(t *testing.T) {
	program, err := NewEnterHoldExit(EnterHoldExitConfig{
		StartX: -640, EnterVelocity: 6, EnterBoundary: 32, EnterInclusive: true,
		ExitVelocity: -6, ExitBoundary: -640, ExitInclusive: true, HoldTicks: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	mode, wait, x, dx := SlideEntering, 200, -640.0, 6.0
	var reference [4]SlideEvent
	doubleTicks := 0
	for tick := 0; tick < 1200; tick++ {
		count := 0
		emit := func(kind SlideEventKind) {
			reference[count] = SlideEvent{Kind: kind, X: x}
			count++
		}
		if mode == SlideEntering {
			x += dx
			if x >= 32 {
				dx, mode = -6, SlideHolding
			}
			emit(SlideContent)
		}
		if mode == SlideHolding {
			wait--
			if wait <= 0 {
				mode = SlideLeaving
			}
			emit(SlideContent)
		}
		if mode == SlideLeaving {
			x += dx
			if x <= -640 {
				dx, mode = 0, SlideFollowing
			}
			emit(SlideContent)
		}
		if mode == SlideFollowing {
			emit(SlideFollowingContent)
		}
		if count > 1 {
			doubleTicks++
		}
		got := program.Step()
		if !slices.Equal(got, reference[:count]) || program.State() != (EnterHoldExitState{Stage: mode, X: x, Wait: wait}) {
			t.Fatalf("tick %d events=%+v state=%+v, want events=%+v state=%+v", tick,
				got, program.State(), reference[:count], EnterHoldExitState{Stage: mode, X: x, Wait: wait})
		}
	}
	if doubleTicks < 3 {
		t.Fatalf("only %d same-tick double draws", doubleTicks)
	}
	if got := testing.AllocsPerRun(100, func() { program.Step() }); got != 0 {
		t.Fatalf("enter-hold-exit step allocated %.2f objects", got)
	}
}
