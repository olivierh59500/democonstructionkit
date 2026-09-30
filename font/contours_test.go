package font

import (
	"math"
	"reflect"
	"testing"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestContourBankCopiesArtworkAndPreservesHoles(t *testing.T) {
	outer := []geometry.Vec2{{X: 0, Y: 0}, {X: 12, Y: 0}, {X: 12, Y: 20}, {X: 0, Y: 20}, {X: 0, Y: 0}}
	hole := []geometry.Vec2{{X: 3, Y: 4}, {X: 3, Y: 16}, {X: 9, Y: 16}, {X: 9, Y: 4}, {X: 3, Y: 4}}
	input := map[rune]ContourGlyph{
		'O': {Contours: [][]geometry.Vec2{outer, hole}, Advance: 14},
		' ': {Advance: 6},
	}
	bank, err := NewContourBank(ContourBankConfig{Glyphs: input, Closed: true})
	if err != nil {
		t.Fatal(err)
	}
	got, exists := bank.Glyph('O')
	if !exists || got.Advance != 14 || !reflect.DeepEqual(got.Contours, [][]geometry.Vec2{outer[:4], hole[:4]}) || bank.VertexCount() != 8 {
		t.Fatalf("lost hole, contour ordering or metrics: %+v, %d vertices", got, bank.VertexCount())
	}
	outer[0].X = 100
	hole[0].Y = 100
	input['O'].Contours[0] = nil
	input['O'] = ContourGlyph{Advance: 99}
	delete(input, ' ')
	got, _ = bank.Glyph('O')
	if got.Advance != 14 || got.Contours[0][0] != (geometry.Vec2{}) || got.Contours[1][0] != (geometry.Vec2{X: 3, Y: 4}) {
		t.Fatal("input mutation changed compiled artwork", got)
	}
	space, exists := bank.Glyph(' ')
	if !exists || space.Advance != 6 || len(space.Contours) != 0 {
		t.Fatal("blank glyph was lost", space, exists)
	}
	if allocs := testing.AllocsPerRun(100, func() { _, _ = bank.Glyph('O') }); allocs != 0 {
		t.Fatalf("glyph lookup allocates: %v", allocs)
	}
}

func TestContourBankNormalizesOnlyExplicitRepeatedEndpoints(t *testing.T) {
	contours := [][]geometry.Vec2{
		{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 0}},
		{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}},
		{{X: 2, Y: 3}},
		{{X: 4, Y: 5}, {X: 4, Y: 5}},
		nil,
	}
	for _, closed := range []bool{false, true} {
		bank, err := NewContourBank(ContourBankConfig{Glyphs: map[rune]ContourGlyph{'A': {Contours: contours}}, Closed: closed})
		if err != nil {
			t.Fatal(err)
		}
		got, _ := bank.Glyph('A')
		lengths := []int{4, 3, 1, 2, 0}
		vertices := 10
		if closed {
			lengths, vertices = []int{3, 3, 1, 1, 0}, 8
		}
		for i, want := range lengths {
			if len(got.Contours[i]) != want {
				t.Fatalf("closed=%v contour %d: got %d points, want %d", closed, i, len(got.Contours[i]), want)
			}
		}
		if bank.VertexCount() != vertices || got.Contours[1][2] != contours[1][2] {
			t.Fatal("unduplicated endpoint was discarded", got)
		}
	}
}

func TestContourBankFallbackUnicodeAndIndependentCharacterList(t *testing.T) {
	fallbackPoints := []geometry.Vec2{{X: 0, Y: 0}, {X: 4, Y: 6}, {X: 8, Y: 0}}
	bank, err := NewContourBank(ContourBankConfig{Glyphs: map[rune]ContourGlyph{
		'雪': {Advance: 24}, 'A': {Advance: 12}, '?': {Advance: 9, Contours: [][]geometry.Vec2{fallbackPoints}}, ' ': {Advance: 5},
	}, Fallback: '?'})
	if err != nil {
		t.Fatal(err)
	}
	if got := bank.Characters(); !reflect.DeepEqual(got, []rune{' ', '?', 'A', '雪'}) {
		t.Fatal("character list is not sorted by Unicode code point", got)
	}
	characters := bank.Characters()
	characters[0] = 'z'
	if bank.Characters()[0] != ' ' {
		t.Fatal("returned character list mutates the bank")
	}
	glyph, exists := bank.Glyph('λ')
	if exists || glyph.Advance != 9 || !reflect.DeepEqual(glyph.Contours, [][]geometry.Vec2{fallbackPoints}) {
		t.Fatal("missing character did not resolve to the fallback", glyph, exists)
	}
	glyph, exists = bank.Glyph('?')
	if !exists || glyph.Advance != 9 {
		t.Fatal("explicit fallback character is unsupported", glyph, exists)
	}
	without, err := NewContourBank(ContourBankConfig{Glyphs: map[rune]ContourGlyph{' ': {Advance: 5}}})
	if err != nil {
		t.Fatal(err)
	}
	glyph, exists = without.Glyph('λ')
	if exists || glyph.Advance != 0 || len(glyph.Contours) != 0 {
		t.Fatal("missing character invented a fallback", glyph, exists)
	}
	var absent *ContourBank
	if glyph, exists := absent.Glyph('A'); exists || glyph.Advance != 0 || absent.VertexCount() != 0 || absent.Characters() != nil {
		t.Fatal("nil bank is unsafe")
	}
}

