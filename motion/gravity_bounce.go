package motion

import (
	"fmt"
	"math"
)

// GravityBounceConfig describes a one-dimensional accelerating image path.
// When Position crosses Floor, Velocity becomes the current Rebound and the
// next Rebound is multiplied by Damping. The position is deliberately not
// clamped, preserving the overshoot of authored sprite animations.
type GravityBounceConfig struct {
	StartPosition, StartVelocity, StartRebound float64
	Gravity, Floor, Damping                    float64
	StopRebound                                float64
	StopWhenAtMost                             bool // Default completion is rebound >= StopRebound.
}

type GravityBounceState struct {
	Position, Velocity, Rebound float64
	Finished                    bool
}

// GravityBounce advances only when its host scene calls Step. Completion holds
// the final pose, so an outro or timeline can take over without an extra jump.
type GravityBounce struct {
	config GravityBounceConfig
	state  GravityBounceState
}

func NewGravityBounce(c GravityBounceConfig) (*GravityBounce, error) {
	for _, value := range [...]float64{c.StartPosition, c.StartVelocity, c.StartRebound,
		c.Gravity, c.Floor, c.Damping, c.StopRebound} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite gravity bounce parameter")
		}
	}
	b := &GravityBounce{config: c}
	b.Reset()
	return b, nil
}

func (b *GravityBounce) Reset() {
	b.state = GravityBounceState{Position: b.config.StartPosition,
		Velocity: b.config.StartVelocity, Rebound: b.config.StartRebound}
}

// Step returns true on and after the terminal rebound threshold is reached.
func (b *GravityBounce) Step() bool {
	if b.state.Finished {
		return true
	}
	b.state.Velocity += b.config.Gravity
	b.state.Position += b.state.Velocity
	if b.state.Position > b.config.Floor {
		b.state.Velocity = b.state.Rebound
		b.state.Rebound *= b.config.Damping
	}
	if b.config.StopWhenAtMost {
		b.state.Finished = b.state.Rebound <= b.config.StopRebound
	} else {
		b.state.Finished = b.state.Rebound >= b.config.StopRebound
	}
	return b.state.Finished
}

func (b *GravityBounce) State() GravityBounceState { return b.state }
func (b *GravityBounce) Position() float64         { return b.state.Position }
