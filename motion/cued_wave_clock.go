package motion

import (
	"fmt"
	"math"
)

// CuedWaveMode optionally changes a clock step, stop state and two sticky
// visibility flags. A mode with SetStep=false retains its previous speed until
// an authored phase landing or later cue changes it.
type CuedWaveMode struct {
	SetStep               bool
	Step                  float64
	SetStop               bool
	Stop                  bool
	ShowSecond, ShowThird bool
}

type CuedWaveClockConfig struct {
	Clock         WaveClockConfig
	Modes         []CuedWaveMode
	Controls      map[rune]int
	InitialMode   int
	LandingAt     float64
	HaltAtLanding bool
}

type CuedWaveState struct {
	Mode                  int
	Step                  float64
	Stop                  bool
	ShowSecond, ShowThird bool
	Phase                 float64
}

// CuedWaveClock binds lookahead text controls to one wave phase. At reads the
// current pose; Step applies the next rune afterward, including a landing reset
// and an optional stop only at that reset boundary.
type CuedWaveClock struct {
	config CuedWaveClockConfig
	clock  *WaveClock
	state  CuedWaveState
}

func NewCuedWaveClock(c CuedWaveClockConfig) (*CuedWaveClock, error) {
	if len(c.Modes) == 0 || len(c.Modes) > 256 ||
		c.InitialMode < 0 || c.InitialMode >= len(c.Modes) ||
		math.IsNaN(c.LandingAt) || math.IsInf(c.LandingAt, 0) || c.LandingAt <= 0 {
		return nil, fmt.Errorf("motion: invalid cued wave modes or landing")
	}
	for _, mode := range c.Modes {
		if math.IsNaN(mode.Step) || math.IsInf(mode.Step, 0) {
			return nil, fmt.Errorf("motion: nonfinite cued wave speed")
		}
	}
	controls := make(map[rune]int, len(c.Controls))
	for ch, mode := range c.Controls {
		if mode < 0 || mode >= len(c.Modes) {
			return nil, fmt.Errorf("motion: invalid cued wave control %q", ch)
		}
		controls[ch] = mode
	}
	c.Controls = controls
	c.Modes = append([]CuedWaveMode(nil), c.Modes...)
	clock, err := NewWaveClock(c.Clock)
	if err != nil {
		return nil, err
	}
	p := &CuedWaveClock{config: c, clock: clock}
	p.Reset()
	return p, nil
}

func (p *CuedWaveClock) Reset() {
	p.clock.Reset()
	p.state = CuedWaveState{Mode: p.config.InitialMode, Step: p.config.Clock.Step,
		Phase: p.clock.Phase()}
}

func (p *CuedWaveClock) At() float64 { return p.clock.At(0) }

// Step applies a peeked rune after the current image has already been drawn.
func (p *CuedWaveClock) Step(next rune) error {
	if mode, ok := p.config.Controls[next]; ok {
		p.state.Mode = mode
	}
	action := p.config.Modes[p.state.Mode]
	if action.ShowSecond {
		p.state.ShowSecond = true
	}
	if action.ShowThird {
		p.state.ShowThird = true
	}
	if action.SetStep {
		p.state.Step = action.Step
	}
	if action.SetStop {
		p.state.Stop = action.Stop
	}
	if p.clock.Phase() >= p.config.LandingAt {
		p.clock.Reset()
		if p.config.HaltAtLanding && p.state.Stop {
			p.state.Step = 0
		}
	}
	if err := p.clock.SetStep(p.state.Step); err != nil {
		return err
	}
	p.clock.Step()
	p.state.Phase = p.clock.Phase()
	return nil
}

func (p *CuedWaveClock) State() CuedWaveState { return p.state }
