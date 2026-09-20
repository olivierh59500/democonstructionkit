package scrolling

import (
	"image"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
)

func TestModeControlsAndSequenceWithMixedFonts(t *testing.T) {
	sequence, err := NewModeSequence([]Cue{{0, "normal"}, {2, "bounce"}, {4, "sine"}, {6, "zoom"}, {8, "3d"}, {10, "dna"}}, 12)
	if err != nil {
		t.Fatal(err)
	}
	dna, err := NewDNA(DNAConfig{RotationSpeed: 10, Twist: motion.Wave{Amplitude: 15, Spatial: .03}})
	if err != nil {
		t.Fatal(err)
	}
	defer dna.Close()
	perspective, err := Perspective(PerspectiveConfig{Focal: 250, Near: 1, Depth: 100, DepthWave: motion.Wave{Amplitude: 90, Spatial: .1}})
	if err != nil {
		t.Fatal(err)
	}
	modes := map[string]Mode{"normal": Normal(), "bounce": Bounce(motion.Wave{Amplitude: 10, Speed: 2}), "sine": Sine(motion.Wave{Amplitude: 20, Spatial: .1, Speed: 2}), "zoom": Zoom(ZoomConfig{BaseX: 1, BaseY: 1, Wave: motion.Wave{Amplitude: .3, Speed: 2}}), "3d": perspective, "dna": dna.Mode()}
	c := Config{Text: "A{font:wide}{shape:dna}A", Fonts: map[string]Face{"default": face(t, 7), "wide": face(t, 19)}, Controls: scrolltext.Braces, Speed: 7, Modes: modes}
	s, err := New(c)
	if err != nil {
		t.Fatal(err)
	}
	if s.StateAt(.5).Shape != "" || s.StateAt(1).Shape != "dna" {
		t.Fatal("shape control timing changed")
	}
	if len(dna.cache) != 2 {
		t.Fatal("DNA did not prepare both independent faces")
	}
	for _, f := range dna.cache {
		if f.Image.Bounds().Dx() != f.Entries[0].Width*30 {
			t.Fatal("font width was hard-coded")
		}
	}
	c.Sequence = sequence
	s, err = New(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		time float64
		name string
	}{{0, "normal"}, {2, "bounce"}, {6, "zoom"}, {10, "dna"}, {1208, "3d"}} {
		if got := s.StateAt(tc.time).Shape; got != tc.name {
			t.Fatalf("mode at %g: %q", tc.time, got)
		}
	}
}

func TestPerspectiveDepthAndGlyphTransforms(t *testing.T) {
	p, err := Perspective(PerspectiveConfig{Focal: 100, Near: 1, Depth: 50, DepthWave: motion.Wave{Amplitude: 40, Spatial: math.Pi / 20}})
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(Config{Text: "AAA", Fonts: map[string]Face{"default": face(t, 10)}, Modes: map[string]Mode{"3d": p}})
	if err != nil {
		t.Fatal(err)
	}
	var order []int
	state := IdentityState()
	state.Shape = "3d"
	state.Paint = func(_ *ebiten.Image, s Sample, op ebiten.DrawImageOptions) {
		order = append(order, s.Index)
		x0, _ := op.GeoM.Apply(0, 0)
		x1, _ := op.GeoM.Apply(10, 0)
		want := 10 * 100 / (100 + 50 + p.Depth(s) - 50)
		if math.Abs((x1-x0)-want) > 1e-10 {
			t.Fatal("wrong scale for glyph", s.Index)
		}
	}
	s.DrawAt(nil, state)
	if len(order) != 3 || order[0] != 1 {
		t.Fatal("incorrect depth order", order)
	}
}

func TestDNAFilmstripMetricsAndPhase(t *testing.T) {
	a, b := ebiten.NewImage(11, 26), ebiten.NewImage(7, 8)
	defer a.Deallocate()
	defer b.Deallocate()
	f, err := NewDNAFrames([]*ebiten.Image{a, b}, DNAFrameConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if f.Image.Bounds() != image.Rect(0, 0, 330, 44) || f.Entries[1] != (DNAEntry{Y: 33, Width: 7, Height: 11}) {
		t.Fatal("incorrect proportional filmstrip", f.Entries, f.Image.Bounds())
	}
	if f.FrameAt(-61, 0) != 29 || f.FrameAt(65, 2) != 7 {
		t.Fatal("negative or large phase did not wrap")
	}
	if _, err := NewDNAFrames([]*ebiten.Image{a}, DNAFrameConfig{Height: 20}); err == nil {
		t.Fatal("cropping font silently accepted")
	}
}

func TestInvalidSequenceAndUnknownMode(t *testing.T) {
	if _, err := NewModeSequence([]Cue{{1, "normal"}}, 0); err == nil {
		t.Fatal("missing initial cue accepted")
	}
	seq, _ := NewModeSequence([]Cue{{0, "missing"}}, 0)
	if _, err := New(Config{Glyphs: []Glyph{{Advance: 1}}, Sequence: seq}); err == nil {
		t.Fatal("unknown mode accepted")
	}
}
