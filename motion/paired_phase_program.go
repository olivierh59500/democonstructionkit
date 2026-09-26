package motion

import (
	"fmt"
	"math"
)

// PhasePass selects one phase interval and its draw order. Material identifies
// an image pair in the renderer. Intervals do not wrap across Period.
type PhasePass struct {
	From, To               float64
	IncludeFrom, IncludeTo bool
	Reverse                bool
	Material               int
	Alpha                  float64
}

type PairedPhaseConfig struct {
	Count                        int
	Start, Spacing, Step, Period float64
	CenterY, AmplitudeY          float64
	Passes                       []PhasePass
}

type PhasePairSample struct {
	Index, Material int
	Y, Alpha        float64
}

// PairedPhaseProgram selects ordered image-pair samples from the phases before
// advancing them. A host can Update and Draw once per active screen tick.
type PairedPhaseProgram struct {
	config  PairedPhaseConfig
	phases  []float64
	samples []PhasePairSample
	tick    int
}

func NewPairedPhaseProgram(c PairedPhaseConfig) (*PairedPhaseProgram, error) {
	if c.Count < 1 || c.Count > 1<<16 || len(c.Passes) == 0 || len(c.Passes) > 256 ||
		c.Count > (1<<16)/len(c.Passes) ||
		!finitePhase(c.Start) || !finitePhase(c.Spacing) || !finitePhase(c.Step) || !finitePhase(c.Period) ||
		!finitePhase(c.CenterY) || !finitePhase(c.AmplitudeY) || c.Period <= 0 ||
		!finitePhase(c.Start+float64(c.Count-1)*c.Spacing) {
		return nil, fmt.Errorf("motion: invalid paired phase program")
	}
	for _, pass := range c.Passes {
		if !finitePhase(pass.From) || !finitePhase(pass.To) ||
			pass.From < 0 || pass.To > c.Period || pass.From > pass.To ||
			!finitePhase(pass.Alpha) || pass.Alpha < 0 || pass.Alpha > 1 || pass.Material < 0 {
			return nil, fmt.Errorf("motion: invalid paired phase pass")
		}
	}
	c.Passes = append([]PhasePass(nil), c.Passes...)
	p := &PairedPhaseProgram{config: c, phases: make([]float64, c.Count),
		samples: make([]PhasePairSample, 0, c.Count*len(c.Passes))}
	p.Reset()
	return p, nil
}

func finitePhase(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func (p *PairedPhaseProgram) Reset() {
	for i := range p.phases {
		phase := p.config.Start + float64(i)*p.config.Spacing
		if phase < 0 || phase >= p.config.Period {
			phase = math.Mod(phase, p.config.Period)
			if phase < 0 {
				phase += p.config.Period
			}
		}
		p.phases[i] = phase
	}
	p.samples = p.samples[:0]
	p.tick = 0
}

func (p *PairedPhaseProgram) Step() {
	p.samples = p.samples[:0]
	for _, pass := range p.config.Passes {
		if pass.Reverse {
			for i := len(p.phases) - 1; i >= 0; i-- {
				p.sample(pass, i)
			}
		} else {
			for i := range p.phases {
				p.sample(pass, i)
			}
		}
	}
	for i := len(p.phases) - 1; i >= 0; i-- {
		p.phases[i] += p.config.Step
		if p.phases[i] >= p.config.Period {
			p.phases[i] -= p.config.Period
			if p.phases[i] >= p.config.Period {
				p.phases[i] = math.Mod(p.phases[i], p.config.Period)
			}
		}
		if p.phases[i] < 0 {
			p.phases[i] += p.config.Period
			if p.phases[i] < 0 {
				p.phases[i] = math.Mod(p.phases[i], p.config.Period)
				if p.phases[i] < 0 {
					p.phases[i] += p.config.Period
				}
			}
		}
	}
	p.tick++
}

func (p *PairedPhaseProgram) sample(pass PhasePass, index int) {
	phase := p.phases[index]
	if phase < pass.From || phase == pass.From && !pass.IncludeFrom ||
		phase > pass.To || phase == pass.To && !pass.IncludeTo {
		return
	}
	p.samples = append(p.samples, PhasePairSample{Index: index, Material: pass.Material,
		Y: p.config.CenterY + p.config.AmplitudeY*math.Cos(phase), Alpha: pass.Alpha})
}

func (p *PairedPhaseProgram) Samples() []PhasePairSample { return p.samples }
func (p *PairedPhaseProgram) Phases() []float64          { return p.phases }
func (p *PairedPhaseProgram) Tick() int                  { return p.tick }
