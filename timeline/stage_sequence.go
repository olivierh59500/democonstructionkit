package timeline

import "fmt"

// StageSequenceConfig combines immediate event-driven stages with one stage's
// timed windows. A zero-frame fade leaves that window fully opaque; other
// fades advance after sampling, so their first visible frame starts at zero.
type StageSequenceConfig struct {
	Stages     EventStagesConfig
	TimedStage string
	Rate       int
	Windows    []CueRange
	Fades      []HoldRampConfig
}

// StageSequence owns its stage, window clock and per-window fade envelopes.
// Trigger can expose a later stage in the same update; Window and drawing do
// not advance the clock. Scene assets and their draw order stay caller-owned.
type StageSequence struct {
	stages     *EventStages
	timedStage string
	windows    *CueRanges
	clock      *CueClock
	fades      []*HoldRamp
}

func NewStageSequence(config StageSequenceConfig) (*StageSequence, error) {
	if config.TimedStage == "" || len(config.Windows) == 0 || len(config.Fades) != len(config.Windows) {
		return nil, fmt.Errorf("timeline: invalid stage sequence windows")
	}
	found := false
	for _, name := range config.Stages.Stages {
		found = found || name == config.TimedStage
	}
	if !found {
		return nil, fmt.Errorf("timeline: timed stage is not declared")
	}
	stages, err := NewEventStages(config.Stages)
	if err != nil {
		return nil, err
	}
	windows, err := NewCueRanges(config.Windows)
	if err != nil {
		return nil, err
	}
	clock, err := NewCueClock(CueClockConfig{Rate: config.Rate})
	if err != nil {
		return nil, err
	}
	sequence := &StageSequence{stages: stages, timedStage: config.TimedStage, windows: windows, clock: clock, fades: make([]*HoldRamp, len(config.Fades))}
	for i, fade := range config.Fades {
		if fade.Frames == 0 && fade.InteriorOffset == 0 {
			continue
		}
		sequence.fades[i], err = NewHoldRamp(fade)
		if err != nil {
			return nil, err
		}
	}
	return sequence, nil
}

func (sequence *StageSequence) Stage() string             { return sequence.stages.Stage() }
func (sequence *StageSequence) Active(name string) bool   { return sequence.stages.Active(name) }
func (sequence *StageSequence) Trigger(event string) bool { return sequence.stages.Trigger(event) }
func (sequence *StageSequence) Tick() int                 { return sequence.clock.Tick() }

// Window returns the current window index and prepared alpha; a gap or another
// stage returns index -1. Fades are read before StepWindow for exact entry.
func (sequence *StageSequence) Window() (index int, alpha float64) {
	if !sequence.stages.Active(sequence.timedStage) {
		return -1, 0
	}
	index, _, active := sequence.windows.At(float64(sequence.clock.Tick()))
	if !active {
		return -1, 0
	}
	if fade := sequence.fades[index]; fade != nil {
		return index, fade.Progress()
	}
	return index, 1
}

// StepWindow advances the current window's fade and tick after drawing it.
func (sequence *StageSequence) StepWindow() {
	if !sequence.stages.Active(sequence.timedStage) {
		return
	}
	index, _, active := sequence.windows.At(float64(sequence.clock.Tick()))
	if active && sequence.fades[index] != nil {
		sequence.fades[index].Step()
	}
	sequence.clock.Step()
}

func (sequence *StageSequence) Reset() {
	sequence.stages.Reset()
	sequence.clock.Reset()
	for _, fade := range sequence.fades {
		if fade != nil {
			fade.Reset()
		}
	}
}
