package scrolling

import (
	"errors"
	"image"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
)

func TestAutomaticScrollBoundsWorkForTinyAdvances(t *testing.T) {
	for _, advance := range []float64{1e-9, 1e-100} {
		s, err := New(Config{Glyphs: []Glyph{{Advance: advance}}, Repeat: true, Speed: 1})
		if err != nil {
			t.Fatal(err)
		}
		bounds := image.Rect(0, 0, 640, 360)
		if !errors.Is(s.ValidateRenderBounds(bounds), ErrDrawBudget) {
			t.Fatal("dense layout passed preflight")
		}
		state := s.repeatState(s.StateAt(0), bounds)
		if state.First != state.End || !errors.Is(s.Err(), ErrDrawBudget) {
			t.Fatal("unbounded automatic ribbon", state, s.Err())
		}
	}
}

type repeatObservation struct {
	sample  Sample
	options ebiten.DrawImageOptions
}

// Capture the public drawing pipeline without submitting glyphs to the GPU.
func repeatCapture(t *testing.T, config Config, width, height int) (*Scrolling, func(float64) []repeatObservation) {
	t.Helper()
	dst := ebiten.NewImage(width, height)
	t.Cleanup(dst.Deallocate)
	var observations []repeatObservation
	mapper := config.Map
	config.Map = func(sample Sample, options *ebiten.DrawImageOptions) bool {
		if mapper != nil && !mapper(sample, options) {
			return false
		}
		observations = append(observations, repeatObservation{sample, *options})
		return false
	}
	s, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	return s, func(seconds float64) []repeatObservation {
		observations = nil
		if err := s.Update(kit.Frame{Time: seconds}); err != nil {
			t.Fatal(err)
		}
		s.Draw(dst)
		return observations
	}
}

// The reference uses absolute glyph indices, independently of the renderer's
// viewport selection. Only glyphs intersecting the requested axis are compared.
func assertRepeatRibbon(t *testing.T, s *Scrolling, got []repeatObservation, origin, extent, position float64, vertical bool) {
	t.Helper()
	period := s.Length() + s.config.Gap
	visible := map[int]Sample{}
	for _, observation := range got {
		sample := observation.sample
		if sample.Index < 0 {
			t.Fatalf("initial entry acquired a negative virtual glyph: %+v", sample)
		}
		if math.Abs(sample.Position-position) > 1e-7 {
			t.Fatalf("glyph %d: travel = %g, want %g", sample.Index, sample.Position, position)
		}
		axis := sample.X
		if vertical {
			axis = sample.Y
		}
		if axis+sample.Glyph.Advance > 0 && axis < extent {
			if _, duplicate := visible[sample.Index]; duplicate {
				t.Fatalf("duplicate visible virtual glyph %d", sample.Index)
			}
			visible[sample.Index] = sample
		}
	}
	firstCycle := max(0, int(math.Floor((position-origin-period)/period)))
	lastCycle := max(0, int(math.Ceil((position-origin+extent)/period)))
	for cycle := firstCycle; cycle <= lastCycle; cycle++ {
		for i, glyph := range s.glyphs {
			offset := float64(cycle)*period + glyph.Offset
			axis := origin + offset - position
			if axis+glyph.Advance <= 0 || axis >= extent {
				continue
			}
			index := cycle*len(s.glyphs) + i
			sample, ok := visible[index]
			if !ok {
				t.Fatalf("missing visible virtual glyph %d at %g (travel %g), got %+v", index, axis, position, visible)
			}
			actual := sample.X
			if vertical {
				actual = sample.Y
			}
			if math.Abs(actual-axis) > 1e-7 || math.Abs(sample.Glyph.Offset-offset) > 1e-7 {
				t.Fatalf("glyph %d: axis/offset = %g/%g, want %g/%g", index, actual, sample.Glyph.Offset, axis, offset)
			}
			delete(visible, index)
		}
	}
	if len(visible) != 0 {
		t.Fatalf("unexpected visible glyphs: %+v", visible)
	}
}

func TestRepeatDrawKeepsTheTailAcrossCycleBoundaries(t *testing.T) {
	for _, gap := range []float64{0, 20} {
		t.Run(map[float64]string{0: "adjacent", 20: "gap"}[gap], func(t *testing.T) {
			s, draw := repeatCapture(t, Config{
				Glyphs: []Glyph{{Advance: 10}, {Advance: 10}, {Advance: 10}},
				X:      100, Speed: 10, Gap: gap, Repeat: true,
			}, 100, 20)
			seam := (30 + gap) / 10
			for _, seconds := range []float64{0, .01, seam - .01, seam, seam + .01, 20.01, 1_000_000.01} {
				got := draw(seconds)
				assertRepeatRibbon(t, s, got, 100, 100, seconds*10, false)
				if len(got) > 100 {
					t.Fatalf("drawing work grows with elapsed time: %d callbacks at %g seconds", len(got), seconds)
				}
				for _, observation := range got {
					x, y := observation.options.GeoM.Apply(0, 0)
					if math.Abs(x-observation.sample.X) > 1e-7 || math.Abs(y-observation.sample.Y) > 1e-7 {
						t.Fatal("draw options lost the sample's continuous position")
					}
				}
			}
		})
	}
}

