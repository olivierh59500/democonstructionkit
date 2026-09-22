package composite

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

func TestCellWarpRejectsInvalidDimensions(t *testing.T) {
	for _, config := range []CellWarpConfig{
		{}, {Cell: image.Pt(-1, 8)}, {Cell: image.Pt(8, 0)},
		{Cell: image.Pt(8, 8), Gap: image.Pt(-8, 0)},
		{Cell: image.Pt(8, 8), Gap: image.Pt(0, -9)},
	} {
		if _, err := NewCellWarp(config); err == nil {
			t.Fatalf("accepted invalid cells: %+v", config)
		}
	}
}

func TestCellWarpRetainsPartialCropsAndSubimageOrigin(t *testing.T) {
	atlas := ebiten.NewImage(40, 30)
	defer atlas.Deallocate()
	source := atlas.SubImage(image.Rect(7, 9, 26, 22)).(*ebiten.Image)
	w, err := NewCellWarp(CellWarpConfig{Cell: image.Pt(8, 6), Gap: image.Pt(2, -1)})
	if err != nil {
		t.Fatal(err)
	}
	w.prepare(source)
	if len(w.cells) != 9 {
		t.Fatalf("got %d cells, want 9", len(w.cells))
	}
	for i, c := range w.cells {
		column, row := i%3, i/3
		want := image.Rect(7+column*8, 9+row*6, min(15+column*8, 26), min(15+row*6, 22))
		if c.image.Bounds() != want || c.left != column*10 || c.top != row*5 || c.row != row || c.column != column {
			t.Fatalf("cell %d lost crop, order or gap: %+v, bounds %v, want %v", i, c, c.image.Bounds(), want)
		}
	}
	last := w.cells[8].image.Bounds()
	if last.Size() != image.Pt(3, 1) {
		t.Fatalf("partial edge was padded or dropped: %v", last)
	}
}

func TestCellWarpClipsSourcePaddingAndCachesCells(t *testing.T) {
	source, dst := ebiten.NewImage(30, 20), ebiten.NewImage(40, 30)
	defer source.Deallocate()
	defer dst.Deallocate()
	var visits []image.Point
	frame := kit.Frame{Tick: 17}
	w, err := NewCellWarp(CellWarpConfig{
		Cell: image.Pt(8, 6), Source: image.Rect(3, 4, 14, 13),
		Sample: func(row, column int, f kit.Frame) CellTransform {
			if f != frame {
				t.Fatalf("frame changed: %+v", f)
			}
			visits = append(visits, image.Pt(column, row))
			return CellTransform{Hidden: true}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	w.DrawAt(dst, source, frame, 4, 5)
	first := w.cells[0].image
	if first.Bounds() != image.Rect(3, 4, 11, 10) || w.cells[3].image.Bounds() != image.Rect(11, 10, 14, 13) {
		t.Fatal("source crop padding was not preserved")
	}
	w.DrawAt(dst, source, frame, 4, 5)
	if first != w.cells[0].image || len(visits) != 8 {
		t.Fatal("repeated drawing recreated crops or lost cells")
	}
	for i, point := range visits {
		if point != image.Pt(i%4%2, i%4/2) {
			t.Fatalf("unstable sample order at %d: %v", i, point)
		}
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	w.Close()
	w.DrawAt(dst, source, frame, 4, 5)
	if len(w.cells) != 4 {
		t.Fatal("closed cell cache could not be recreated")
	}
	// The caller still owns the source after closing a warp.
	if source.Bounds() != image.Rect(0, 0, 30, 20) {
		t.Fatal("close modified the borrowed source")
	}
}

func TestCellWarpSwitchesSourceAndHandlesEmptyIntersection(t *testing.T) {
	a, b := ebiten.NewImage(15, 11), ebiten.NewImage(5, 4)
	defer a.Deallocate()
	defer b.Deallocate()
	w, _ := NewCellWarp(CellWarpConfig{Cell: image.Pt(8, 6)})
	w.prepare(a)
	capacity := cap(w.cells)
	w.prepare(b)
	if len(w.cells) != 1 || cap(w.cells) != capacity || w.cells[0].image.Bounds() != b.Bounds() {
		t.Fatal("source change did not reuse storage or crop the new image")
	}
	out, _ := NewCellWarp(CellWarpConfig{Cell: image.Pt(8, 6), Source: image.Rect(100, 100, 120, 120)})
	out.prepare(a)
	if len(out.cells) != 0 {
		t.Fatal("out-of-bounds source crop created fragments")
	}
}
