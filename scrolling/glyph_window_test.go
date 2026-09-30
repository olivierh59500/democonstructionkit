package scrolling

import (
	"math"
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
)

func TestGlyphWindowSamplesAuthoredCharactersWithoutAdvancingClock(t *testing.T) {
	message := []rune("A雪BC")
	cursor, calls := 1, 0
	var slots []int
	s, err := New(Config{GlyphWindow: &GlyphWindowConfig{Count: 3, Advance: 4, Glyph: func(slot int) Glyph {
		calls++
		slots = append(slots, slot)
		return Glyph{Rune: message[(cursor+slot)%len(message)], Font: []string{"small", "large"}[slot%2], Advance: 99, Offset: 999, X: 1, Y: 2}
	}}})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	frame := kit.Frame{Tick: 195, Time: 3.25, Delta: 1.0 / 60}
	if err := s.Update(frame); err != nil {
		t.Fatal(err)
	}
	if calls != 0 || s.GlyphCount() != 3 || s.Length() != 12 {
		t.Fatal("construction or Update consumed authored glyphs", calls, s.Length())
	}
	state := IdentityState()
	state.X, state.Y, state.Time = 10, 20, frame.Time
	var samples []Sample
	state.Paint = func(_ *ebiten.Image, sample Sample, op ebiten.DrawImageOptions) {
		samples = append(samples, sample)
		x, y := op.GeoM.Apply(0, 0)
		if x != sample.X || y != sample.Y {
			t.Fatal("sampled layout and draw transform disagree", sample, x, y)
		}
	}
	s.DrawAt(nil, state)
	first := append([]Sample(nil), samples...)
	samples = nil
	s.DrawAt(nil, state)
	if calls != 6 || cursor != 1 || s.frame != frame || !reflect.DeepEqual(slots, []int{0, 1, 2, 0, 1, 2}) || !reflect.DeepEqual(samples, first) {
		t.Fatal("drawing advanced transport or failed to resample its fixed clock", calls, slots, samples)
	}
	for i, sample := range samples {
		if sample.Glyph.Rune != message[i+1] || sample.Glyph.Font != []string{"small", "large"}[i%2] || sample.Glyph.Advance != 4 || sample.Glyph.Offset != float64(i*4) || sample.Glyph.ScaleX != 1 || sample.Glyph.ScaleY != 1 || sample.X != float64(11+i*4) || sample.Y != 22 || sample.Time != 3.25 {
			t.Fatal("authored character/font or fixed layout was lost", sample)
		}
	}
	cursor = 2
	samples = nil
	s.DrawAt(nil, state)
	if samples[0].Glyph.Rune != 'B' || samples[2].Glyph.Rune != 'A' {
		t.Fatal("draw did not sample the current authored cursor", samples)
	}
}

func TestGlyphWindowCopiesConfigurationAndWrapsVirtualIndices(t *testing.T) {
	window := GlyphWindowConfig{Count: 3, Advance: 4, Glyph: func(slot int) Glyph { return Glyph{Rune: []rune("ABC")[slot]} }}
	s, err := New(Config{GlyphWindow: &window})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	window.Count, window.Advance = 1, 99
	window.Glyph = func(int) Glyph { return Glyph{Rune: 'X'} }
	state := s.Window(-1, 12)
	var runes []rune
	var indices []int
	var positions []float64
	state.Paint = func(_ *ebiten.Image, sample Sample, _ ebiten.DrawImageOptions) {
		runes = append(runes, sample.Glyph.Rune)
		indices = append(indices, sample.Index)
		positions = append(positions, sample.X)
	}
	s.DrawAt(nil, state)
	if s.GlyphCount() != 3 || s.Length() != 12 || !reflect.DeepEqual(runes, []rune("CAB")) || !reflect.DeepEqual(indices, []int{-1, 0, 1}) || !reflect.DeepEqual(positions, []float64{0, 4, 8}) {
		t.Fatal("window configuration changed or circular pen layout broke", runes, indices, positions)
	}
}

func TestGlyphWindowRejectsAmbiguousOrUnboundedLayouts(t *testing.T) {
	glyph := func(int) Glyph { return Glyph{Rune: 'A'} }
	for _, window := range []GlyphWindowConfig{
		{}, {Count: -1, Advance: 1, Glyph: glyph}, {Count: 65537, Advance: 1, Glyph: glyph},
		{Count: 1, Advance: 0, Glyph: glyph}, {Count: 1, Advance: -1, Glyph: glyph},
		{Count: 1, Advance: math.NaN(), Glyph: glyph}, {Count: 1, Advance: math.Inf(1), Glyph: glyph},
		{Count: 2, Advance: math.MaxFloat64, Glyph: glyph}, {Count: 1, Advance: 1},
	} {
		if s, err := New(Config{GlyphWindow: &window}); err == nil {
			s.Close()
			t.Fatal("accepted invalid authored window", window)
		}
	}
	window := GlyphWindowConfig{Count: 3, Advance: 4, Glyph: glyph}
	for _, config := range []Config{
		{Text: "A"}, {Tokens: []scrolltext.Token{{Kind: scrolltext.Text, Text: "A"}}},
		{Glyphs: []Glyph{{Rune: 'A', Advance: 1}}}, {Page: &PageConfig{}},
		{Recycled: &RecycledConfig{}}, {Projected: &ProjectedConfig{}}, {Crawl: &CrawlConfig{}},
		{MaxGlyphsPerDraw: 2},
	} {
		config.GlyphWindow = &window
		if s, err := New(config); err == nil {
			s.Close()
			t.Fatal("accepted two transports or an insufficient drawing budget")
		}
	}
}
