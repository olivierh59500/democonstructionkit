package timeline

import (
	"testing"
	"time"
)

func TestSceneBoundaryAndLoop(t *testing.T) {
	s, err := NewSequence([]time.Duration{time.Second, 2 * time.Second}, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		at    time.Duration
		index int
		local time.Duration
	}{{0, 0, 0}, {time.Second, 1, 0}, {3 * time.Second, 0, 0}, {6500 * time.Millisecond, 0, 500 * time.Millisecond}} {
		i, local, ok := s.At(tt.at)
		if !ok || i != tt.index || local != tt.local {
			t.Fatal(tt, i, local, ok)
		}
	}
	s.loop = false
	if _, _, ok := s.At(s.Duration()); ok {
		t.Fatal("end must be exclusive")
	}
}
func TestClockPauseAndSpeed(t *testing.T) {
	c, _ := NewClock(50)
	first := c.Step()
	if first.Time != 0 || first.Delta != .02 {
		t.Fatal(first)
	}
	c.Pause(true)
	a, b := c.Step(), c.Step()
	if a.Time != b.Time || b.Delta != 0 {
		t.Fatal(a, b)
	}
	c.Pause(false)
	if err := c.SetSpeed(2); err != nil {
		t.Fatal(err)
	}
	if c.Step().Delta != .04 {
		t.Fatal("wrong delta")
	}
}
