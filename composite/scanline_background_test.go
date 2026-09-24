package composite

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

func TestScanlineBackgroundClockAndLifetime(t *testing.T) {
	tile := ebiten.NewImage(8, 64)
	t.Cleanup(tile.Deallocate)
	program, err := NewDisplacementProgram(nil, []int{0, 1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	background, err := NewScanlineBackground(ScanlineBackgroundConfig{
		Tile: tile, Program: program, Width: 32, Height: 12,
		WaveDivisor: 2, WaveStep: 5, BounceAmplitude: 30, BounceRate: .1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := background.Update(kit.Frame{Tick: 3}); err != nil {
		t.Fatal(err)
	}
	wantRows := make([]int, 12)
	program.Fill(wantRows, 15)
	for row, want := range wantRows {
		if background.rows[row] != want {
			t.Fatalf("row %d = %d, want %d", row, background.rows[row], want)
		}
	}
	if want := int(30 * math.Abs(math.Sin(.3))); background.bounce != want {
		t.Fatalf("bounce = %d, want %d", background.bounce, want)
	}
	var updateErr error
	if allocations := testing.AllocsPerRun(30, func() {
		updateErr = background.Update(kit.Frame{Tick: 3})
	}); allocations != 0 || updateErr != nil {
		t.Fatalf("update allocations = %v, error = %v", allocations, updateErr)
	}
	if err := background.Close(); err != nil {
		t.Fatal(err)
	}
	if err := background.Update(kit.Frame{Tick: 4}); err == nil {
		t.Fatal("updated a closed background")
	}
}

func TestScanlineBackgroundValidation(t *testing.T) {
	tile := ebiten.NewImage(8, 64)
	t.Cleanup(tile.Deallocate)
	program, err := NewDisplacementProgram(nil, []int{0})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []ScanlineBackgroundConfig{
		{}, {Tile: tile, Width: 32, Height: 12},
		{Tile: tile, Program: program, Width: 32, Height: 12, SurfaceWidth: 32},
		{Tile: tile, Program: program, Width: 32, Height: 12, SurfaceHeight: 65},
		{Tile: tile, Program: program, Width: 32, Height: 12, BounceAmplitude: math.Inf(1)},
	} {
		if _, err := NewScanlineBackground(c); err == nil {
			t.Fatalf("accepted invalid scanline background: %+v", c)
		}
	}
	if got := wrapBackground(-1, 8); got != 7 {
		t.Fatalf("negative source wrap = %d, want 7", got)
	}
}
