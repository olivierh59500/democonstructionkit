package motion

import (
	"fmt"
	"math"
)

// ScrollCue changes the text or animation clock when a control token is read.
// PauseTicks holds text insertion while rotation continues. StopBudget asks a
// strip transport to stop the remaining insertion budget after this token.
type ScrollCue struct {
	PauseTicks   int
	SetRotation  bool
	RotationStep float64
	SetTextStep  bool
	TextStep     int
	StopBudget   bool
}

type CuedScrollClockConfig struct {
	InitialRotation, InitialRotationStep float64
	InitialTextStep, InitialPauseTicks   int
	ResumeRotationStep                   float64
	ResumeTextStep                       int
	RotationFrames                       int
	Cues                                 map[string]ScrollCue
}

type CuedScrollState struct {
	Rotation, RotationStep float64
	TextStep, PauseTicks   int
	Paused                 bool
}

// CuedScrollClock keeps text pauses and animation rotation independent. Step
// invokes insertion before rotating, so a control reached on this tick affects
// the same tick's film frame. A pause expiry restores configured speeds before
// rotating but does not insert another text slice on that expiry tick.
type CuedScrollClock struct {
	config CuedScrollClockConfig
	state  CuedScrollState
}

func NewCuedScrollClock(c CuedScrollClockConfig) (*CuedScrollClock, error) {
	if c.RotationFrames < 1 || c.InitialTextStep < 0 || c.InitialPauseTicks < 0 || c.ResumeTextStep < 0 ||
		!finiteCuedClock(c.InitialRotation) || !finiteCuedClock(c.InitialRotationStep) ||
		!finiteCuedClock(c.ResumeRotationStep) {
		return nil, fmt.Errorf("motion: invalid cued scroll clock")
	}
	copyCues := make(map[string]ScrollCue, len(c.Cues))
	for name, cue := range c.Cues {
		if name == "" || cue.PauseTicks < 0 || cue.TextStep < 0 || !finiteCuedClock(cue.RotationStep) {
			return nil, fmt.Errorf("motion: invalid scroll cue %q", name)
		}
		copyCues[name] = cue
	}
	c.Cues = copyCues
	clock := &CuedScrollClock{config: c}
	clock.Reset()
	return clock, nil
}

func finiteCuedClock(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func (c *CuedScrollClock) Reset() {
	c.state = CuedScrollState{Rotation: c.config.InitialRotation,
		RotationStep: c.config.InitialRotationStep, TextStep: c.config.InitialTextStep,
		PauseTicks: c.config.InitialPauseTicks}
}

// OnControl applies a named cue after its control token has been consumed.
// Unknown controls are inert, so a scene can reserve commands for other layers.
func (c *CuedScrollClock) OnControl(name string) bool {
	cue, ok := c.config.Cues[name]
	if !ok {
		return false
	}
	if cue.PauseTicks > 0 {
		c.state.Paused = true
		c.state.PauseTicks = cue.PauseTicks
	}
	if cue.SetRotation {
		c.state.RotationStep = cue.RotationStep
	}
	if cue.SetTextStep {
		c.state.TextStep = cue.TextStep
	}
	return cue.StopBudget
}

// Step calls insert only when text was not already paused at the start of this
// tick. insert may synchronously call OnControl before rotation advances.
func (c *CuedScrollClock) Step(insert func(count int)) error {
	if !c.state.Paused {
		if insert != nil {
			insert(c.state.TextStep)
		}
	} else {
		c.state.PauseTicks--
		if c.state.PauseTicks == 0 {
			c.state.Paused = false
			c.state.RotationStep = c.config.ResumeRotationStep
			c.state.TextStep = c.config.ResumeTextStep
		}
	}
	return c.AdvanceRotation(c.state.RotationStep)
}

// AdvanceRotation can also be used during an intro pre-roll that inserts text
// regardless of a newly encountered pause command.
func (c *CuedScrollClock) AdvanceRotation(step float64) error {
	if !finiteCuedClock(step) || !finiteCuedClock(c.state.Rotation+step) {
		return fmt.Errorf("motion: invalid cued scroll rotation step")
	}
	c.state.Rotation += step
	period := float64(c.config.RotationFrames)
	if c.state.Rotation >= period {
		c.state.Rotation -= period
	}
	if c.state.Rotation < 0 {
		c.state.Rotation += period
	}
	if c.state.Rotation < 0 || c.state.Rotation >= period {
		c.state.Rotation = math.Mod(c.state.Rotation, period)
		if c.state.Rotation < 0 {
			c.state.Rotation += period
		}
	}
	return nil
}

func (c *CuedScrollClock) State() CuedScrollState { return c.state }
func (c *CuedScrollClock) RotationFrames() int    { return c.config.RotationFrames }
