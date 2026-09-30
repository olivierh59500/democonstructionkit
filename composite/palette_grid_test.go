package composite

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestPaletteGridValidationAndOwnedMapping(t *testing.T) {
	for _, c := range []PaletteGridConfig{
		{}, {Width: 8, Height: 8, Columns: 2049, Rows: 1},
		{Width: 8, Height: 8, Columns: 2, Rows: 2, ControlThreshold: float32(math.NaN())},
		{Width: 8, Height: 8, Columns: 2, Rows: 2, X: PaletteGridAxis{CellSize: -1}},
		{Width: 8, Height: 8, Columns: 2, Rows: 2, Y: PaletteGridAxis{Indices: []int{0, 2}}},
	} {
		if p, err := NewPaletteGrid(c); err == nil {
			p.Close()
			t.Fatalf("accepted invalid grid %+v", c)
		}
	}
	indices := []int{0, 2, 1}
	p, err := NewPaletteGrid(PaletteGridConfig{Width: 8, Height: 8, Columns: 2, Rows: 3, Y: PaletteGridAxis{Indices: indices}})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	indices[1] = 0
	if p.options.Uniforms["YMap"].([]float32)[1] != 2 || p.palette.Bounds().Dx() != 4 || p.palette.Bounds().Dy() != 3 || len(p.pixels) != 48 {
		t.Fatal("mapping was not copied or palette storage is larger than its cells")
	}
	if err := p.SetColors(make([]color.NRGBA, 5), make([]color.NRGBA, 5)); err == nil {
		t.Fatal("accepted incomplete color banks")
	}
	if err := p.SetBodyColor(color.NRGBA{}); err == nil {
		t.Fatal("enabled an unconfigured body material")
	}
	source, dst := ebiten.NewImage(8, 8), ebiten.NewImage(8, 8)
	defer source.Deallocate()
	defer dst.Deallocate()
	if err := p.Draw(dst, source, nil, PaletteGridState{ControlOffset: [2]float32{float32(math.Inf(1))}}); err == nil {
		t.Fatal("accepted a nonfinite source offset")
	}
	p.Close()
	p.Close()
	if err := p.SetColors(nil, nil); err == nil {
		t.Fatal("updated closed storage")
	}
}
