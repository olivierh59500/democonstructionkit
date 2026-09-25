package presets

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestColorshockOrbitAndTableClockPreserveAuthoredCadence(t *testing.T) {
	orbit, err := motion.NewFormulaFormation(CuddlyColorshockOrbit())
	if err != nil {
		t.Fatal(err)
	}
	clock, err := motion.NewWrapBank(CuddlyColorshockTableClock())
	if err != nil {
		t.Fatal(err)
	}
	phase, position := 0.0, 0
	for tick := 0; tick < 3000; tick++ {
		point := orbit.At(phase, 0, 0, 0, 1)
		wantX := 400 - math.Sin(phase*math.Pi/100)*400
		wantY := 180 - math.Cos(phase*math.Pi/200)*300
		if math.Abs(point.X-wantX) > 1e-11 || math.Abs(point.Y-wantY) > 1e-11 || clock.At(0) != float64(position) {
			t.Fatalf("tick %d: orbit %+v, table %v; want %v,%v and %d", tick, point, clock.At(0), wantX, wantY, position)
		}
		phase += .8
		position += 2
		if position > 1872 {
			position = 0
		}
		clock.Step()
	}
}
