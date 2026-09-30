package scrolltext

import "fmt"

// ByteCrossing selects which signed half of a wrapping byte advances the text.
// This keeps byte-sized clocks independent of glyph dimensions and rendering.
type ByteCrossing uint8

const (
	BelowZero ByteCrossing = iota
	NonNegative
)

// ByteCommand consumes a command byte followed by Payload bytes. Apply runs only
// when the command reaches the head, never when filling the visible lookahead.
// Its payload is borrowed and must not be changed or retained. A positive pause
// replaces the current pause; zero leaves a pause from an earlier command intact.
type ByteCommand struct {
	Payload int
	Apply   func([]byte) (pauseTicks int, err error)
}

// ByteWindowConfig describes a bounded byte stream and a fixed glyph window.
// Text includes a terminator at or above End (128 by default). Command payloads
// are opaque and may contain bytes at or above End. Cursor is a byte offset to a
// glyph, command or the terminator. Text and the command table are copied at
// construction. Moving the head increments the raw byte offset before consuming
// commands. Thus an initial command is skipped by lookahead, but its payload may
// later be visited as ordinary text; the constructor never executes commands.
type ByteWindowConfig struct {
	Text     []byte
	Slots    int
	Position byte
	Step     int8
	Advance  byte
	Crossing ByteCrossing
	Cursor   int
	End      byte
	Repeat   bool
	Commands map[byte]ByteCommand
}

// ByteWindow owns byte transport, head commands, pauses and glyph lookahead.
// Geometry and palette clocks remain independent. After construction, Step does
// not allocate, apart from errors or allocations made by a command callback.
type ByteWindow struct {
	text     []byte
	commands map[byte]ByteCommand
	next     []int
	letters  []byte
	position byte
	step     int8
	advance  byte
	crossing ByteCrossing
	cursor   int
	terminal int
	end      byte
	pause    int
	repeat   bool
	finished bool
}

func NewByteWindow(c ByteWindowConfig) (*ByteWindow, error) {
	if c.Slots < 1 || c.Slots > 65536 || len(c.Text) == 0 || len(c.Text) > 1<<20 || c.Advance == 0 || c.Crossing > NonNegative {
		return nil, fmt.Errorf("scrolltext: invalid byte window dimensions or transport")
	}
	end := c.End
	if end == 0 {
		end = 128
	}
	commands := make(map[byte]ByteCommand, len(c.Commands))
	for value, command := range c.Commands {
		if value >= end || command.Payload < 0 || command.Payload > len(c.Text)-1 {
			return nil, fmt.Errorf("scrolltext: invalid byte command %d", value)
		}
		commands[value] = command
	}
	terminal, glyphs := -1, 0
	validCursor := false
	for index := 0; index < len(c.Text); {
		value := c.Text[index]
		if value >= end {
			terminal = index
			validCursor = validCursor || index == c.Cursor
			break
		}
		if command, ok := commands[value]; ok {
			if command.Payload > len(c.Text)-index-1 {
				return nil, fmt.Errorf("scrolltext: truncated byte command %d at %d", value, index)
			}
			validCursor = validCursor || index == c.Cursor
			index += 1 + command.Payload
			continue
		}
		validCursor = validCursor || index == c.Cursor
		glyphs++
		index++
	}
	if terminal < 0 || !validCursor || (c.Repeat && glyphs == 0) {
		return nil, fmt.Errorf("scrolltext: missing terminator, invalid token cursor or empty repeating byte stream")
	}
	w := &ByteWindow{text: append([]byte(nil), c.Text[:terminal+1]...), commands: commands,
		next: make([]int, terminal+1), letters: make([]byte, c.Slots), position: c.Position,
		step: c.Step, advance: c.Advance, crossing: c.Crossing, cursor: c.Cursor, terminal: terminal, end: end, repeat: c.Repeat}
	// Compute lookahead at every raw byte offset, including payload bytes that a
	// byte cursor can subsequently visit. Reaching a high byte ends lookahead;
	// high bytes are opaque only while an actual command skips its payload.
	for index := terminal; index >= 0; index-- {
		w.next[index] = index
		if w.text[index] >= end {
			continue
		}
		if command, ok := commands[w.text[index]]; ok {
			destination := index + 1 + command.Payload
			if destination > terminal {
				// A command found inside another payload may escape the final
				// terminator. Bounded lookahead then ends as if it reached EOF.
				w.next[index] = terminal
			} else {
				w.next[index] = w.next[destination]
			}
		}
	}
	return w, nil
}

// Step advances one clock tick. A tick that encounters the end returns true and
// marks Finished; subsequent calls return false. Slots already written on that
// final tick retain their values, matching ordinary sequential window filling.
func (w *ByteWindow) Step() (bool, error) {
	if w == nil || w.finished {
		return false, nil
	}
	if w.pause > 0 {
		w.pause--
	} else {
		w.position += byte(w.step)
		crossed := int8(w.position) < 0
		if w.crossing == NonNegative {
			crossed = !crossed
		}
		if crossed {
			w.position += w.advance
			if w.cursor < w.terminal {
				w.cursor++
			}
			if err := w.consumeCommands(); err != nil {
				w.finished = true
				return false, err
			}
		}
	}
	index := w.cursor
	for slot := range w.letters {
		if index < len(w.next) {
			index = w.next[index]
		}
		if index >= len(w.text) || w.text[index] >= w.end {
			if !w.repeat {
				w.finished = true
				return true, nil
			}
			index = w.next[0]
		}
		w.letters[slot] = w.text[index]
		index++
	}
	return true, nil
}

func (w *ByteWindow) consumeCommands() error {
	for consumed := 0; consumed <= len(w.text); consumed++ {
		if w.cursor >= len(w.text) || w.text[w.cursor] >= w.end {
			if !w.repeat {
				return nil
			}
			w.cursor = 0
		}
		command, ok := w.commands[w.text[w.cursor]]
		if !ok {
			return nil
		}
		if command.Payload > len(w.text)-w.cursor-1 {
			return fmt.Errorf("scrolltext: truncated raw byte command at %d", w.cursor)
		}
		payload := w.text[w.cursor+1 : w.cursor+1+command.Payload]
		if command.Apply != nil {
			pause, err := command.Apply(payload)
			if err != nil {
				return fmt.Errorf("scrolltext: byte command %d at %d: %w", w.text[w.cursor], w.cursor, err)
			}
			if pause < 0 {
				return fmt.Errorf("scrolltext: byte command returned a negative pause")
			}
			if pause > 0 {
				w.pause = pause
			}
		}
		w.cursor += 1 + command.Payload
	}
	return fmt.Errorf("scrolltext: byte command loop exceeded stream bounds")
}

// Letters returns the borrowed fixed-size glyph window; callers must not mutate it.
func (w *ByteWindow) Letters() []byte { return w.letters }
func (w *ByteWindow) Position() byte  { return w.position }
func (w *ByteWindow) Cursor() int     { return w.cursor }
func (w *ByteWindow) Pause() int      { return w.pause }
func (w *ByteWindow) Finished() bool  { return w.finished }
