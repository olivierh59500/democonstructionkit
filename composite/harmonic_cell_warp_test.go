package composite

import (
	"image"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestHarmonicCellWarpCachesRowsAndColumnsIndependently(t *testing.T) {
	source, dst := ebiten.NewImage(73, 37), ebiten.NewImage(200, 160)
	defer source.Deallocate()
	defer dst.Deallocate()
	config := HarmonicCellWarpConfig{
		Cell: image.Pt(16, 9), Source: image.Rect(4, 3, 60, 31),
		Waves: CellWaveBank{
			XRows:    motion.Waves{{Amplitude: 32, Spatial: .3, Speed: .08}},
			YColumns: motion.Waves{{Amplitude: 16, Spatial: .3, Speed: .08}},
		},
	}
	warp, err := NewHarmonicCellWarp(config)
	if err != nil {
		t.Fatal(err)
	}
	defer warp.Close()
	for _, tick := range []uint64{0, 1, 17, 120} {
		warp.DrawAt(dst, source, kit.Frame{Tick: tick}, 10, 20)
		if len(warp.xRows) != 4 || len(warp.yColumns) != 4 {
			t.Fatalf("incorrect clipped grid %d x %d", len(warp.xRows), len(warp.yColumns))
		}
		for row := 0; row < 4; row++ {
			for column := 0; column < 4; column++ {
				pose := warp.sample(row, column, kit.Frame{Tick: tick})
				x, y := pose.GeoM.Apply(0, 0)
				wantX := 32 * math.Sin(float64(tick)*.08+float64(row)*.3)
				wantY := 16 * math.Sin(float64(tick)*.08+float64(column)*.3)
				if math.Abs(x-wantX) > 1e-12 || math.Abs(y-wantY) > 1e-12 {
					t.Fatalf("tick %d cell %d,%d = %v,%v; want %v,%v", tick, row, column, x, y, wantX, wantY)
				}
			}
		}
	}
	rows := &warp.xRows[0]
	warp.DrawAt(dst, source, kit.Frame{Tick: 120}, 10, 20)
	if &warp.xRows[0] != rows {
		t.Fatal("same-frame draw reallocated row offsets")
	}
	bank := CellWaveBank{XColumns: motion.Waves{{Amplitude: 5, Spatial: .1, Speed: .02}}}
	if err := warp.SetWaves(bank); err != nil {
		t.Fatal(err)
	}
	bank.XColumns[0].Amplitude = 500
	warp.DrawAt(dst, source, kit.Frame{Tick: 120}, 10, 20)
	pose := warp.sample(0, 1, kit.Frame{})
	x, _ := pose.GeoM.Apply(0, 0)
	if want := 5 * math.Sin(.1+120*.02); math.Abs(x-want) > 1e-12 {
		t.Fatalf("cue wave = %v, want %v", x, want)
	}
}

func TestHarmonicCellWarpRejectsNonfiniteWaves(t *testing.T) {
	_, err := NewHarmonicCellWarp(HarmonicCellWarpConfig{
		Cell: image.Pt(8, 8), Waves: CellWaveBank{XRows: motion.Waves{{Amplitude: math.NaN()}}},
	})
	if err == nil {
		t.Fatal("accepted NaN wave amplitude")
	}
}
