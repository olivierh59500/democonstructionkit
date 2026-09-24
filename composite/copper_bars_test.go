package composite

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

func TestCopperBarsClocks(t *testing.T) {
	image := ebiten.NewImage(46, 20)
	t.Cleanup(image.Deallocate)
	base := CopperBarsConfig{Image: image, Offsets: make([]int, 1024), Height: 72, Count: 36,
		RowStep: 2, SourceStep: 2, SourcePeriod: 20, BaseX: 60, XShift: 1,
		VelocityA: 3, VelocityB: -5, IndexStepA: 7, IndexStepB: 10, DrawMode: CopperImages}
	masked, err := NewCopperBars(base)
	if err != nil {
		t.Fatal(err)
	}
	if err := masked.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	if a, b := masked.Phases(); a != 3 || b != 1019 {
		t.Fatalf("masked phases = %v, %v", a, b)
	}
	base.Clock = SingleWrapClock
	fractional, err := NewCopperBars(base)
	if err != nil {
		t.Fatal(err)
	}
	if err := fractional.SetSpeed(2); err != nil {
		t.Fatal(err)
	}
	if err := fractional.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	if a, b := fractional.Phases(); a != 6 || b != 1014 {
		t.Fatalf("fractional phases = %v, %v", a, b)
	}
	var updateErr error
	if got := testing.AllocsPerRun(30, func() { updateErr = masked.Update(kit.Frame{}) }); got != 0 || updateErr != nil {
		t.Fatalf("masked update allocations = %v, error = %v", got, updateErr)
	}
}

func TestCopperBarsRejectsInvalidSettings(t *testing.T) {
	image := ebiten.NewImage(46, 20)
	t.Cleanup(image.Deallocate)
	for _, c := range []CopperBarsConfig{
		{}, {Image: image, Offsets: []int{0}, Height: 72, RowStep: 2, SourceStep: 2, SourcePeriod: 20, Clock: SingleWrapClock, DrawMode: CopperImages},
		{Image: image, Offsets: []int{0, 1}, Height: 72, Count: 36, RowStep: 2, SourceStep: 2, SourcePeriod: 21},
		{Image: image, Offsets: []int{0, 1}, Height: 72, Count: 36, RowStep: 2, SourceStep: 2, SourcePeriod: 20, VelocityA: math.NaN()},
	} {
		if _, err := NewCopperBars(c); err == nil {
			t.Fatalf("accepted invalid copper config %+v", c)
		}
	}
	if copperIndex(-1, 1024) != 1023 {
		t.Fatal("negative phase did not wrap")
	}
}
