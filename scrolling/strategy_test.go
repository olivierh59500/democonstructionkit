package scrolling

import (
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
)

func TestPageLayoutKeepsLinesHorizontalWithMixedFonts(t *testing.T) {
	s, err := New(Config{Text: "AA\n{font:wide}A\n\n{font:default}A", Fonts: map[string]Face{"default": face(t, 10), "wide": face(t, 20)}, Controls: scrolltext.Braces,
		Page: &PageConfig{Width: 60, LineHeight: 12, Align: AlignCenter}, Speed: 6, Y: 100, Repeat: true, Gap: 4})
	if err != nil {
		t.Fatal(err)
	}
	if !s.config.Vertical || s.Length() != 48 || len(s.glyphs) != 4 {
		t.Fatal(s.length, s.glyphs)
	}
	want := [][2]float64{{20, 0}, {30, 0}, {20, 12}, {25, 36}}
	for i, g := range s.glyphs {
		if g.X != want[i][0] || g.Offset != want[i][1] {
			t.Fatal(i, g)
		}
	}
	state := s.StateAt(2)
	if state.Y != 88 {
		t.Fatal(state)
	}
}

func TestCommonRecycledTransportRetainsCommandAndWaveTiming(t *testing.T) {
	a := ebiten.NewImage(64, 8)
	defer a.Deallocate()
	cfg := RingConfig{Text: "AB^S2AB^P1AB", Font: BitmapGrid{Image: a, Width: 8, Height: 8, Columns: 8, First: 'A'}, Viewport: 32, Speed: 3, Controls: true, Waves: []RingWave{{Amplitude: 3, LetterStep: .2, TickStep: .07}}}
	original, err := NewRing(cfg)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(Config{Recycled: &RecycledConfig{Ring: cfg}})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for tick := 0; tick < 900; tick++ {
		original.Step()
		if err := s.Update(kit.Frame{Tick: uint64(tick)}); err != nil {
			t.Fatal(err)
		}
		got := s.backend.(*recycledTransport).ring
		if !reflect.DeepEqual(original.Letters(), got.Letters()) || original.Cursor() != got.Cursor() || original.speed != got.speed || original.pause != got.pause {
			t.Fatalf("recycled transport diverged at %d", tick)
		}
	}
}

func TestCommonProjectedTransportRetainsVisibleSlotTiming(t *testing.T) {
	face := face(t, 10)
	cfg := PlanesConfig{Slots: []PlaneSlot{{Rune: 'A', Advance: 10, Form: 0}, {Rune: 'A', Advance: 20, Form: 1}, {Rune: 'A', Advance: 10, Form: -1}}, Forms: []PlaneForm{{Height: 20, VerticalStep: .2, VerticalSpeed: 2}, {DepthAmplitude: 40, DepthStep: .3, Height: 7}}, Projection: PlaneProjection{Focal: 250, Depth: 100}, Visible: 9, PhaseStep: .03}
	original, err := NewPlanes(cfg)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(Config{Projected: &ProjectedConfig{Planes: cfg, Face: face, Draw: PlaneDraw{ScaleX: 1, ScaleY: 1}, PixelsPerUpdate: 2}})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for tick := 0; tick < 300; tick++ {
		original.Step(2)
		s.Update(kit.Frame{Tick: uint64(tick)})
		if !reflect.DeepEqual(original.Points(), s.backend.(*projectedTransport).planes.Points()) {
			t.Fatalf("projected transport diverged at %d", tick)
		}
	}
}

func TestOutputPassOrderAndFeedbackAdvanceOncePerUpdate(t *testing.T) {
	var calls []int
	glyphSamples := 0
	s, err := New(Config{Text: "AA", Fonts: map[string]Face{"default": face(t, 4)}, Speed: 1,
		Map: func(Sample, *ebiten.DrawImageOptions) bool { glyphSamples++; return true },
		Output: &OutputConfig{Width: 20, Height: 10, Feedback: []FeedbackLayer{{Config: FeedbackDNAConfig{Width: 20, Height: 10, HorizontalSpeed: 1, ColumnWidth: 1, Direction: 1, Profile: []int{0, 1, 2}}}}, Passes: []kit.ImagePass{
			{Apply: func(dst, src *ebiten.Image, _ kit.Frame) { calls = append(calls, 1); dst.DrawImage(src, nil) }},
			{Apply: func(dst, src *ebiten.Image, _ kit.Frame) { calls = append(calls, 2); dst.DrawImage(src, nil) }},
		}}})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	dst := ebiten.NewImage(20, 10)
	defer dst.Deallocate()
	if err := s.Update(kit.Frame{Tick: 1, Time: .1}); err != nil {
		t.Fatal(err)
	}
	if glyphSamples != 2 {
		t.Fatal("feedback did not sample its source once", glyphSamples)
	}
	s.Draw(dst)
	s.Draw(dst)
	if glyphSamples != 2 {
		t.Fatal("drawing feedback advanced its source", glyphSamples)
	}
	if !reflect.DeepEqual(calls, []int{1, 2, 1, 2}) {
		t.Fatal(calls)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewRejectsAmbiguousTransportAndBadPage(t *testing.T) {
	for _, c := range []Config{{Page: &PageConfig{Width: -1}}, {Recycled: &RecycledConfig{}, Text: "ignored"}, {Recycled: &RecycledConfig{}, Projected: &ProjectedConfig{}}, {Glyphs: []Glyph{{Advance: 1}}, Page: &PageConfig{}}, {Glyphs: []Glyph{{Advance: 1}}, Output: &OutputConfig{Width: -1}}} {
		if _, err := New(c); err == nil {
			t.Fatal("invalid configuration accepted", c)
		}
	}
}

func TestCommonSlicedTransportKeepsRotationIndependentFromText(t *testing.T) {
	img := ebiten.NewImage(7, 8)
	defer img.Deallocate()
	film, err := NewDNAFrames([]*ebiten.Image{img}, DNAFrameConfig{Frames: 30})
	if err != nil {
		t.Fatal(err)
	}
	defer film.Close()
	cfg := SliceStreamConfig{Tokens: []SliceToken{{Glyph: 0, Width: 7}, {Control: "pause"}, {Glyph: 0, Width: 4}}, Capacity: 12, SliceWidth: 1, Repeat: true, LoopStart: 0, Initial: DNASlice{Glyph: -1}}
	direct, err := NewSliceStream(cfg)
	if err != nil {
		t.Fatal(err)
	}
	controls := 0
	s, err := New(Config{Sliced: &SlicedConfig{Stream: cfg, Film: film, Draw: DNADrawConfig{SliceWidth: 1, ScaleX: 1, ScaleY: 1}, RotationSpeed: 5,
		AdvanceAt: func(f kit.Frame) int {
			if f.Tick%5 == 0 {
				return 0
			}
			return 2
		}, OnControl: func(SliceControl) bool { controls++; return true }}})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for tick := uint64(0); tick < 100; tick++ {
		count := 2
		if tick%5 == 0 {
			count = 0
		}
		direct.Step(count, func(SliceControl) bool { return true })
		time := float64(tick) / 60
		direct.SetFrames(time*5, nil, film.Count)
		s.Update(kit.Frame{Tick: tick, Time: time})
		got := s.backend.(*slicedTransport).stream
		if got.Head() != direct.Head() || !reflect.DeepEqual(got.Slices(), direct.Slices()) {
			t.Fatal("sliced transport diverged", tick)
		}
	}
	if controls == 0 {
		t.Fatal("control event was lost")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if film.Image == nil {
		t.Fatal("scrolling closed caller-owned filmstrip")
	}
}
