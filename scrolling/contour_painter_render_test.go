//go:build dck_gpu_rendercheck

package scrolling

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
)

func contourTestBank(t *testing.T, character rune, advance float64, contours ...[]geometry.Vec2) *font.ContourBank {
	t.Helper()
	bank, err := font.NewContourBank(font.ContourBankConfig{Glyphs: map[rune]font.ContourGlyph{character: {Contours: contours, Advance: advance}}, Closed: true})
	if err != nil {
		t.Fatal(err)
	}
	return bank
}

func contourTestRect(x, y, width, height float64) []geometry.Vec2 {
	return []geometry.Vec2{{X: x, Y: y}, {X: x + width, Y: y}, {X: x + width, Y: y + height}, {X: x, Y: y + height}}
}

// Ray crossing is independent of DCK's triangle fans and Ebitengine's stencil
// implementation. Pixel-centre samples avoid ambiguous axis-aligned boundaries.
func contourTestInside(x, y float64, contours [][]geometry.Vec2) bool {
	inside := false
	for _, contour := range contours {
		if len(contour) < 3 {
			continue
		}
		previous := contour[len(contour)-1]
		for _, current := range contour {
			if (current.Y > y) != (previous.Y > y) && x < (previous.X-current.X)*(y-current.Y)/(previous.Y-current.Y)+current.X {
				inside = !inside
			}
			previous = current
		}
	}
	return inside
}

func contourTestPixels(t *testing.T, target *ebiten.Image, want func(int, int) color.NRGBA) {
	t.Helper()
	bounds := target.Bounds()
	pixels := make([]byte, bounds.Dx()*bounds.Dy()*4)
	target.ReadPixels(pixels)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			expected := want(x, y)
			at := ((y-bounds.Min.Y)*bounds.Dx() + x - bounds.Min.X) * 4
			if pixels[at] != expected.R || pixels[at+1] != expected.G || pixels[at+2] != expected.B || pixels[at+3] != expected.A {
				t.Fatalf("pixel %d,%d: got %v, want %v", x, y, pixels[at:at+4], expected)
			}
		}
	}
}

func TestContourPainterGpuPreservesHolesAndDisconnectedPieces(t *testing.T) {
	testutil.RequireGPU(t)
	contours := [][]geometry.Vec2{contourTestRect(4, 4, 16, 16), contourTestRect(8, 8, 8, 8), contourTestRect(24, 6, 6, 6)}
	bank := contourTestBank(t, 'O', 32, contours...)
	paint := color.NRGBA{R: 255, G: 80, B: 40, A: 255}
	painter, err := NewContourPainter(ContourPainterConfig{Fonts: map[string]*font.ContourBank{"default": bank}, Color: &paint, FillRule: ebiten.FillRuleEvenOdd})
	if err != nil {
		t.Fatal(err)
	}
	defer painter.Close()
	target := ebiten.NewImage(36, 24)
	defer target.Deallocate()
	painter.Paint(target, Sample{Glyph: Glyph{Rune: 'O'}}, ebiten.DrawImageOptions{})
	contourTestPixels(t, target, func(x, y int) color.NRGBA {
		if contourTestInside(float64(x)+.5, float64(y)+.5, contours) {
			return paint
		}
		return color.NRGBA{}
	})
}

func TestContourPainterGpuSharesEvenOddParityAcrossTheCompleteGlyphRun(t *testing.T) {
	testutil.RequireGPU(t)
	bank := contourTestBank(t, 'A', 8, contourTestRect(0, 0, 10, 8))
	for _, count := range []int{2, 3} {
		glyphs := make([]Glyph, count)
		for i := range glyphs {
			glyphs[i] = Glyph{Rune: 'A', Advance: 8}
		}
		config := ContourPainterConfig{Fonts: map[string]*font.ContourBank{"default": bank}, FillRule: ebiten.FillRuleEvenOdd}
		s, err := New(Config{Glyphs: glyphs, Shape: "contours", Modes: map[string]Mode{"contours": {Contours: &config, Map: func(sample Sample, options *ebiten.DrawImageOptions) bool {
			options.GeoM.Translate(-sample.Glyph.Offset, 0)
			return true
		}}}})
		if err != nil {
			t.Fatal(err)
		}
		target := ebiten.NewImage(20, 18)
		state := IdentityState()
		state.X, state.Y, state.Shape = 3, 4, "contours"
		s.DrawAt(target, state)
		contourTestPixels(t, target, func(x, y int) color.NRGBA {
			if count%2 != 0 && x >= 3 && x < 13 && y >= 4 && y < 12 {
				return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			}
			return color.NRGBA{}
		})
		target.Deallocate()
		s.Close()
	}
}

