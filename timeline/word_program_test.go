package timeline

import "testing"

func TestWordProgramBoundariesWrapOwnershipAndRepeat(t *testing.T) {
	initial := []int16{32766, -32767}
	cues := []WordCue{{Ticks: 2, Delta: []int16{2, -2}}, {Ticks: 2, Delta: []int16{-3, 4}}}
	p, err := NewWordProgram(WordProgramConfig{Initial: initial, Cues: cues})
	if err != nil {
		t.Fatal(err)
	}
	initial[0] = 0
	cues[0].Delta[0] = 9
	want := [][2]int16{{-32768, 32767}, {-32766, 32765}, {32767, -32767}, {32764, -32763}}
	for i, v := range want {
		if !p.Step() || p.State()[0] != v[0] || p.State()[1] != v[1] {
			t.Fatal("cue phase or word wrap", i, p.State(), v)
		}
	}
	if p.Step() || !p.Finished() || p.Step() {
		t.Fatal("finite completion changed")
	}
	p.Reset()
	if p.State()[0] != 32766 || !p.Step() {
		t.Fatal("reset failed")
	}
	r, err := NewWordProgram(WordProgramConfig{Initial: []int16{5}, Cues: []WordCue{{Ticks: 1, Delta: []int16{2}}}, Repeat: true, ResetOnRepeat: true})
	if err != nil {
		t.Fatal(err)
	}
	for range 100 {
		if !r.Step() || r.State()[0] != 7 {
			t.Fatal("repeat reset moved")
		}
	}
	if n := testing.AllocsPerRun(100, func() { r.Step() }); n != 0 {
		t.Fatal("program allocates per tick", n)
	}
	for _, bad := range []WordProgramConfig{{}, {Initial: []int16{1}, Cues: []WordCue{{Ticks: 0, Delta: []int16{1}}}}, {Initial: []int16{1}, Cues: []WordCue{{Ticks: 1}}}} {
		if _, err := NewWordProgram(bad); err == nil {
			t.Fatal("invalid program accepted")
		}
	}
}
