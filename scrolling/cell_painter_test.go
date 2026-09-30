package scrolling

import (
	"errors"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

func testCellBank(t *testing.T) *font.CellBank {
	t.Helper()
	b, err := font.NewCellBank(font.CellBankConfig{Width: 1, Height: 1, Characters: "A", Pixel: func(rune, int, int) bool { return true }})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestOwnedCellModeDoesNotMutateCallerAndClosesResources(t *testing.T) {
	bank := testCellBank(t)
	cells := CellPainterConfig{Fonts: map[string]*font.CellBank{"default": bank}}
	modes := map[string]Mode{"cells": {Cells: &cells}}
	s, err := New(Config{Glyphs: []Glyph{{Rune: 'A', Advance: 2}}, Shape: "cells", Modes: modes})
	if err != nil {
		t.Fatal(err)
	}
	p := s.CellPainterController("cells")
	if p == nil || p.white == nil || modes["cells"].Paint != nil {
		t.Fatal("cell renderer not owned independently")
	}
	p.SetOutlined(true)
	if !p.outlined {
		t.Fatal("outline controller failed")
	}
	s.Close()
	if p.white != nil || s.CellPainterController("cells") != nil {
		t.Fatal("cell resources survived Close")
	}
	s.Close()
}

func TestCellModeFailureClosesCreatedResources(t *testing.T) {
	cells := CellPainterConfig{Fonts: map[string]*font.CellBank{"default": testCellBank(t)}}
	s := &Scrolling{config: Config{Modes: map[string]Mode{}}}
	want := errors.New("invalid prepared glyph data")
	var owned *CellPainter
	s.config.Modes["cells"] = Mode{Cells: &cells, Prepare: func([]Glyph) error { owned = s.CellPainterController("cells"); return want }}
	if _, err := s.finish(); !errors.Is(err, want) {
		t.Fatal("lost prepare error", err)
	}
	if owned == nil || owned.white != nil || len(s.cellPainters) != 0 {
		t.Fatal("failed cell preparation leaked resources")
	}
}

func TestCellPainterRejectsInvalidGeometryAndAmbiguousModes(t *testing.T) {
	bank := testCellBank(t)
	fonts := map[string]*font.CellBank{"default": bank}
	for _, c := range []CellPainterConfig{{}, {Fonts: fonts, Shape: CellShape(2)}, {Fonts: fonts, LineWidth: -1},
		{Fonts: fonts, Shape: CellCuboid}, {Fonts: fonts, Font: "unknown"},
		{Fonts: fonts, Flat: FlatCellConfig{Size: geometry.Vec2{X: -1, Y: 1}}}} {
		if p, err := NewCellPainter(c); err == nil {
			p.Close()
			t.Fatal("accepted invalid cell settings")
		}
	}
	c := CellPainterConfig{Fonts: fonts}
	if _, err := New(Config{Glyphs: []Glyph{{Rune: 'A', Advance: 1}}, Shape: "cells", Modes: map[string]Mode{
		"cells": {Cells: &c, Paint: func(*ebiten.Image, Sample, ebiten.DrawImageOptions) {}},
	}}); err == nil {
		t.Fatal("accepted two painters for one mode")
	}
}
