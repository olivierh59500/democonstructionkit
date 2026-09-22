package modulation

import (
	"fmt"
	"math"
)

// DecayConfig describes a retriggerable peak followed by linear decay. Use
// seconds and units/second, or delta=1 and units/tick for an exact authored clock.
type DecayConfig struct{ Floor, Peak, Rate float64 }

type Decay struct {
	config DecayConfig
	value  float64
}

func NewDecay(c DecayConfig) (*Decay, error) {
	if !finite(c.Floor) || !finite(c.Peak) || !finite(c.Rate) || c.Peak < c.Floor || c.Rate < 0 {
		return nil, fmt.Errorf("modulation: invalid decay envelope")
	}
	return &Decay{config: c, value: c.Floor}, nil
}
func (d *Decay) Step(trigger bool, delta float64) float64 {
	if trigger {
		d.value = d.config.Peak
	} else if finite(delta) && delta >= 0 {
		d.value = math.Max(d.config.Floor, d.value-d.config.Rate*delta)
	}
	return d.value
}
func (d *Decay) Value() float64 { return d.value }
func (d *Decay) Reset()         { d.value = d.config.Floor }

// Change emits one trigger when a sampled input changes. Its zero value uses
// the type's zero value as the initial input, preserving silent voice startup.
// Call once per Update, then share the resulting event among several envelopes.
type Change[T comparable] struct{ previous T }

func (c *Change[T]) Sample(value T) bool {
	changed := value != c.previous
	c.previous = value
	return changed
}
