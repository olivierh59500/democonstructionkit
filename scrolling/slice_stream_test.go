package scrolling

import (
	"math"
	"slices"
	"testing"
)

func TestSliceStreamMixedWidthsControlsAndLoopPrefix(t *testing.T) {
	stream, err := NewSliceStream(SliceStreamConfig{Capacity: 5, SliceWidth: 1, Tokens: []SliceToken{{Glyph: 4, Width: 2}, {Control: "pause"}, {Glyph: 7, Width: 3}, {Glyph: -1, Width: 1}}, Repeat: true, LoopStart: 2, Initial: DNASlice{Glyph: -1, Frame: 9}})
	if err != nil {
		t.Fatal(err)
	}
	if n := stream.Step(4, func(event SliceControl) bool {
		if event.Token != 1 || event.Name != "pause" {
			t.Fatal(event)
		}
		return true
	}); n != 2 {
		t.Fatal(n)
	}
	if token, part := stream.Cursor(); token != 2 || part != 0 {
		t.Fatal(token, part)
	}
	if n := stream.Step(4, nil); n != 4 {
		t.Fatal(n)
	}
	if token, part := stream.Cursor(); token != 2 || part != 0 {
		t.Fatal(token, part)
	}
	logical := make([]DNASlice, 5)
	for i := range logical {
		logical[i] = stream.Slices()[(stream.Head()+i)%5]
	}
	want := []DNASlice{{Glyph: 4, Frame: 9, Slice: 1}, {Glyph: 7, Frame: 9, Slice: 0}, {Glyph: 7, Frame: 9, Slice: 1}, {Glyph: 7, Frame: 9, Slice: 2}, {Glyph: -1, Frame: 9, Slice: 0}}
	if !slices.Equal(logical, want) {
		t.Fatalf("%v != %v", logical, want)
	}
	stream.Reset()
	if stream.Head() != 0 || stream.Ended() {
		t.Fatal("reset state")
	}
	for _, s := range stream.Slices() {
		if s != (DNASlice{Glyph: -1, Frame: 9}) {
			t.Fatal(s)
		}
	}
}

func TestSliceStreamCoarseFinalColumnFramesAndOwnership(t *testing.T) {
	tokens := []SliceToken{{Glyph: 3, Width: 5}}
	s, err := NewSliceStream(SliceStreamConfig{Tokens: tokens, Capacity: 4, SliceWidth: 2, Initial: DNASlice{Frame: 2}})
	if err != nil {
		t.Fatal(err)
	}
	tokens[0].Width = 200
	if n := s.Step(9, nil); n != 3 || !s.Ended() {
		t.Fatal(n, s.Ended())
	}
	if err = s.SetFrames(-1, []float64{0, 1, 31, -61}, 30); err != nil {
		t.Fatal(err)
	}
	want := []int{29, 0, 0, 28}
	for i, w := range want {
		if got := s.Slices()[(s.Head()+i)%4].Frame; got != w {
			t.Fatalf("frame%d %d!=%d", i, got, w)
		}
	}
	if s.Step(1, nil) != 0 {
		t.Fatal("advanced completed stream")
	}
	single, err := NewSliceStream(SliceStreamConfig{Tokens: []SliceToken{{Glyph: 1, Width: 1}}, Capacity: 1, SliceWidth: 1, Repeat: true, Initial: DNASlice{Frame: 4}})
	if err != nil {
		t.Fatal(err)
	}
	single.Step(100, nil)
	if single.Slices()[0] != (DNASlice{Glyph: 1, Frame: 4}) {
		t.Fatal(single.Slices())
	}
	if got := testing.AllocsPerRun(100, func() { single.Step(8, nil); _ = single.SetFrames(4, nil, 30) }); got != 0 {
		t.Fatal(got)
	}
}

func TestSliceStreamRejectsInvalidConfigurationAndProfiles(t *testing.T) {
	c := SliceStreamConfig{Tokens: []SliceToken{{Glyph: 1, Width: 1}}, Capacity: 2, SliceWidth: 1}
	for _, bad := range []SliceStreamConfig{{}, {Tokens: c.Tokens, Capacity: 0, SliceWidth: 1}, {Tokens: c.Tokens, Capacity: 2, SliceWidth: 0}, {Tokens: c.Tokens, Capacity: 2, SliceWidth: 1, LoopStart: 1}, {Tokens: []SliceToken{{Width: 1, Control: "event"}}, Capacity: 2, SliceWidth: 1}} {
		if _, err := NewSliceStream(bad); err == nil {
			t.Fatal("invalid config accepted")
		}
	}
	s, err := NewSliceStream(c)
	if err != nil {
		t.Fatal(err)
	}
	if s.SetFrames(math.NaN(), nil, 30) == nil || s.SetFrames(0, nil, 0) == nil || s.SetFrames(0, []float64{1}, 30) == nil || s.SetFrames(0, []float64{0, math.Inf(1)}, 30) == nil || s.SetFrames(math.MaxFloat64, []float64{0, math.MaxFloat64}, 30) == nil {
		t.Fatal("invalid profile accepted")
	}
}