func TestRepeatDrawPreservesGlyphEffectPhaseAcrossWrap(t *testing.T) {
	_, draw := repeatCapture(t, Config{
		Glyphs: []Glyph{{Advance: 10}, {Advance: 10}, {Advance: 10}},
		X:      100, Speed: 10, Repeat: true,
		Map: func(sample Sample, options *ebiten.DrawImageOptions) bool {
			options.GeoM.Translate(0, math.Sin(sample.Glyph.Offset*.07+float64(sample.Index)*.3)*12)
			return true
		},
	}, 100, 40)
	before := draw(2.99)
	after := draw(3.01)
	for index := 0; index < 3; index++ {
		var a, b *repeatObservation
		for i := range before {
			if before[i].sample.Index == index {
				a = &before[i]
			}
		}
		for i := range after {
			if after[i].sample.Index == index {
				b = &after[i]
			}
		}
		if a == nil || b == nil {
			t.Fatalf("glyph %d disappeared at the cycle seam", index)
		}
		ax, ay := a.options.GeoM.Apply(0, 0)
		bx, by := b.options.GeoM.Apply(0, 0)
		if math.Abs(bx-ax+.2) > 1e-7 || ay != by || a.sample.Glyph.Offset != b.sample.Glyph.Offset {
			t.Fatalf("glyph %d jumped or changed deformation phase across the seam", index)
		}
	}
}

func TestRepeatDrawKeepsFractionalDurationSeamsContinuous(t *testing.T) {
	s, draw := repeatCapture(t, Config{
		Glyphs: []Glyph{{Advance: 1}}, X: 10, Speed: 10, Repeat: true,
	}, 10, 10)
	// The quotient can round to five while Mod(.5, .1) remains almost .1.
	for _, seconds := range []float64{.5 - 1e-8, .5, .5 + 1e-8} {
		assertRepeatRibbon(t, s, draw(seconds), 10, 10, seconds*10, false)
	}
}

func TestRepeatDrawHonorsExplicitBoundsBeforeProjection(t *testing.T) {
	s, draw := repeatCapture(t, Config{
		Glyphs: []Glyph{{Advance: 10}}, X: 1000, Speed: 10, Repeat: true,
		RepeatBounds: image.Rect(0, 0, 1000, 20),
		Map: func(_ Sample, options *ebiten.DrawImageOptions) bool {
			options.GeoM.Scale(.1, 1)
			return true
		},
	}, 100, 20)
	got := draw(100)
	assertRepeatRibbon(t, s, got, 1000, 1000, 1000, false)
	for _, observation := range got {
		if observation.sample.Index == 90 {
			x, _ := observation.options.GeoM.Apply(0, 0)
			if math.Abs(x-90) > 1e-7 {
				t.Fatalf("projected glyph x = %g, want 90", x)
			}
			return
		}
	}
	t.Fatal("explicit pen bounds did not retain a glyph projected into the viewport")
}

func TestRepeatDrawIntegratesMixedFontControlsAcrossLoops(t *testing.T) {
	s, draw := repeatCapture(t, Config{
		Text:     "A{speed:20}{pause:2}{font:wide}AA",
		Fonts:    map[string]Face{"default": face(t, 10), "wide": face(t, 20)},
		Controls: scrolltext.Braces, X: 100, Speed: 10, Gap: 10, Repeat: true,
	}, 100, 20)
	for _, test := range []struct{ seconds, position float64 }{
		{2, 10}, {3, 10}, {5.49, 59.8}, {5.5, 60}, {5.51, 60.1}, {7.5, 70}, {5504, 60030},
	} {
		got := draw(test.seconds)
		assertRepeatRibbon(t, s, got, 100, 100, test.position, false)
		for _, observation := range got {
			want := s.glyphs[observation.sample.Index%3]
			if observation.sample.Glyph.Font != want.Font || observation.sample.Glyph.Advance != want.Advance {
				t.Fatal("a repeated copy lost its mixed-font metrics")
			}
		}
	}
}

func TestRepeatDrawSupportsVerticalShortMessages(t *testing.T) {
	s, draw := repeatCapture(t, Config{
		Glyphs: []Glyph{{Advance: 8}, {Advance: 12}},
		Y:      100, X: 4, Speed: 10, Vertical: true, Repeat: true,
	}, 20, 100)
	for _, seconds := range []float64{0, 1.99, 2.01, 20.01} {
		got := draw(seconds)
		assertRepeatRibbon(t, s, got, 100, 100, seconds*10, true)
		for _, observation := range got {
			if observation.sample.X != 4 {
				t.Fatal("vertical repeat changed the horizontal origin")
			}
		}
	}
}

func TestRepeatDrawHandlesStopsLongGapsAndEmptyText(t *testing.T) {
	t.Run("stopped", func(t *testing.T) {
		s, draw := repeatCapture(t, Config{Glyphs: []Glyph{{Advance: 10}}, X: 20, Repeat: true}, 100, 20)
		assertRepeatRibbon(t, s, draw(1000), 20, 100, 0, false)
	})
	t.Run("long-gap", func(t *testing.T) {
		s, draw := repeatCapture(t, Config{Glyphs: []Glyph{{Advance: 10}}, X: 100, Speed: 10, Gap: 200, Repeat: true}, 100, 20)
		for _, seconds := range []float64{15, 20.99, 21, 21.01, 31.01} {
			assertRepeatRibbon(t, s, draw(seconds), 100, 100, seconds*10, false)
		}
	})
	t.Run("empty", func(t *testing.T) {
		_, draw := repeatCapture(t, Config{Glyphs: []Glyph{}, Speed: 10, Repeat: true}, 100, 20)
		if got := draw(1000); len(got) != 0 {
			t.Fatal("an empty repeated message emitted glyphs")
		}
	})
}
