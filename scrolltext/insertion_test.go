package scrolltext

import (
	"reflect"
	"testing"
)

func TestInsertionMatchesIndependentTransportTiming(t *testing.T) {
	tokens := []InsertionToken{{Command: 'p', Payload: []byte{2}}, {Command: 's', Payload: []byte{9}}, {Glyph: true, Rune: 'A', Advance: 13}, {Glyph: true, Rune: 'B', Advance: 9}, {Command: 's', Payload: []byte{3}}, {Glyph: true, Rune: 'C', Advance: 17}}
	p, err := NewInsertionProgram(InsertionProgramConfig{Tokens: tokens, Speed: 4, TargetSpeed: 4, SpeedStep: 1, Divisor: 2, Entry: 368, AlignMask: 15, AlignBias: 14, RetireBefore: -16, OnCommand: func(p *InsertionProgram, t InsertionToken) error {
		if t.Command == 'p' {
			return p.SetPause(int(t.Payload[0]))
		}
		return p.SetTargetSpeed(int(t.Payload[0]))
	}})
	if err != nil {
		t.Fatal(err)
	}
	cursor, remaining, speed, target, pause, distance, first, fetched := 0, 0, 4, 4, 0, 0, 0, 0
	origins := make([]int, 3)
	for tick := 0; tick < 100; tick++ {
		active := true
		if pause > 0 {
			pause--
		} else {
			remaining -= speed / 2
			if remaining < 0 {
				inserted := false
				for cursor < len(tokens) && !inserted {
					token := tokens[cursor]
					cursor++
					if token.Glyph {
						origins[fetched] = distance + 368 + ((remaining + 14) & 15)
						fetched++
						remaining += token.Advance
						inserted = true
					} else if token.Command == 'p' {
						pause = int(token.Payload[0])
					} else {
						target = int(token.Payload[0])
					}
				}
				if !inserted {
					active = false
				}
			}
			if active {
				distance += speed / 2
				if speed < target {
					speed++
				} else if speed > target {
					speed--
				}
			}
		}
		if active {
			for first < fetched && origins[first]-distance < -16 {
				first++
			}
		}
		got, err := p.Step()
		if err != nil {
			t.Fatal(err)
		}
		s := p.State()
		if got != active || s.Cursor != cursor || s.Remaining != remaining || s.Speed != speed || s.TargetSpeed != target || s.Pause != pause || s.Distance != distance || s.First != first || s.Fetched != fetched || !reflect.DeepEqual(s.Origins, origins) {
			t.Fatalf("tick%d got%+v native%v", tick, s, []int{cursor, remaining, speed, target, pause, distance, first, fetched})
		}
		if !active {
			break
		}
	}
}

func TestInsertionOwnsInputsAndKeepsUpdatesBounded(t *testing.T) {
	tokens := []InsertionToken{{Glyph: true, Rune: '雪', Advance: 7}, {Command: 1, Payload: []byte{3}}, {Glyph: true, Rune: 'Z', Advance: 8}}
	calls := 0
	p, err := NewInsertionProgram(InsertionProgramConfig{Tokens: tokens, Speed: 2, TargetSpeed: 2, Entry: 10, OnCommand: func(_ *InsertionProgram, t InsertionToken) error { calls += int(t.Payload[0]); return nil }})
	if err != nil {
		t.Fatal(err)
	}
	tokens[0].Rune = 'X'
	tokens[1].Payload[0] = 9
	if p.Glyphs()[0].Rune != '雪' {
		t.Fatal("glyph data stayed borrowed")
	}
	for i := 0; i < 30; i++ {
		p.Step()
	}
	if calls != 3 || !p.State().Finished {
		t.Fatal("command ownership or completion", calls, p.State())
	}
	origins := p.State().Origins
	p.Reset()
	s := p.State()
	if s.Finished || s.Cursor != 0 || s.Distance != 0 || s.Speed != 2 || s.TargetSpeed != 2 || &s.Origins[0] != &origins[0] {
		t.Fatal("reset did not retain the origin bank and initial transport", s)
	}
	p, err = NewInsertionProgram(InsertionProgramConfig{Tokens: []InsertionToken{{Glyph: true, Rune: 'A', Advance: 1000}}, Speed: 2, TargetSpeed: 2})
	if err != nil {
		t.Fatal(err)
	}
	if n := testing.AllocsPerRun(100, func() {
		if _, err := p.Step(); err != nil {
			panic(err)
		}
	}); n != 0 {
		t.Fatal("insertion allocates", n)
	}
	for _, bad := range []InsertionProgramConfig{{Speed: -1}, {Divisor: -1}, {AlignMask: -1}, {Tokens: []InsertionToken{{Glyph: true, Advance: 0}}}} {
		if _, err := NewInsertionProgram(bad); err == nil {
			t.Fatal("invalid insertion config", bad)
		}
	}
}
