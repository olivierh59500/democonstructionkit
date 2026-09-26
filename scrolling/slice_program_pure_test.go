package scrolling

import (
	"math"
	"slices"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

type legacyCuedSlices struct {
	tokens                 []SliceToken
	items                  []DNASlice
	offsets                []float64
	head, token, strip     int
	pauseTicks             int
	paused                 bool
	rotation, rotationStep float64
}

func (r *legacyCuedSlices) insert() {
	token := r.tokens[r.token]
	if token.Control != "" {
		r.paused = true
		switch token.Control {
		case "^":
			r.pauseTicks, r.rotationStep = 275, -1
		case "&":
			r.pauseTicks, r.rotationStep = 275, 1
		case "#":
			r.pauseTicks, r.rotationStep = 250, -1
		case "%":
			r.pauseTicks, r.rotationStep = 225, -1
		}
		r.nextToken()
		return
	}
	r.head++
	if r.head == len(r.items) {
		r.head = 0
	}
	previous := (r.head + len(r.items) - 2) % len(r.items)
	tail := (r.head + len(r.items) - 1) % len(r.items)
	r.items[tail] = DNASlice{Glyph: token.Glyph, Frame: r.items[previous].Frame, Slice: r.strip}
	r.strip++
	if r.strip == 8 {
		r.strip = 0
		r.nextToken()
	}
}

func (r *legacyCuedSlices) nextToken() {
	r.token++
	if r.token == len(r.tokens) {
		r.token = 2
	}
}

func (r *legacyCuedSlices) rotate() {
	r.rotation += r.rotationStep
	if r.rotation >= 30 {
		r.rotation -= 30
	}
	if r.rotation < 0 {
		r.rotation += 30
	}
	for i := range r.items {
		index := r.head + i
		if index >= len(r.items) {
			index -= len(r.items)
		}
		frame := r.rotation + r.offsets[i]
		if frame >= 30 {
			frame -= 30
		} else if frame < 0 {
			frame += 30
		}
		r.items[index].Frame = int(frame)
	}
}

func (r *legacyCuedSlices) step() {
	if !r.paused {
		r.insert()
	} else {
		r.pauseTicks--
		if r.pauseTicks == 0 {
			r.paused = false
			r.rotationStep = .35
		}
	}
	r.rotate()
}

func TestSliceProgramMatchesSourceThroughControlsAndLoop(t *testing.T) {
	tokens := []SliceToken{
		{Glyph: 0, Width: 16}, {Control: "^"},
		{Glyph: 1, Width: 16}, {Control: "&"},
		{Glyph: 2, Width: 16}, {Control: "#"},
		{Glyph: 3, Width: 16}, {Control: "%"},
		{Glyph: 4, Width: 16},
	}
	offsets := make([]float64, 24)
	for i := range offsets {
		offsets[i] = math.Sin(float64(i)*.05) * 15
	}
	p, err := NewSliceProgram(SliceProgramConfig{
		Stream: SliceStreamConfig{Tokens: tokens, Capacity: 24, SliceWidth: 2, Repeat: true, LoopStart: 2},
		Clock: motion.CuedScrollClockConfig{
			InitialPauseTicks: 250, InitialTextStep: 1, InitialRotationStep: .35,
			ResumeTextStep: 1, ResumeRotationStep: .35, RotationFrames: 30,
			Cues: map[string]motion.ScrollCue{
				"^": {PauseTicks: 275, SetRotation: true, RotationStep: -1},
				"&": {PauseTicks: 275, SetRotation: true, RotationStep: 1},
				"#": {PauseTicks: 250, SetRotation: true, RotationStep: -1},
				"%": {PauseTicks: 225, SetRotation: true, RotationStep: -1},
			},
		},
		Offsets: offsets,
	})
	if err != nil {
		t.Fatal(err)
	}
	reference := legacyCuedSlices{tokens: tokens, items: make([]DNASlice, 24),
		offsets: offsets, pauseTicks: 250, rotationStep: .35}
	for range 320 {
		reference.insert()
		reference.rotate()
	}
	if err := p.Warmup(320, 1); err != nil {
		t.Fatal(err)
	}
	for tick := 0; tick < 4000; tick++ {
		reference.step()
		if err := p.Step(); err != nil {
			t.Fatal(err)
		}
		state := p.Clock().State()
		token, strip := p.Stream().Cursor()
		if token != reference.token || strip != reference.strip || p.Stream().Head() != reference.head ||
			state.Paused != reference.paused || state.PauseTicks != reference.pauseTicks ||
			state.RotationStep != reference.rotationStep || state.Rotation != reference.rotation || state.TextStep != 1 ||
			!slices.Equal(p.Stream().Slices(), reference.items) {
			t.Fatalf("slice program diverged at tick %d: cursor=%d,%d state=%+v", tick, token, strip, state)
		}
	}
	if got := testing.AllocsPerRun(100, func() { _ = p.Step() }); got != 0 {
		t.Fatalf("slice program step allocated %.2f objects", got)
	}
}

func BenchmarkSliceProgram240Strips(b *testing.B) {
	offsets := make([]float64, 240)
	for i := range offsets {
		offsets[i] = math.Sin(float64(i)*.05) * 15
	}
	p, err := NewSliceProgram(SliceProgramConfig{
		Stream: SliceStreamConfig{Tokens: []SliceToken{{Glyph: 0, Width: 16},
			{Control: "^"}, {Glyph: 1, Width: 16}}, Capacity: 240, SliceWidth: 2,
			Repeat: true, LoopStart: 2},
		Clock: motion.CuedScrollClockConfig{InitialTextStep: 1,
			InitialRotationStep: .35, ResumeTextStep: 1, ResumeRotationStep: .35,
			RotationFrames: 30, Cues: map[string]motion.ScrollCue{
				"^": {PauseTicks: 275, SetRotation: true, RotationStep: -1}}},
		Offsets: offsets,
	})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if err := p.Step(); err != nil {
			b.Fatal(err)
		}
	}
}