func TestContourPainterGpuUsesDefaultTransformsMixedFontsAndColorScale(t *testing.T) {
	testutil.RequireGPU(t)
	a := contourTestBank(t, 'A', 7, contourTestRect(0, 0, 4, 6))
	b := contourTestBank(t, '雪', 8, contourTestRect(0, 0, 2, 3))
	paint := color.NRGBA{R: 240, G: 128, B: 64, A: 255}
	config := ContourPainterConfig{Fonts: map[string]*font.ContourBank{"a": a, "b": b}, Font: "a", Color: &paint}
	s, err := New(Config{Glyphs: []Glyph{{Rune: 'A', Advance: 7, ScaleX: 2}, {Rune: '雪', Font: "b", Advance: 8, ScaleY: 2}},
		Shape: "contours", Modes: map[string]Mode{"contours": {Contours: &config, Map: func(sample Sample, options *ebiten.DrawImageOptions) bool {
			if sample.Index == 0 {
				options.ColorScale.Scale(1, .5, 1, 1)
			} else {
				options.ColorScale.Scale(.5, 1, 0, 1)
			}
			return true
		}}}})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	paint.R = 0
	config.Fonts["a"] = nil
	target := ebiten.NewImage(28, 14)
	defer target.Deallocate()
	state := IdentityState()
	state.X, state.Y, state.Shape = 14, 3, "contours"
	state.Options.GeoM.Rotate(math.Pi / 2)
	s.DrawAt(target, state)
	contourTestPixels(t, target, func(x, y int) color.NRGBA {
		if x >= 2 && x < 14 && y >= 3 && y < 7 {
			return color.NRGBA{R: 240, G: 64, B: 64, A: 255}
		}
		if x >= 18 && x < 21 && y >= 3 && y < 7 {
			return color.NRGBA{R: 120, G: 128, A: 255}
		}
		return color.NRGBA{}
	})
}

func TestContourPainterGpuRendersControlledTextFromVectorFacesWithoutAtlases(t *testing.T) {
	testutil.RequireGPU(t)
	small := contourTestBank(t, 'A', 40, contourTestRect(0, 0, 2, 3))
	large := contourTestBank(t, 'B', 70, contourTestRect(0, 0, 3, 4))
	smallMetrics, err := font.New(font.Config{Bounds: image.Rect(0, 0, 2, 3), Glyphs: map[rune]font.Glyph{
		'A': {Rect: image.Rect(0, 0, 2, 3), Advance: 4, OffsetX: 1, OffsetY: 1},
	}, LineHeight: 4, SpaceAdvance: 2})
	if err != nil {
		t.Fatal(err)
	}
	largeMetrics, err := font.New(font.Config{Bounds: image.Rect(0, 0, 3, 4), Glyphs: map[rune]font.Glyph{
		'B': {Rect: image.Rect(0, 0, 3, 4), Advance: 7, OffsetX: -1, OffsetY: 2},
	}, LineHeight: 6, SpaceAdvance: 3})
	if err != nil {
		t.Fatal(err)
	}
	contours := ContourPainterConfig{FillRule: ebiten.FillRuleEvenOdd}
	config := Config{Text: "A{font:large}B", Controls: scrolltext.Braces, Font: "small", Shape: "vector",
		Fonts: map[string]Face{"small": {Contours: small, Metrics: smallMetrics}, "large": {Contours: large, Metrics: largeMetrics}},
		Modes: map[string]Mode{"vector": {Contours: &contours, Map: func(sample Sample, options *ebiten.DrawImageOptions) bool {
			if sample.Glyph.Font == "small" {
				options.ColorScale.Scale(1, 0, 0, 1)
			} else {
				options.ColorScale.Scale(0, 0, 1, 1)
			}
			return true
		}}}}
	s, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	if s.GlyphCount() != 2 || s.Length() != 11 || s.glyphs[0].Advance != 4 || s.glyphs[1].Offset != 4 || s.glyphs[1].Font != "large" || s.glyphs[0].Image != nil || s.glyphs[1].Image != nil {
		t.Fatal("vector text used artwork advances or required dummy bitmaps", s.glyphs, s.Length())
	}
	target := ebiten.NewImage(12, 12)
	defer target.Deallocate()
	state := IdentityState()
	state.X, state.Y, state.Shape = 2, 3, "vector"
	s.DrawAt(target, state)
	contourTestPixels(t, target, func(x, y int) color.NRGBA {
		if x >= 3 && x < 5 && y >= 4 && y < 7 {
			return color.NRGBA{R: 255, A: 255}
		}
		if x >= 5 && x < 8 && y >= 5 && y < 9 {
			return color.NRGBA{B: 255, A: 255}
		}
		return color.NRGBA{}
	})
	s.Close()
	for _, bank := range []*font.ContourBank{small, large} {
		if bank.VertexCount() != 4 {
			t.Fatal("closing the text renderer changed borrowed vector artwork")
		}
	}
	if metric, supported := smallMetrics.Glyph('A'); !supported || metric.Advance != 4 {
		t.Fatal("closing the text renderer changed borrowed layout metrics")
	}
	for _, face := range []Face{{Metrics: smallMetrics}, {Contours: small}} {
		invalid := config
		invalid.Text = "A"
		invalid.Fonts = map[string]Face{"small": face}
		if bad, err := New(invalid); err == nil {
			bad.Close()
			t.Fatal("accepted a face missing its artwork or layout metrics")
		}
	}
}

