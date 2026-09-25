package scrolling

import (
	"image"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	bitmap "github.com/olivierh59500/democonstructionkit/font"
)

func TestPseudo3DBanksKeepIndependentPixelTransport(t *testing.T) {
	atlas := ebiten.NewImage(8, 8)
	defer atlas.Deallocate()
	metrics, err := bitmap.NewGrid(bitmap.Grid{Bounds: atlas.Bounds(), Cell: image.Pt(8, 8), Columns: 1, Order: "A"})
	if err != nil {
		t.Fatal(err)
	}
	config := Pseudo3DConfig{
		Banks:   []Pseudo3DBank{{Text: "AAAA", BaseY: 50}, {Text: "AA", BaseY: 30, InvertScale: true}},
		Face:    Face{Atlas: atlas, Metrics: metrics},
		Advance: 8, PixelsPerUpdate: 2, Visible: 2, TicksPerSecond: 60,
		Width: 64, Height: 64,
		Pose: Pseudo3DPose{ZRate: 1, ZIndexStep: 1, XTimeRate: 1, XIndexRate: 1,
			YRate: 1, YIndexStep: 1, ZBase: 1.5, ZAmplitude: .5, ScaleSum: 3,
			OutputScaleX: 2, CullMargin: 10, Alpha: 1},
	}
	scroll, err := New(Config{Pseudo3D: &config})
	if err != nil {
		t.Fatal(err)
	}
	defer scroll.Close()
	bank := scroll.Pseudo3DController()
	if bank == nil {
		t.Fatal("pseudo-3d controller is unavailable through scrolling.New")
	}
	positions := [2]float64{}
	for tick := 0; tick < 1000; tick++ {
		if tick == 500 {
			if err := bank.SetBankSpeed(1, 0); err != nil {
				t.Fatal(err)
			}
		}
		if tick == 750 {
			if err := scroll.SetTransportMultiplier(.5); err != nil {
				t.Fatal(err)
			}
		}
		if err := scroll.Update(kit.Frame{Tick: uint64(tick + 1)}); err != nil {
			t.Fatal(err)
		}
		for index, limit := range [...]float64{32, 16} {
			if index == 0 || tick < 500 {
				step := 2.0
				if tick >= 750 {
					step = 1
				}
				positions[index] += step
				if positions[index] >= limit {
					positions[index] -= limit
				}
			}
			if bank.BankPosition(index) != positions[index] {
				t.Fatalf("tick %d bank %d at %v, want %v", tick, index, bank.BankPosition(index), positions[index])
			}
		}
	}
	if err := bank.SetBankPosition(0, 13); err != nil || bank.BankPosition(0) != 13 {
		t.Fatal("per-bank seek did not retain its supplied position")
	}
	if err := bank.SetTimeOffset(2); err != nil {
		t.Fatal(err)
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = scroll.Update(kit.Frame{}) }); allocations != 0 {
		t.Fatalf("pseudo-3d update allocates %.1f objects", allocations)
	}
}

func TestPseudo3DReverseHarmonicRecurrenceMatchesDirectTrig(t *testing.T) {
	const start = 12.345
	sinValue, cosValue := math.Sincos(start)
	stepSin, stepCos := math.Sincos(.7)
	for index := 1; index <= 100; index++ {
		sinValue, cosValue = pseudoSinCosBackward(sinValue, cosValue, stepSin, stepCos)
		wantSin, wantCos := math.Sincos(start - float64(index)*.7)
		if math.Abs(sinValue-wantSin) > 1e-12 || math.Abs(cosValue-wantCos) > 1e-12 {
			t.Fatalf("reverse harmonic drift at glyph %d", index)
		}
	}
}
