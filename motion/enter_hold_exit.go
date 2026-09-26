package motion

import (
	"fmt"
	"math"
)

type SlideStage uint8

const (
	SlideEntering SlideStage = iota
	SlideHolding
	SlideLeaving
	SlideFollowing
)

type SlideEventKind uint8

const (
	SlideContent SlideEventKind = iota
	SlideFollowingContent
)

type SlideEvent struct {
	Kind SlideEventKind
	X    float64
}

type EnterHoldExitConfig struct {
	StartX, EnterVelocity, EnterBoundary float64
	ExitVelocity, ExitBoundary           float64
	EnterInclusive, ExitInclusive        bool
	HoldTicks                            int
}

type EnterHoldExitState struct {
	Stage SlideStage
	X     float64
	Wait  int
}

// EnterHoldExit keeps independent stage checks in one Step. A boundary tick
// may emit two content poses or emit the last content and first following pose
// on the same tick, preserving authored wave and text clocks.
type EnterHoldExit struct {
	config EnterHoldExitConfig
	state  EnterHoldExitState
	events [4]SlideEvent
	count  int
}

func NewEnterHoldExit(c EnterHoldExitConfig) (*EnterHoldExit, error) {
	for _, value := range [...]float64{c.StartX, c.EnterVelocity, c.EnterBoundary,
		c.ExitVelocity, c.ExitBoundary} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite enter-hold-exit parameter")
		}
	}
	if c.EnterVelocity <= 0 || c.ExitVelocity >= 0 || c.EnterBoundary <= c.StartX ||
		c.ExitBoundary >= c.EnterBoundary || c.HoldTicks < 1 {
		return nil, fmt.Errorf("motion: invalid enter-hold-exit path")
	}
	program := &EnterHoldExit{config: c}
	program.Reset()
	return program, nil
}

func (p *EnterHoldExit) Reset() {
	p.state = EnterHoldExitState{X: p.config.StartX, Wait: p.config.HoldTicks}
	p.count = 0
}

func (p *EnterHoldExit) emit(kind SlideEventKind) {
	p.events[p.count] = SlideEvent{Kind: kind, X: p.state.X}
	p.count++
}

// Step returns a borrowed ordered event slice valid until the next Step.
func (p *EnterHoldExit) Step() []SlideEvent {
	p.count = 0
	if p.state.Stage == SlideEntering {
		p.state.X += p.config.EnterVelocity
		if p.state.X > p.config.EnterBoundary || p.config.EnterInclusive && p.state.X == p.config.EnterBoundary {
			p.state.Stage = SlideHolding
		}
		p.emit(SlideContent)
	}
	if p.state.Stage == SlideHolding {
		p.state.Wait--
		if p.state.Wait <= 0 {
			p.state.Stage = SlideLeaving
		}
		p.emit(SlideContent)
	}
	if p.state.Stage == SlideLeaving {
		p.state.X += p.config.ExitVelocity
		if p.state.X < p.config.ExitBoundary || p.config.ExitInclusive && p.state.X == p.config.ExitBoundary {
			p.state.Stage = SlideFollowing
		}
		p.emit(SlideContent)
	}
	if p.state.Stage == SlideFollowing {
		p.emit(SlideFollowingContent)
	}
	return p.events[:p.count]
}

func (p *EnterHoldExit) State() EnterHoldExitState { return p.state }
