package timeline

import (
	"fmt"
	"math"
)

// IntroCue selects when a one-shot soundtrack or scene event becomes ready.
type IntroCue uint8

const (
	IntroCueNone IntroCue = iota
	IntroCueOnEntry
	IntroCueOnFirstMainTick
	IntroCueAboveFade
)

// IntroHandoffConfig describes a finite intro-to-main transition. FadeStart is
// applied on the entry tick; FadeStep is added on later main ticks and clamped
// at FadeMax. CueThreshold is strict by default, matching scenes that start
// music when fade > threshold; CueInclusive also accepts equality.
type IntroHandoffConfig struct {
	FadeStart, FadeStep, FadeMax float64
	CueThreshold                 float64
	Cue                          IntroCue
	CueInclusive                 bool
}

// IntroHandoff owns transition, fade and one-shot cue state without knowing
// about graphics or audio. Draw reads Main and Fade but never advances them.
type IntroHandoff struct {
	config      IntroHandoffConfig
	main        bool
	justEntered bool
	mainTicks   uint64
	fade        float64
	cueFired    bool
}

func NewIntroHandoff(config IntroHandoffConfig) (*IntroHandoff, error) {
	for _, value := range [...]float64{config.FadeStart, config.FadeStep, config.FadeMax, config.CueThreshold} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("timeline: nonfinite intro handoff setting")
		}
	}
	if config.FadeStart < 0 || config.FadeStep < 0 || config.FadeMax < config.FadeStart || config.Cue > IntroCueAboveFade {
		return nil, fmt.Errorf("timeline: invalid intro handoff setting")
	}
	return &IntroHandoff{config: config, fade: config.FadeStart}, nil
}

// Step samples the intro's completion after its own update. Entry consumes one
// tick without advancing the main fade; subsequent calls advance it once.
func (handoff *IntroHandoff) Step(introFinished bool) {
	handoff.justEntered = false
	if !handoff.main {
		if introFinished {
			handoff.main = true
			handoff.justEntered = true
			handoff.fade = handoff.config.FadeStart
		}
		return
	}
	if handoff.fade < handoff.config.FadeMax {
		handoff.fade = min(handoff.config.FadeMax, handoff.fade+handoff.config.FadeStep)
	}
	handoff.mainTicks++
}

func (handoff *IntroHandoff) Main() bool        { return handoff.main }
func (handoff *IntroHandoff) JustEntered() bool { return handoff.justEntered }
func (handoff *IntroHandoff) Fade() float64     { return handoff.fade }

// CueReady stays true after the fade threshold until MarkCue is called. Entry
// and first-main-tick cues are available only on their respective boundary.
func (handoff *IntroHandoff) CueReady() bool {
	if !handoff.main || handoff.cueFired {
		return false
	}
	switch handoff.config.Cue {
	case IntroCueOnEntry:
		return handoff.justEntered
	case IntroCueOnFirstMainTick:
		return handoff.mainTicks == 1
	case IntroCueAboveFade:
		return handoff.fade > handoff.config.CueThreshold || handoff.config.CueInclusive && handoff.fade == handoff.config.CueThreshold
	default:
		return false
	}
}

// MarkCue acknowledges an event after the production has handled it.
func (handoff *IntroHandoff) MarkCue() { handoff.cueFired = true }

// Reset restores the intro phase and the configured fade start.
func (handoff *IntroHandoff) Reset() {
	handoff.main, handoff.justEntered, handoff.cueFired = false, false, false
	handoff.mainTicks = 0
	handoff.fade = handoff.config.FadeStart
}
