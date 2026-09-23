package effects

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

func TestPerspectiveCheckerboardClockAndCustomPosition(t *testing.T) {
	base := DefaultPerspectiveCheckerboardConfig()
	before, err := NewPerspectiveCheckerboard(base)
	if err != nil {
		t.Fatal(err)
	}
	defer before.Close()
	if err := before.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	x, y := before.Offset()
	if math.Abs(x-30.72) > 1e-9 || math.Abs(y-53.92) > 1e-9 {
		t.Fatalf("oscillator-before-move offset = (%g, %g)", x, y)
	}
	base.ClockOrder = CheckerboardAfterMove
	base.Wrap = CheckerboardSingleWrap
	after, err := NewPerspectiveCheckerboard(base)
	if err != nil {
		t.Fatal(err)
	}
	defer after.Close()
	if err := after.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	x, y = after.Offset()
	if x != 0 || math.Abs(y-10.08) > 1e-9 {
		t.Fatalf("oscillator-after-move offset = (%g, %g)", x, y)
	}
	base.Position = func(tick uint64) (float64, float64) { return float64(tick) * 12, float64(tick) * 3 }
	custom, err := NewPerspectiveCheckerboard(base)
	if err != nil {
		t.Fatal(err)
	}
	defer custom.Close()
	for range 2 {
		if err := custom.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
	}
	canvas := ebiten.NewImage(384, 270)
	custom.Draw(canvas)
	custom.Draw(canvas)
	x, y = custom.Offset()
	if x != 24 || y != 6 {
		t.Fatalf("custom path changed during Draw: (%g, %g)", x, y)
	}
}
