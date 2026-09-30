package timeline

import "fmt"

type WordCue struct {
	Ticks int
	Delta []int16
}
type WordProgramConfig struct {
	Initial               []int16
	Cues                  []WordCue
	Repeat, ResetOnRepeat bool
}

// WordProgram owns a bounded sequence of wrapping signed-word increments.
// The first Step applies the first cue; a boundary applies the next cue on that
// same update. Drawing and state inspection never advance the program.
type WordProgram struct {
	initial, state               []int16
	cues                         []WordCue
	cue, within                  int
	started, done, repeat, reset bool
}

func NewWordProgram(c WordProgramConfig) (*WordProgram, error) {
	if len(c.Initial) < 1 || len(c.Initial) > 1024 || len(c.Cues) < 1 || len(c.Cues) > 65536 || len(c.Cues) > 1<<20/len(c.Initial) {
		return nil, fmt.Errorf("timeline: invalid word program dimensions")
	}
	for _, cue := range c.Cues {
		if cue.Ticks < 1 || cue.Ticks > 1<<30 || len(cue.Delta) != len(c.Initial) {
			return nil, fmt.Errorf("timeline: invalid word cue")
		}
	}
	p := &WordProgram{initial: append([]int16(nil), c.Initial...), state: append([]int16(nil), c.Initial...), cues: make([]WordCue, len(c.Cues)), repeat: c.Repeat, reset: c.ResetOnRepeat}
	for i, cue := range c.Cues {
		p.cues[i] = WordCue{Ticks: cue.Ticks, Delta: append([]int16(nil), cue.Delta...)}
	}
	return p, nil
}
func (p *WordProgram) Step() bool {
	if p == nil || p.done {
		return false
	}
	if !p.started {
		p.started = true
	} else {
		p.within++
		if p.within >= p.cues[p.cue].Ticks {
			p.cue++
			p.within = 0
		}
	}
	if p.cue >= len(p.cues) {
		if !p.repeat {
			p.done = true
			return false
		}
		p.cue = 0
		if p.reset {
			copy(p.state, p.initial)
		}
	}
	for i, delta := range p.cues[p.cue].Delta {
		p.state[i] = int16(uint16(p.state[i]) + uint16(delta))
	}
	return true
}
func (p *WordProgram) State() []int16 {
	if p == nil {
		return nil
	}
	return p.state
}
func (p *WordProgram) Finished() bool { return p == nil || p.done }
func (p *WordProgram) Position() (cue, tick int) {
	if p == nil {
		return 0, 0
	}
	return p.cue, p.within
}
func (p *WordProgram) Reset() {
	if p != nil {
		copy(p.state, p.initial)
		p.cue, p.within = 0, 0
		p.started, p.done = false, false
	}
}
