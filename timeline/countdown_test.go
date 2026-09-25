package timeline

import "testing"

func TestCountdownPreservesTwoStageBoundaries(t *testing.T) {
	countdown, err := NewCountdown(CountdownConfig{First: 49, Second: 162, Hold: 24, FadeLead: 42})
	if err != nil {
		t.Fatal(err)
	}
	if countdown.BlankTick() != 235 || countdown.FadeStartTick() != 169 {
		t.Fatalf("thresholds = %d, %d", countdown.BlankTick(), countdown.FadeStartTick())
	}
	for _, check := range []struct {
		tick, first, second int
		blank               bool
	}{
		{0, 49, 162, false}, {48, 1, 162, false}, {49, 0, 161, false},
		{169, 0, 41, false}, {234, 0, -24, false}, {235, 0, -25, true},
	} {
		got := countdown.At(check.tick)
		if got.First != check.first || got.Second != check.second || got.Blank != check.blank {
			t.Fatalf("tick %d = %+v, want %+v", check.tick, got, check)
		}
	}
}

func TestCountdownRejectsOverflowingOrNegativeValues(t *testing.T) {
	for _, config := range []CountdownConfig{
		{First: -1}, {First: 1, Second: -2}, {First: int(^uint(0) >> 1), Second: 1},
	} {
		if _, err := NewCountdown(config); err == nil {
			t.Fatalf("accepted invalid countdown: %+v", config)
		}
	}
}
