package composite

import (
	"image"
	"math"
	"testing"
)

func TestStripWarpRejectsInvalidGeometry(t *testing.T) {
	valid := StripWarpConfig{RowHeight: 2, ColumnWidth: 16}
	for _, size := range []image.Point{{}, {X: -1, Y: 171}, {X: 800, Y: 0}} {
		if _, err := NewStripWarp(size, valid); err == nil {
			t.Fatalf("accepted size %v", size)
		}
	}
	for _, config := range []StripWarpConfig{
		{RowHeight: 0, ColumnWidth: 16},
		{RowHeight: 2, ColumnWidth: -1},
		{RowHeight: 2, ColumnWidth: 16, VerticalBias: math.NaN()},
	} {
		if _, err := NewStripWarp(image.Pt(800, 171), config); err == nil {
			t.Fatalf("accepted config %+v", config)
		}
	}
}

func TestStripWarpOwnsIndependentSurfaces(t *testing.T) {
	config := StripWarpConfig{RowHeight: 2, ColumnWidth: 16}
	text, err := NewStripWarp(image.Pt(800, 64), config)
	if err != nil {
		t.Fatal(err)
	}
	logo, err := NewStripWarp(image.Pt(800, 171), config)
	if err != nil {
		t.Fatal(err)
	}
	defer logo.Close()
	if text.surface == logo.surface || logo.Size() != image.Pt(800, 171) {
		t.Fatal("shared configuration lost independent content dimensions")
	}
	if err = text.Close(); err != nil {
		t.Fatal(err)
	}
	if err = text.Close(); err != nil {
		t.Fatal(err)
	}
	if logo.surface == nil {
		t.Fatal("closing text released logo resources")
	}
}
