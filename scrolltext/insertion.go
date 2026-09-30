package scrolltext

import "fmt"

// InsertionToken separates a glyph's layout advance from an opaque scene command.
type InsertionToken struct {
	Rune    rune
	Font    string
	Advance int
	Glyph   bool
	Command byte
	Payload []byte
}
type InsertionProgramConfig struct {
	Tokens                                    []InsertionToken
	Speed, TargetSpeed, SpeedStep, Divisor    int
	Entry, AlignBias, AlignMask, RetireBefore int
	OnCommand                                 func(*InsertionProgram, InsertionToken) error
}

// InsertionState borrows the fixed origin bank; values are absolute transport
// coordinates. First/Fetched select the visible original glyph indices.
type InsertionState struct {
	Cursor, Remaining, Speed, TargetSpeed, Pause, Distance, First, Fetched int
	Origins                                                                []int
	Finished                                                               bool
}
type InsertionProgram struct {
	config InsertionProgramConfig
	state  InsertionState
	tokens []InsertionToken
	glyphs []InsertionToken
	err    error
}

func NewInsertionProgram(c InsertionProgramConfig) (*InsertionProgram, error) {
	if c.Divisor == 0 {
		c.Divisor = 1
	}
	if c.SpeedStep == 0 {
		c.SpeedStep = 1
	}
	if len(c.Tokens) > 65536 || c.Speed < 0 || c.Speed > 1<<20 || c.TargetSpeed < 0 || c.TargetSpeed > 1<<20 || c.SpeedStep < 1 || c.SpeedStep > 1<<20 || c.Divisor < 1 || c.Divisor > 1<<20 || c.Entry < 0 || c.Entry > 1<<20 || c.AlignMask < 0 || c.AlignMask > 65535 || c.AlignBias < -(1<<20) || c.AlignBias > 1<<20 || c.RetireBefore < -(1<<20) || c.RetireBefore > 1<<20 {
		return nil, fmt.Errorf("scrolltext: invalid insertion transport")
	}
	p := &InsertionProgram{config: c, tokens: make([]InsertionToken, len(c.Tokens)), state: InsertionState{Speed: c.Speed, TargetSpeed: c.TargetSpeed}}
	bytes := 0
	for i, t := range c.Tokens {
		if t.Glyph && (t.Advance < 1 || t.Advance > 1<<20) || len(t.Font) > 1024 {
			return nil, fmt.Errorf("scrolltext: invalid insertion glyph")
		}
		bytes += len(t.Payload)
		if bytes > 1<<20 {
			return nil, fmt.Errorf("scrolltext: insertion payload exceeds budget")
		}
		t.Payload = append([]byte(nil), t.Payload...)
		p.tokens[i] = t
		if t.Glyph {
			p.glyphs = append(p.glyphs, t)
		}
	}
	p.config.Tokens = nil
	p.state.Origins = make([]int, len(p.glyphs))
	return p, nil
}
func (p *InsertionProgram) SetTargetSpeed(speed int) error {
	if p == nil || speed < 0 || speed > 1<<20 {
		return fmt.Errorf("scrolltext: invalid insertion target speed")
	}
	p.state.TargetSpeed = speed
	return nil
}
func (p *InsertionProgram) SetPause(ticks int) error {
	if p == nil || ticks < 0 || ticks > 1<<30 {
		return fmt.Errorf("scrolltext: invalid insertion pause")
	}
	p.state.Pause = ticks
	return nil
}
func (p *InsertionProgram) State() InsertionState {
	if p == nil {
		return InsertionState{Finished: true}
	}
	return p.state
}
func (p *InsertionProgram) Glyphs() []InsertionToken {
	if p == nil {
		return nil
	}
	return p.glyphs
}

// Reset restarts the transport without reallocating its borrowed origin bank.
// Scene state changed by OnCommand belongs to the caller and is not reset.
func (p *InsertionProgram) Reset() {
	if p == nil {
		return
	}
	origins := p.state.Origins
	clear(origins)
	p.state = InsertionState{Speed: p.config.Speed, TargetSpeed: p.config.TargetSpeed, Origins: origins}
	p.err = nil
}

// Step applies controls while fetching one glyph; a newly scheduled pause starts
// on the next tick. Speed changes after movement, preserving ramp timing.
func (p *InsertionProgram) Step() (bool, error) {
	if p == nil {
		return false, fmt.Errorf("scrolltext: absent insertion transport")
	}
	if p.err != nil {
		return false, p.err
	}
	s := &p.state
	if s.Finished {
		return false, nil
	}
	if s.Pause > 0 {
		s.Pause--
	} else {
		step := s.Speed / p.config.Divisor
		s.Remaining -= step
		if s.Remaining < 0 {
			inserted := false
			for s.Cursor < len(p.tokens) && !inserted {
				token := p.tokens[s.Cursor]
				s.Cursor++
				if !token.Glyph {
					if p.config.OnCommand != nil {
						if err := p.config.OnCommand(p, token); err != nil {
							p.err = err
							return false, err
						}
					}
					continue
				}
				origin := int64(s.Distance) + int64(p.config.Entry) + int64((s.Remaining+p.config.AlignBias)&p.config.AlignMask)
				if origin > 1<<30 || origin < -(1<<30) {
					p.err = fmt.Errorf("scrolltext: insertion origin exceeds coordinate budget")
					return false, p.err
				}
				s.Origins[s.Fetched] = int(origin)
				s.Fetched++
				s.Remaining += token.Advance
				inserted = true
			}
			if !inserted {
				s.Finished = true
				return false, nil
			}
		}
		if int64(s.Distance)+int64(step) > 1<<30 {
			p.err = fmt.Errorf("scrolltext: insertion distance exceeds coordinate budget")
			return false, p.err
		}
		s.Distance += step
		if s.Speed < s.TargetSpeed {
			s.Speed = min(s.TargetSpeed, s.Speed+p.config.SpeedStep)
		} else if s.Speed > s.TargetSpeed {
			s.Speed = max(s.TargetSpeed, s.Speed-p.config.SpeedStep)
		}
	}
	for s.First < s.Fetched && s.Origins[s.First]-s.Distance < p.config.RetireBefore {
		s.First++
	}
	return true, nil
}
