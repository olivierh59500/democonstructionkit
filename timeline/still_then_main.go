package timeline

import (
	"fmt"
	"math"
)

// StillThenMainConfig defines an image hold followed by a blank hold. The
// first main tick emits one cue; the host chooses its music or scene action.
type StillThenMainConfig struct {
	StillTicks, BlankTicks int
}

type StillThenMainPhase uint8

const (
	StillImage StillThenMainPhase = iota
	StillBlank
	StillMain
)

type StillThenMainFrame struct {
	Phase StillThenMainPhase
	Tick  uint64
	Cue   bool
}

// StillThenMain advances one fixed simulation tick per Next. It owns no
// graphics or audio and has no per-frame allocation.
type StillThenMain struct {
	still, main uint64
	tick        uint64
}

func NewStillThenMain(config StillThenMainConfig) (*StillThenMain, error) {
	if config.StillTicks < 0 || config.BlankTicks < 0 || config.StillTicks > math.MaxInt-config.BlankTicks {
		return nil, fmt.Errorf("timeline: invalid still/blank durations")
	}
	return &StillThenMain{still: uint64(config.StillTicks), main: uint64(config.StillTicks + config.BlankTicks)}, nil
}

func (sequence *StillThenMain) Next() StillThenMainFrame {
	frame := StillThenMainFrame{Tick: sequence.tick}
	switch {
	case sequence.tick < sequence.still:
		frame.Phase = StillImage
	case sequence.tick < sequence.main:
		frame.Phase = StillBlank
	default:
		frame.Phase = StillMain
		frame.Cue = sequence.tick == sequence.main
	}
	if sequence.tick < math.MaxUint64 {
		sequence.tick++
	}
	return frame
}

func (sequence *StillThenMain) Reset()       { sequence.tick = 0 }
func (sequence *StillThenMain) Tick() uint64 { return sequence.tick }
