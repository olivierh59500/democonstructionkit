package timeline

import "testing"

func TestPacedIndexMatchesPaletteAndMovementCadence(t *testing.T) {
	palette, err := NewPacedIndex(PacedIndexConfig{Every: 3, First: 4, Count: 84})
	if err != nil {
		t.Fatal(err)
	}
	index, delay := 0, 0
	for tick := 1; tick <= 5000; tick++ {
		if delay >= 3 {
			index = (index + 1) % 84
			delay = 0
		}
		delay++
		if got := palette.Step(); got != index {
			t.Fatalf("palette tick %d = %d, want %d", tick, got, index)
		}
	}
	palette.Reset()
	if palette.Current() != 0 {
		t.Fatal("palette reset kept position")
	}
	character, err := NewPacedIndex(PacedIndexConfig{Every: 5, First: 5, Count: 8})
	if err != nil {
		t.Fatal(err)
	}
	index, delay = 0, 0
	for event := 1; event <= 300; event++ {
		delay++
		if delay >= 5 {
			index = (index + 1) % 8
			delay = 0
		}
		if got := character.Step(); got != index {
			t.Fatalf("movement %d = %d, want %d", event, got, index)
		}
	}
	if allocs := testing.AllocsPerRun(100, func() { _ = character.Step() }); allocs != 0 {
		t.Fatalf("paced index step allocates %v times", allocs)
	}
}

func TestPacedIndexKeepsBarAndCrosshairOneColorApart(t *testing.T) {
	clock, err := NewPacedIndex(PacedIndexConfig{Count: 67, Every: 1, First: 1})
	if err != nil {
		t.Fatal(err)
	}
	for tick := 0; tick < 200; tick++ {
		if bar := clock.Current(); bar != tick%67 {
			t.Fatalf("tick %d bar palette index = %d", tick, bar)
		}
		if crosshair := clock.Step(); crosshair != (tick+1)%67 {
			t.Fatalf("tick %d crosshair palette index = %d", tick, crosshair)
		}
	}
}
