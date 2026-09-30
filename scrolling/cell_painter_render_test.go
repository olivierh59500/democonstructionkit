//go:build dck_gpu_rendercheck

package scrolling

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestCellPainterGpuMixedFontSizesAndCachedRows(t *testing.T) {
	testutil.RequireGPU(t)
	a, _ := font.NewCellBank(font.CellBankConfig{Width: 2, Height: 2, Characters: "A", Pixel: func(_ rune, x, y int) bool { return x == y }, Color: func(rune, int, int) color.NRGBA { return color.NRGBA{R: 255, A: 255} }})
	b, _ := font.NewCellBank(font.CellBankConfig{Width: 3, Height: 1, Characters: "雪", Pixel: func(rune, int, int) bool { return true }, Color: func(rune, int, int) color.NRGBA { return color.NRGBA{B: 255, A: 255} }})
	calls := 0
	p, err := NewCellPainter(CellPainterConfig{Fonts: map[string]*font.CellBank{"a": a, "b": b}, Font: "a",
		Flat: FlatCellConfig{Rectangles: true}, Rows: func(string, int, float64) geometry.Vec3 { calls++; return geometry.Vec3{} }})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	dst := ebiten.NewImage(32, 24)
	defer dst.Deallocate()
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(5, 6)
	op.GeoM.Translate(3, 4)
	p.Paint(dst, Sample{Glyph: Glyph{Rune: 'A', Font: "a"}}, op)
	p.Paint(dst, Sample{Glyph: Glyph{Rune: 'A', Font: "a"}}, op)
	op = ebiten.DrawImageOptions{}
	op.GeoM.Scale(2, 3)
	op.GeoM.Translate(20, 4)
	p.Paint(dst, Sample{Glyph: Glyph{Rune: '雪', Font: "b"}}, op)
	if calls != 3 {
		t.Fatalf("row geometry recomputed for repeated draw: %d", calls)
	}
	got := make([]byte, 32*24*4)
	dst.ReadPixels(got)
	for y := 0; y < 24; y++ {
		for x := 0; x < 32; x++ {
			want := color.NRGBA{}
			if x >= 3 && x < 8 && y >= 4 && y < 10 || x >= 8 && x < 13 && y >= 10 && y < 16 {
				want = color.NRGBA{R: 255, A: 255}
			}
			if x >= 20 && x < 26 && y >= 4 && y < 7 {
				want = color.NRGBA{B: 255, A: 255}
			}
			at := (y*32 + x) * 4
			if got[at] != want.R || got[at+1] != want.G || got[at+2] != want.B || got[at+3] != want.A {
				t.Fatalf("mixed cell fonts at %d,%d: %v != %v", x, y, got[at:at+4], want)
			}
		}
	}
	p.InvalidateRows()
	p.Paint(dst, Sample{Glyph: Glyph{Rune: '雪', Font: "b"}}, op)
	if calls != 4 {
		t.Fatal("live row parameters did not invalidate the cache")
	}
}

func TestCellPainterGpuProjectedCuboidAndOutline(t *testing.T) {
	testutil.RequireGPU(t)
	colors := make([]color.NRGBA, 8)
	for i := range colors {
		colors[i] = color.NRGBA{R: 255, A: 255}
	}
	p, err := NewCellPainter(CellPainterConfig{Fonts: map[string]*font.CellBank{"default": testCellBank(t)}, Shape: CellCuboid,
		Cuboid: CuboidCellConfig{Size: geometry.Vec3{X: 16, Y: 16, Z: 16}, Origin: geometry.Vec3{Z: 64}, Colors: colors,
			Camera: geometry.Camera{Center: geometry.Vec2{X: 32, Y: 32}, Focal: 64, Near: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	dst := ebiten.NewImage(64, 64)
	defer dst.Deallocate()
	sample := Sample{Glyph: Glyph{Rune: 'A'}}
	p.Paint(dst, sample, ebiten.DrawImageOptions{})
	got := make([]byte, 64*64*4)
	dst.ReadPixels(got)
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			want := x >= 23 && x < 41 && y >= 23 && y < 41
			at := (y*64 + x) * 4
			if (got[at+3] != 0) != want {
				t.Fatalf("cuboid projection at %d,%d differs from the front-plane rectangle", x, y)
			}
		}
	}
	dst.Clear()
	p.SetOutlined(true)
	p.Paint(dst, sample, ebiten.DrawImageOptions{})
	dst.ReadPixels(got)
	if got[(32*64+32)*4+3] != 0 {
		t.Fatal("outlined cuboid filled its centre")
	}
	visible := 0
	for i := 3; i < len(got); i += 4 {
		if got[i] != 0 {
			visible++
		}
	}
	if visible == 0 {
		t.Fatal("outlined cuboid has no edges")
	}
}