func TestContourPainterGpuBorrowsRowMaterialWithNonzeroOrigins(t *testing.T) {
	testutil.RequireGPU(t)
	material := ebiten.NewImage(16, 16)
	defer material.Deallocate()
	rows := []color.NRGBA{{R: 255, A: 255}, {G: 255, A: 255}, {B: 255, A: 255}, {R: 255, G: 255, A: 255}}
	for i, paint := range rows {
		material.SubImage(image.Rect(5, 7+i, 6, 8+i)).(*ebiten.Image).Fill(paint)
	}
	texture := material.SubImage(image.Rect(5, 7, 6, 11)).(*ebiten.Image)
	bank := contourTestBank(t, 'A', 8, contourTestRect(0, 0, 6, 4))
	painter, err := NewContourPainter(ContourPainterConfig{Fonts: map[string]*font.ContourBank{"default": bank}, Texture: texture,
		UV: func(sample ContourSample, mapped geometry.Vec2) geometry.Vec2 {
			if sample.Glyph.Rune != 'A' || mapped.Y != sample.Point.Y+10 {
				t.Fatal("material callback lost authored points or destination mapping", sample, mapped)
			}
			return geometry.Vec2{X: 5.5, Y: mapped.Y - 3}
		}})
	if err != nil {
		t.Fatal(err)
	}
	target := ebiten.NewImage(24, 24)
	defer target.Deallocate()
	destination := target.SubImage(image.Rect(8, 9, 16, 17)).(*ebiten.Image)
	var options ebiten.DrawImageOptions
	options.GeoM.Translate(9, 10)
	painter.Paint(destination, Sample{Glyph: Glyph{Rune: 'A'}}, options)
	contourTestPixels(t, target, func(x, y int) color.NRGBA {
		if x >= 9 && x < 15 && y >= 10 && y < 14 {
			return rows[y-10]
		}
		return color.NRGBA{}
	})
	painter.Close()
	painter.Close()
	copy := ebiten.NewImage(1, 4)
	defer copy.Deallocate()
	copy.DrawImage(texture, nil)
	contourTestPixels(t, copy, func(_ int, y int) color.NRGBA { return rows[y] })
	if glyph, supported := bank.Glyph('A'); !supported || len(glyph.Contours) != 1 {
		t.Fatal("closing a painter damaged its borrowed font")
	}
}

