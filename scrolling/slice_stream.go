package scrolling

import (
	"fmt"
	"math"
)

// SliceToken is either a glyph with an integer pixel advance or a control event.
// Glyph indexes the caller's filmstrip. A negative glyph emits transparent slots.
// Width need not equal the bitmap width: additional columns provide spacing.
type SliceToken struct {
	Glyph, Width int
	Control      string
}

// SliceStreamConfig describes a bounded strip-history transport. Each ordinary
// Step moves one SliceWidth-wide slot; a partial final glyph strip is retained.
// Use SliceWidth=1 for exact odd/proportional pixel advances. Coarser strips
// quantize advances to ceil(Width/SliceWidth)*SliceWidth, matching strip effects.
// LoopStart is a token index after any one-time prefix, not a byte offset.
type SliceStreamConfig struct {
	Tokens               []SliceToken
	Capacity, SliceWidth int
	Repeat               bool
	LoopStart            int
	Initial              DNASlice
}

// SliceControl is emitted when transport reaches a control token. Controls
// consume one requested step without shifting history. Scene choreography
// (pause, music cue, rotation reversal, palette change) belongs to the callback.
type SliceControl struct {
	Token int
	Name  string
}

// SliceStream owns a fixed-size ring of source strips. Transport and animation
// frames are independent: paused text can continue twisting, for example.
// Methods mutate only this stream and allocate nothing after construction.
type SliceStream struct {
	config             SliceStreamConfig
	slices             []DNASlice
	head, token, slice int
	ended              bool
}

func NewSliceStream(c SliceStreamConfig) (*SliceStream, error) {
	if c.Capacity < 1 || c.Capacity > 1<<20 || c.SliceWidth < 1 || c.SliceWidth > 1<<20 || c.LoopStart < 0 || c.LoopStart >= len(c.Tokens) {
		return nil, fmt.Errorf("scrolling: invalid strip stream dimensions or loop start")
	}
	for _, token := range c.Tokens {
		if token.Control == "" && (token.Width < 1 || token.Width > 1<<24 || token.Glyph < -1) {
			return nil, fmt.Errorf("scrolling: invalid strip glyph")
		}
		if token.Control != "" && token.Width != 0 {
			return nil, fmt.Errorf("scrolling: control token must not consume glyph width")
		}
	}
	c.Tokens = append([]SliceToken(nil), c.Tokens...)
	s := &SliceStream{config: c, slices: make([]DNASlice, c.Capacity)}
	s.Reset()
	return s, nil
}

// Reset restores the initial history and the start of the one-time prefix.
func (s *SliceStream) Reset() {
	s.head = 0
	s.token = 0
	s.slice = 0
	s.ended = false
	for i := range s.slices {
		s.slices[i] = s.config.Initial
	}
}
func (s *SliceStream) Head() int { return s.head }

// Slices borrows the physical ring. Pass it with Head to DNAFrames.DrawSlices;
// do not modify or retain it across concurrent stream updates.
func (s *SliceStream) Slices() []DNASlice { return s.slices }

// Cursor returns the next token index and strip index within that glyph.
func (s *SliceStream) Cursor() (token, slice int) { return s.token, s.slice }
func (s *SliceStream) Ended() bool                { return s.ended }

// Step advances at most count slots and returns the number actually shifted.
// Returning true from onControl stops the remaining budget after that event.
// A nil callback consumes controls without changing animation properties.
func (s *SliceStream) Step(count int, onControl func(SliceControl) bool) int {
	shifted := 0
	for step := 0; step < count && !s.ended; step++ {
		token := s.config.Tokens[s.token]
		if token.Control != "" {
			event := SliceControl{Token: s.token, Name: token.Control}
			s.nextToken()
			if onControl != nil && onControl(event) {
				break
			}
			continue
		}
		s.head++
		if s.head == len(s.slices) {
			s.head = 0
		}
		previous := (s.head + len(s.slices) - 2) % len(s.slices)
		tail := (s.head + len(s.slices) - 1) % len(s.slices)
		s.slices[tail] = DNASlice{Glyph: token.Glyph, Frame: s.slices[previous].Frame, Slice: s.slice}
		shifted++
		s.slice++
		if s.slice >= (token.Width+s.config.SliceWidth-1)/s.config.SliceWidth {
			s.slice = 0
			s.nextToken()
		}
	}
	return shifted
}
func (s *SliceStream) nextToken() {
	s.token++
	if s.token >= len(s.config.Tokens) {
		if s.config.Repeat {
			s.token = s.config.LoopStart
		} else {
			s.ended = true
			s.token = len(s.config.Tokens)
		}
	}
}

// SetFrames samples rotation plus a per-visible-slot phase profile, independent
// of transport. Empty offsets selects a flat profile. Frame indices wrap before
// truncation; existing small-range profiles retain their original arithmetic.
func (s *SliceStream) SetFrames(rotation float64, offsets []float64, frames int) error {
	if frames < 1 || !finite(rotation) || (len(offsets) != 0 && len(offsets) != len(s.slices)) {
		return fmt.Errorf("scrolling: invalid strip animation profile")
	}
	for _, offset := range offsets {
		if !finite(offset) || !finite(rotation+offset) {
			return fmt.Errorf("scrolling: nonfinite strip animation phase")
		}
	}
	period := float64(frames)
	for i := range s.slices {
		frame := rotation
		if len(offsets) != 0 {
			frame += offsets[i]
		}
		if frame >= period {
			frame -= period
		} else if frame < 0 {
			frame += period
		}
		if frame < 0 || frame >= period {
			frame = math.Mod(frame, period)
			if frame < 0 {
				frame += period
			}
		}
		index := s.head + i
		if index >= len(s.slices) {
			index -= len(s.slices)
		}
		s.slices[index].Frame = int(frame)
	}
	return nil
}