func TestContourBankAllowsSharedAuthoredAliasesWithIndependentMetrics(t *testing.T) {
	points := []geometry.Vec2{{X: 1, Y: 2}, {X: 5, Y: 2}, {X: 3, Y: 7}}
	bank, err := NewContourBank(ContourBankConfig{Glyphs: map[rune]ContourGlyph{
		'A': {Contours: [][]geometry.Vec2{points}, Advance: 10},
		'a': {Contours: [][]geometry.Vec2{points}, Advance: 8},
	}})
	if err != nil {
		t.Fatal(err)
	}
	points[0] = geometry.Vec2{X: 99, Y: 99}
	upper, upperExists := bank.Glyph('A')
	lower, lowerExists := bank.Glyph('a')
	if !upperExists || !lowerExists || upper.Advance != 10 || lower.Advance != 8 || !reflect.DeepEqual(upper.Contours, lower.Contours) || upper.Contours[0][0] != (geometry.Vec2{X: 1, Y: 2}) {
		t.Fatal("shared authored aliases lost artwork or independent metrics", upper, lower)
	}
}

func TestContourBankRejectsInvalidArtwork(t *testing.T) {
	for _, c := range []ContourBankConfig{
		{},
		{Glyphs: map[rune]ContourGlyph{'A': {}}, Fallback: '?'},
		{Glyphs: map[rune]ContourGlyph{'A': {Advance: -1}}},
		{Glyphs: map[rune]ContourGlyph{'A': {Advance: math.NaN()}}},
		{Glyphs: map[rune]ContourGlyph{'A': {Advance: math.Inf(1)}}},
		{Glyphs: map[rune]ContourGlyph{-1: {}}},
		{Glyphs: map[rune]ContourGlyph{0xd800: {}}},
		{Glyphs: map[rune]ContourGlyph{'A': {Contours: [][]geometry.Vec2{{{X: math.NaN(), Y: 0}}}}}},
		{Glyphs: map[rune]ContourGlyph{'A': {Contours: [][]geometry.Vec2{{{X: 0, Y: math.Inf(-1)}}}}}},
	} {
		if _, err := NewContourBank(c); err == nil {
			t.Fatal("accepted invalid contour artwork", c)
		}
	}
}

func TestContourBankBoundsResourcesBeforeCopying(t *testing.T) {
	t.Run("glyphs", func(t *testing.T) {
		glyphs := make(map[rune]ContourGlyph, 65537)
		for i := 0; i < 65537; i++ {
			glyphs[rune(0x10000+i)] = ContourGlyph{}
		}
		if _, err := NewContourBank(ContourBankConfig{Glyphs: glyphs}); err == nil {
			t.Fatal("accepted too many glyphs")
		}
	})
	t.Run("one contour", func(t *testing.T) {
		glyphs := map[rune]ContourGlyph{'A': {Contours: [][]geometry.Vec2{make([]geometry.Vec2, 65537)}}}
		if _, err := NewContourBank(ContourBankConfig{Glyphs: glyphs}); err == nil {
			t.Fatal("accepted too many points in one contour")
		}
	})
	t.Run("one glyph", func(t *testing.T) {
		glyphs := map[rune]ContourGlyph{'A': {Contours: make([][]geometry.Vec2, 65537)}}
		if _, err := NewContourBank(ContourBankConfig{Glyphs: glyphs}); err == nil {
			t.Fatal("accepted too many contours in one glyph")
		}
	})
	t.Run("total points", func(t *testing.T) {
		points := make([]geometry.Vec2, 65536)
		glyphs := make(map[rune]ContourGlyph)
		for i := 0; i < 17; i++ {
			glyphs[rune('A'+i)] = ContourGlyph{Contours: [][]geometry.Vec2{points}}
		}
		if _, err := NewContourBank(ContourBankConfig{Glyphs: glyphs}); err == nil {
			t.Fatal("accepted too many points across the bank")
		}
	})
	t.Run("total contours", func(t *testing.T) {
		contours := make([][]geometry.Vec2, 65536)
		glyphs := make(map[rune]ContourGlyph)
		for i := 0; i < 17; i++ {
			glyphs[rune('A'+i)] = ContourGlyph{Contours: contours}
		}
		if _, err := NewContourBank(ContourBankConfig{Glyphs: glyphs}); err == nil {
			t.Fatal("accepted too many empty contours across the bank")
		}
	})
}
