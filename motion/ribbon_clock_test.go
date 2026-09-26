package motion

import "testing"

func TestRibbonClockKeepsStrictHorizontalAndVerticalRestarts(t *testing.T) {
	horizontal, err := NewRibbonClock(RibbonClockConfig{
		Length: 96, Offset: 0, Velocity: -4, Restart: 640,
		Multiplier: 1, Wrap: RibbonWrapBelow,
	})
	if err != nil {
		t.Fatal(err)
	}
	for tick := 1; tick <= 25; tick++ {
		if err := horizontal.Step(); err != nil {
			t.Fatal(err)
		}
		want := -float64(tick * 4)
		if tick == 25 {
			want = 640
		}
		if got := horizontal.Offset(); got != want {
			t.Fatalf("horizontal tick %d offset %v, want %v", tick, got, want)
		}
	}
	vertical, err := NewRibbonClock(RibbonClockConfig{
		Length: 200, Offset: -100, Velocity: 3, Restart: -100,
		Extent: 400, Multiplier: 1, Wrap: RibbonWrapAbove,
	})
	if err != nil {
		t.Fatal(err)
	}
	for tick := 1; tick <= 234; tick++ {
		if err := vertical.Step(); err != nil {
			t.Fatal(err)
		}
		want := -100.0 + float64(tick*3)
		if tick == 234 {
			want = -100
		}
		if got := vertical.Offset(); got != want {
			t.Fatalf("vertical tick %d offset %v, want %v", tick, got, want)
		}
	}
	if got := testing.AllocsPerRun(100, func() { _ = horizontal.Step(); _ = vertical.Step() }); got != 0 {
		t.Fatalf("ribbon clocks allocate %.2f objects", got)
	}
}
