package scrolltext

import (
	"bytes"
	"errors"
	"testing"
)

func TestByteWindowNativeSignedClocksAndSequentialEnd(t *testing.T) {
	text := append([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"), 128)
	for _, example := range []struct {
		name     string
		advance  byte
		crossing ByteCrossing
	}{
		{"circular", 30, NonNegative},
		{"perspective", 22, BelowZero},
	} {
		t.Run(example.name, func(t *testing.T) {
			w, err := NewByteWindow(ByteWindowConfig{Text: text, Slots: 8, Step: -2, Advance: example.advance, Crossing: example.crossing})
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(w.Letters(), make([]byte, 8)) || w.Cursor() != 0 || w.Position() != 0 {
				t.Fatal("constructor advanced or filled the window")
			}
			position, cursor := byte(0), 0
			var letters [8]byte
			finished := false
			for tick := 0; tick < 600; tick++ {
				if !finished {
					position -= 2
					crossed := int8(position) < 0
					if example.crossing == NonNegative {
						crossed = int8(position) >= 0
					}
					if crossed {
						position += example.advance
						cursor++
					}
					for slot := range letters {
						if text[cursor+slot] >= 128 {
							finished = true
							break
						}
						letters[slot] = text[cursor+slot]
					}
				}
				_, err := w.Step()
				if err != nil {
					t.Fatal(err)
				}
				if w.Position() != position || w.Cursor() != cursor || w.Finished() != finished || !bytes.Equal(w.Letters(), letters[:]) {
					t.Fatalf("tick %d: position %d/%d cursor %d/%d end %v/%v letters %q/%q", tick, w.Position(), position, w.Cursor(), cursor, w.Finished(), finished, w.Letters(), letters)
				}
			}
		})
	}
}

func TestByteWindowNativeInitialControlPrefixKeepsRawCursor(t *testing.T) {
	// The authored circular stream starts with controls. Its first lookahead
	// skips the prefix, but the first head crossing lands on the pause payload
	// '1'. The following crossing executes the profile command at byte two.
	text := append([]byte{'|', 49, '>', 0, 22, 0, 64, 192}, []byte("  ABCDEFGH")...)
	text = append(text, '|', 33, '>', 2, 18, 4, 32, 96)
	text = append(text, []byte("IJKLMNOPQRSTUV")...)
	text = append(text, 128)
	profile := [5]byte{2, 22, 4, 64, 192}
	actualProfile := profile
	pauses := 0
	w, err := NewByteWindow(ByteWindowConfig{Text: text, Slots: 8, Step: -2, Advance: 30, Crossing: NonNegative,
		Commands: map[byte]ByteCommand{
			'|': {Payload: 1, Apply: func(p []byte) (int, error) { pauses++; return (int(p[0]) - 32) * 25, nil }},
			'>': {Payload: 5, Apply: func(p []byte) (int, error) { copy(actualProfile[:], p); return 0, nil }},
		}})
	if err != nil {
		t.Fatal(err)
	}
	position, cursor, pause := byte(0), 0, 0
	var letters [8]byte
	finished := false
	seenRawPausePayload := false
	for tick := 0; tick < 600; tick++ {
		wasFinished := finished
		if !finished {
			if pause > 0 {
				pause--
			} else {
				position -= 2
				if int8(position) >= 0 {
					position += 30
					cursor++
					if text[cursor] == '|' {
						pause = (int(text[cursor+1]) - 32) * 25
						cursor += 2
					}
					if text[cursor] == '>' {
						copy(profile[:], text[cursor+1:cursor+6])
						cursor += 6
					}
				}
			}
			index := cursor
			for slot := range letters {
				for {
					if index >= len(text) || text[index] >= 128 {
						finished = true
						break
					}
					value := text[index]
					index++
					if value == '|' {
						index++
						continue
					}
					if value == '>' {
						index += 5
						continue
					}
					letters[slot] = value
					break
				}
				if finished {
					break
				}
			}
		}
		ok, err := w.Step()
		if err != nil || ok == wasFinished || w.Position() != position || w.Cursor() != cursor || w.Pause() != pause || w.Finished() != finished || actualProfile != profile || !bytes.Equal(w.Letters(), letters[:]) {
			t.Fatalf("tick %d: ok=%v err=%v position=%d/%d cursor=%d/%d pause=%d/%d end=%v/%v profile=%v/%v letters=%q/%q", tick, ok, err, w.Position(), position, w.Cursor(), cursor, w.Pause(), pause, w.Finished(), finished, actualProfile, profile, w.Letters(), letters)
		}
		if cursor == 1 {
			seenRawPausePayload = true
			if letters[0] != '1' || pauses != 0 {
				t.Fatal("initial pause command was executed or its raw payload was skipped")
			}
		}
	}
	if !seenRawPausePayload || !finished || pauses != 1 {
		t.Fatal("reference did not exercise raw payload, later pause and termination")
	}
}

func TestByteWindowRawPayloadTerminatorAndTruncatedLookahead(t *testing.T) {
	// A high byte is opaque when skipped as payload, but ends the stream if
	// the raw head increments into it instead of consuming its command.
	w, err := NewByteWindow(ByteWindowConfig{Text: []byte{'!', 200, 'A', 128}, Slots: 1,
		Step: -2, Advance: 30, Crossing: NonNegative, Commands: map[byte]ByteCommand{'!': {Payload: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	w.Step()
	if w.Finished() || string(w.Letters()) != "A" {
		t.Fatal("lookahead interpreted opaque payload as a terminator")
	}
	for i := 0; i < 100 && !w.Finished(); i++ {
		if _, err := w.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if !w.Finished() || w.Cursor() != 1 || string(w.Letters()) != "A" {
		t.Fatal("raw high payload did not end the window")
	}
	// A command value inside a payload may have its own different length. Raw
	// lookahead must finish safely when that interpretation skips past EOF.
	w, err = NewByteWindow(ByteWindowConfig{Text: []byte{'!', '?', 'A', 128}, Slots: 1,
		Step: -2, Advance: 30, Crossing: NonNegative,
		Commands: map[byte]ByteCommand{'!': {Payload: 1}, '?': {Payload: 3}}})
	if err != nil {
		t.Fatal(err)
	}
	w.Step()
	for i := 0; i < 100 && !w.Finished(); i++ {
		_, err = w.Step()
	}
	if !w.Finished() || err == nil {
		t.Fatal("truncated raw head command did not stop safely")
	}
	// Consuming a raw command can skip the normal terminator and arrive exactly
	// at EOF. This is a clean end, rather than an out-of-bounds lookahead.
	w, err = NewByteWindow(ByteWindowConfig{Text: []byte{'!', '?', 'A', 128}, Slots: 1,
		Step: -2, Advance: 30, Crossing: NonNegative,
		Commands: map[byte]ByteCommand{'!': {Payload: 1}, '?': {Payload: 2}}})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100 && !w.Finished(); i++ {
		if _, err := w.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if !w.Finished() || w.Cursor() != 4 {
		t.Fatal("raw command arriving at EOF did not end cleanly")
	}
}

func TestByteWindowCommandsOnlyRunAtHeadAndPauseKeepsPosition(t *testing.T) {
	text := []byte{'A', '|', 35, '>', 128, 255, 'B', 'C', 'D', 128}
	var pauses, profiles int
	w, err := NewByteWindow(ByteWindowConfig{Text: text, Slots: 4, Step: -2, Advance: 22, Crossing: BelowZero,
		Commands: map[byte]ByteCommand{
			'|': {Payload: 1, Apply: func(payload []byte) (int, error) { pauses++; return int(payload[0]) - 32, nil }},
			'>': {Payload: 2, Apply: func(payload []byte) (int, error) {
				profiles++
				if !bytes.Equal(payload, []byte{128, 255}) {
					t.Fatal("payload was interpreted as a terminator")
				}
				return 0, nil
			}},
		}})
	if err != nil {
		t.Fatal(err)
	}
	// The first crossing consumes the pause and profile commands together. The
	// profile's zero pause must keep the pending pause from the first command.
	if ok, err := w.Step(); !ok || err != nil {
		t.Fatal(ok, err)
	}
	if pauses != 1 || profiles != 1 || w.Cursor() != 6 || w.Pause() != 3 || w.Position() != 20 || !w.Finished() {
		t.Fatalf("head command state: pause=%d profile=%d cursor=%d pauseTicks=%d position=%d finished=%v", pauses, profiles, w.Cursor(), w.Pause(), w.Position(), w.Finished())
	}
	// A longer stream allows observing the three stationary paused ticks.
	text = append([]byte{'A', '|', 35, '>', 128, 255}, []byte("BCDEFGHIJKLMNOPQRSTUVWXYZ")...)
	text = append(text, 128)
	w, err = NewByteWindow(ByteWindowConfig{Text: text, Slots: 4, Step: -2, Advance: 22, Crossing: BelowZero,
		Commands: map[byte]ByteCommand{
			'|': {Payload: 1, Apply: func(p []byte) (int, error) { return int(p[0]) - 32, nil }},
			'>': {Payload: 2},
		}})
	if err != nil {
		t.Fatal(err)
	}
	w.Step()
	for remaining := 2; remaining >= 0; remaining-- {
		w.Step()
		if w.Position() != 20 || w.Cursor() != 6 || w.Pause() != remaining || string(w.Letters()) != "BCDE" {
			t.Fatal("pause tick changed position, head or lookahead")
		}
	}
	w.Step()
	if w.Position() != 18 {
		t.Fatal("motion did not resume immediately after the pause")
	}
	// Without a crossing, lookahead skips both commands without executing them.
	calls := 0
	w, err = NewByteWindow(ByteWindowConfig{Text: text, Slots: 4, Step: -2, Advance: 30, Crossing: NonNegative,
		Commands: map[byte]ByteCommand{
			'|': {Payload: 1, Apply: func([]byte) (int, error) { calls++; return 3, nil }},
			'>': {Payload: 2, Apply: func([]byte) (int, error) { calls++; return 0, nil }},
		}})
	if err != nil {
		t.Fatal(err)
	}
	w.Step()
	if calls != 0 || string(w.Letters()) != "ABCD" || w.Cursor() != 0 || w.Pause() != 0 {
		t.Fatal("lookahead executed a command")
	}
}

func TestByteWindowRepeatCopiesInputsAndHasBoundedLookahead(t *testing.T) {
	text := []byte{'A', '!', 255, 'B', 128}
	commands := map[byte]ByteCommand{'!': {Payload: 1}}
	w, err := NewByteWindow(ByteWindowConfig{Text: text, Slots: 65536, Step: 0, Advance: 1, Crossing: BelowZero, Repeat: true, Commands: commands})
	if err != nil {
		t.Fatal(err)
	}
	text[0] = 'Z'
	commands['!'] = ByteCommand{Payload: 0}
	if ok, err := w.Step(); !ok || err != nil || w.Finished() {
		t.Fatal("short repeat did not fill the fixed window", ok, err)
	}
	for i, value := range w.Letters() {
		want := byte('A')
		if i%2 == 1 {
			want = 'B'
		}
		if value != want {
			t.Fatalf("repeat slot %d: %d/%d", i, value, want)
		}
	}
	if allocs := testing.AllocsPerRun(10, func() {
		if _, err := w.Step(); err != nil {
			t.Fatal(err)
		}
	}); allocs != 0 {
		t.Fatalf("repeat allocated %g objects per tick", allocs)
	}
	// Many command tokens between two glyphs are jumped once per slot, rather
	// than scanned anew for every repeated occurrence.
	long := []byte{'A'}
	for i := 0; i < 10000; i++ {
		long = append(long, '!', 255)
	}
	long = append(long, 'B', 128)
	w, err = NewByteWindow(ByteWindowConfig{Text: long, Slots: 65536, Advance: 1, Repeat: true, Commands: map[byte]ByteCommand{'!': {Payload: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	w.Step()
	if string(w.Letters()[:6]) != "ABABAB" {
		t.Fatal("command-heavy repeat lookahead changed glyph order")
	}
}

func TestByteWindowRepeatExecutesHeadCommandsOncePerTraversal(t *testing.T) {
	calls := 0
	w, err := NewByteWindow(ByteWindowConfig{Text: []byte{'A', '!', 'B', 128}, Slots: 5, Position: 127,
		Step: -128, Advance: 128, Crossing: BelowZero, Repeat: true,
		Commands: map[byte]ByteCommand{'!': {Apply: func([]byte) (int, error) { calls++; return 0, nil }}}})
	if err != nil {
		t.Fatal(err)
	}
	for tick := 0; tick < 20; tick++ {
		w.Step()
		if w.Finished() || w.Position() != 127 {
			t.Fatal("wrapping byte clock changed")
		}
	}
	if calls != 10 {
		t.Fatalf("lookahead or head replayed commands: %d/10", calls)
	}
}

func TestByteWindowValidationAndCallbackFailure(t *testing.T) {
	good := ByteWindowConfig{Text: []byte{'A', 128}, Slots: 1, Advance: 1}
	cases := []ByteWindowConfig{
		{}, {Text: []byte{'A'}, Slots: 1, Advance: 1},
		{Text: []byte{'A', 128}, Slots: 65537, Advance: 1},
		{Text: []byte{'A', 128}, Slots: 1},
		{Text: []byte{'A', 128}, Slots: 1, Advance: 1, Crossing: 2},
		{Text: []byte{'A', '!', 128}, Slots: 1, Advance: 1, Commands: map[byte]ByteCommand{'!': {Payload: 2}}},
		{Text: []byte{'A', '!', 128}, Slots: 1, Advance: 1, Commands: map[byte]ByteCommand{'!': {Payload: 1}}},
		{Text: []byte{'A', 128}, Slots: 1, Advance: 1, Commands: map[byte]ByteCommand{128: {}}},
		{Text: []byte{'A', 128}, Slots: 1, Advance: 1, Commands: map[byte]ByteCommand{'!': {Payload: -1}}},
		{Text: []byte{'A', 128}, Slots: 1, Advance: 1, Cursor: -1},
		{Text: []byte{'A', '!', 1, 128}, Slots: 1, Advance: 1, Cursor: 2, Commands: map[byte]ByteCommand{'!': {Payload: 1}}},
		{Text: []byte{128}, Slots: 1, Advance: 1, Repeat: true},
	}
	for i, c := range cases {
		if _, err := NewByteWindow(c); err == nil {
			t.Fatalf("invalid case %d accepted", i)
		}
	}
	if _, err := NewByteWindow(good); err != nil {
		t.Fatal(err)
	}
	// The terminator and initial offset are configurable independently of ASCII.
	w, err := NewByteWindow(ByteWindowConfig{Text: []byte{1, 2, 3, 7}, Slots: 2, Cursor: 1, End: 7, Advance: 1, Crossing: BelowZero})
	if err != nil {
		t.Fatal(err)
	}
	w.Step()
	if !bytes.Equal(w.Letters(), []byte{2, 3}) {
		t.Fatal("custom byte alphabet or cursor was ignored")
	}
	for _, fail := range []func([]byte) (int, error){
		func([]byte) (int, error) { return 0, errors.New("bad command") },
		func([]byte) (int, error) { return -1, nil },
	} {
		w, err = NewByteWindow(ByteWindowConfig{Text: []byte{'A', '!', 'B', 128}, Slots: 1, Step: -1, Advance: 1, Crossing: BelowZero,
			Commands: map[byte]ByteCommand{'!': {Apply: fail}}})
		if err != nil {
			t.Fatal(err)
		}
		if ok, err := w.Step(); ok || err == nil || !w.Finished() {
			t.Fatal("callback failure did not stop transport")
		}
		if ok, err := w.Step(); ok || err != nil {
			t.Fatal("failed window was stepped again")
		}
	}
}