func TestContourPainterGpuDiscardsAnEntireInvalidContour(t *testing.T) {
	testutil.RequireGPU(t)
	bank := contourTestBank(t, 'A', 20, contourTestRect(2, 2, 6, 6), contourTestRect(12, 2, 6, 6))
	for _, failure := range []string{"hidden", "nonfinite", "float32-overflow", "uv"} {
		painter, err := NewContourPainter(ContourPainterConfig{Fonts: map[string]*font.ContourBank{"default": bank},
			Map: func(sample ContourSample, transform ebiten.GeoM) (geometry.Vec2, bool) {
				x, y := transform.Apply(sample.Point.X, sample.Point.Y)
				if sample.Contour == 0 && sample.Vertex == 2 {
					switch failure {
					case "hidden":
						return geometry.Vec2{}, false
					case "nonfinite":
						x = math.Inf(1)
					case "float32-overflow":
						x = float64(math.MaxFloat32) * 2
					}
				}
				return geometry.Vec2{X: x, Y: y}, true
			}, UV: func(sample ContourSample, _ geometry.Vec2) geometry.Vec2 {
				if failure == "uv" && sample.Contour == 0 && sample.Vertex == 2 {
					return geometry.Vec2{X: math.NaN()}
				}
				return geometry.Vec2{}
			}})
		if err != nil {
			t.Fatal(err)
		}
		target := ebiten.NewImage(24, 12)
		painter.Paint(target, Sample{Glyph: Glyph{Rune: 'A'}}, ebiten.DrawImageOptions{})
		contourTestPixels(t, target, func(x, y int) color.NRGBA {
			if x >= 12 && x < 18 && y >= 2 && y < 8 {
				return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			}
			return color.NRGBA{}
		})
		if painter.Err() != nil {
			t.Fatal("an intentionally hidden contour became a resource error", failure, painter.Err())
		}
		target.Deallocate()
		painter.Close()
	}
}

func TestContourPainterGpuRejectsBudgetsAndAbortsIncompleteParityRuns(t *testing.T) {
	testutil.RequireGPU(t)
	bank := contourTestBank(t, 'A', 8, contourTestRect(0, 0, 6, 6))
	fonts := map[string]*font.ContourBank{"default": bank}
	for _, config := range []ContourPainterConfig{
		{}, {Fonts: fonts, Font: "missing"}, {Fonts: map[string]*font.ContourBank{"default": nil}},
		{Fonts: fonts, BatchTriangles: 1}, {Fonts: fonts, BatchTriangles: 20001},
		{Fonts: fonts, FillRule: ebiten.FillRule(-1)}, {Fonts: fonts, FillRule: ebiten.FillRule(3)},
		{Fonts: map[string]*font.ContourBank{"default": contourTestBank(t, 'A', 8, []geometry.Vec2{{X: 0}, {X: 3}, {X: 6, Y: 3}, {X: 6, Y: 6}, {Y: 6}})}, BatchTriangles: 2},
	} {
		if painter, err := NewContourPainter(config); err == nil {
			painter.Close()
			t.Fatal("accepted invalid contour configuration")
		}
	}
	config := ContourPainterConfig{Fonts: fonts, FillRule: ebiten.FillRuleEvenOdd, BatchTriangles: 2}
	s, err := New(Config{Glyphs: []Glyph{{Rune: 'A', Advance: 8}, {Rune: 'A', Advance: 8}}, Shape: "contours", Modes: map[string]Mode{"contours": {Contours: &config}}})
	if err != nil {
		t.Fatal(err)
	}
	target := ebiten.NewImage(20, 10)
	defer target.Deallocate()
	s.Draw(target)
	if s.Err() == nil {
		t.Fatal("overflowing a complete parity run did not report its resource budget")
	}
	contourTestPixels(t, target, func(int, int) color.NRGBA { return color.NRGBA{} })
	owned := s.ContourPainterController("contours")
	if owned == nil || !owned.ownTexture {
		t.Fatal("contour mode did not own its generated material")
	}
	s.Close()
	s.Close()
	if owned.texture != nil || owned.batch != nil || s.ContourPainterController("contours") != nil {
		t.Fatal("closed mode retained generated drawing resources")
	}
	owned.Paint(target, Sample{Glyph: Glyph{Rune: 'A'}}, ebiten.DrawImageOptions{})
}
