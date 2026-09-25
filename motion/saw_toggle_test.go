package motion

import "testing"

func TestSawTogglePreservesStrictWrapAndFaceParity(t *testing.T) {
	cycle, err := NewSawToggle(SawToggleConfig{Start: 0, Velocity: .08, Boundary: 1, Restart: -1})
	if err != nil {
		t.Fatal(err)
	}
	value, face := 0.0, false
	for tick := 0; tick < 1500; tick++ {
		value += .08
		if value > 1 {
			value = -1
			face = !face
		}
		cycle.Step()
		if cycle.At() != value || cycle.Alternate() != face {
			t.Fatalf("tick %d: (%v,%v), want (%v,%v)", tick, cycle.At(), cycle.Alternate(), value, face)
		}
	}
	cycle.Reset()
	if cycle.At() != 0 || cycle.Alternate() {
		t.Fatal("reset retained phase or face")
	}
	if allocations := testing.AllocsPerRun(100, cycle.Step); allocations != 0 {
		t.Fatalf("saw step allocates %v times", allocations)
	}
}

func TestSawToggleRejectsInvalidDirection(t *testing.T) {
	for _, config := range []SawToggleConfig{
		{}, {Velocity: 1, Boundary: 1, Restart: 1},
		{Velocity: -1, Boundary: 0, Restart: -1},
	} {
		if _, err := NewSawToggle(config); err == nil {
			t.Fatalf("accepted invalid saw %+v", config)
		}
	}
}
