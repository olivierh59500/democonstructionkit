package scrolling

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
	"image"
	"math"
	"testing"
)

func face(t *testing.T, width int) Face {
	t.Helper()
	atlas := ebiten.NewImage(width, 8)
	t.Cleanup(atlas.Deallocate)
	metrics, err := font.NewGrid(font.Grid{Bounds: atlas.Bounds(), Cell: image.Pt(width, 8), Columns: 1, Order: "A"})
	if err != nil {
		t.Fatal(err)
	}
	return Face{Atlas: atlas, Metrics: metrics}
}
func TestMixedFontsControlsAndLargeTimeJumps(t *testing.T) {
	s, err := New(Config{Text: "A{speed:20}{pause:2}{font:wide}{shape:wave}A{effect:gold}A", Fonts: map[string]Face{"default": face(t, 10), "wide": face(t, 20)}, Controls: scrolltext.Braces, Speed: 10, Repeat: true, Gap: 10, Shapes: map[string]Mapper{"wave": func(Sample, *ebiten.DrawImageOptions) bool { return true }}, Effects: map[string]Mapper{"gold": func(Sample, *ebiten.DrawImageOptions) bool { return true }}})
	if err != nil {
		t.Fatal(err)
	}
	if s.GlyphCount() != 3 || s.Length() != 50 || s.glyphs[1].Font != "wide" || s.glyphs[2].Effect != "gold" {
		t.Fatal(s.glyphs)
	}
	for _, tt := range []struct {
		time, position float64
		shape          string
	}{{.5, 5, ""}, {1, 10, ""}, {2.9, 10, ""}, {3, 10, "wave"}, {4, 30, "wave"}, {5.5, 0, ""}, {5504, 30, "wave"}} {
		got := s.StateAt(tt.time)
		if math.Abs(got.Position-tt.position) > 1e-7 || got.Shape != tt.shape {
			t.Fatal(tt, got)
		}
	}
}
func TestFontScaleAndLiteralControlOptOut(t *testing.T) {
	s, err := New(Config{Text: "A{scale:2,3}A", Fonts: map[string]Face{"default": face(t, 10)}, Controls: scrolltext.Braces, Speed: 10})
	if err != nil {
		t.Fatal(err)
	}
	if s.Length() != 30 || s.glyphs[1].ScaleY != 3 {
		t.Fatal(s.glyphs)
	}
	if _, err = New(Config{Text: "{font:missing}A", Fonts: map[string]Face{"default": face(t, 10)}, Controls: scrolltext.Braces}); err == nil {
		t.Fatal("unknown face accepted")
	}
}
