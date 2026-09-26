package modulation

import (
	"fmt"
	"math"
)

// PeriodicDecayConfig starts an envelope at one counter value in a looping
// cycle. The peak decays on its trigger tick, unlike a standard retriggerable
// envelope that holds its peak for one full frame. Trigger may add live music
// or input events without changing the periodic cadence.
type PeriodicDecayConfig struct {
	Period, TriggerAt, Start int
	Floor, Peak, Rate, Gain  float64
	Trigger                  func(tick int) bool
}

type PeriodicDecay struct {
	config        PeriodicDecayConfig
	counter, tick int
	value         float64
}

func NewPeriodicDecay(c PeriodicDecayConfig) (*PeriodicDecay, error) {
	if c.Period < 1 || c.TriggerAt < 0 || c.TriggerAt >= c.Period ||
		c.Start < 0 || c.Start >= c.Period ||
		math.IsNaN(c.Floor) || math.IsInf(c.Floor, 0) ||
		math.IsNaN(c.Peak) || math.IsInf(c.Peak, 0) ||
		math.IsNaN(c.Rate) || math.IsInf(c.Rate, 0) ||
		math.IsNaN(c.Gain) || math.IsInf(c.Gain, 0) ||
		c.Peak < c.Floor || c.Rate < 0 {
		return nil, fmt.Errorf("modulation: invalid periodic decay")
	}
	d := &PeriodicDecay{config: c}
	d.Reset()
	return d, nil
}

func (d *PeriodicDecay) Reset() {
	d.counter, d.tick, d.value = d.config.Start, 0, d.config.Floor
}

func (d *PeriodicDecay) Step() float64 {
	d.counter++
	if d.counter >= d.config.Period {
		d.counter = 0
	}
	trigger := d.counter == d.config.TriggerAt
	if d.config.Trigger != nil && d.config.Trigger(d.tick) {
		trigger = true
	}
	if trigger {
		d.value = d.config.Peak
	}
	if d.value > d.config.Floor {
		d.value -= d.config.Rate
	} else {
		d.value = d.config.Floor
	}
	d.tick++
	return d.Value()
}

func (d *PeriodicDecay) Value() float64 { return d.value * d.config.Gain }
func (d *PeriodicDecay) Counter() int   { return d.counter }
func (d *PeriodicDecay) Tick() int      { return d.tick }
