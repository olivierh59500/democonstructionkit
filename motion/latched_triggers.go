package motion

import (
	"fmt"
	"math"
)

type TriggerSource uint8

const (
	TriggerRandom TriggerSource = iota
	TriggerSignal
)

// LatchedTriggersConfig samples either seeded random values or a caller's
// music/event signal. Signals can refresh each tick or only at the period
// boundary. Triggered slots remain latched until ReleaseAt after decrement.
type LatchedTriggersConfig struct {
	Count, Period, ReleaseAt int
	Source                   TriggerSource
	RandomFloat              func() float64
	RandomRange, RandomBias  float64
	Threshold                float64
	Signal                   func(index int, tick int) bool
	SignalEveryTick          bool
}

// LatchedTriggers prepares visibility before decrement/release. A trigger may
// reassert on the next tick even after the current release clears its latch.
type LatchedTriggers struct {
	config  LatchedTriggersConfig
	values  []float64
	latched []bool
	visible []bool
	timer   int
	tick    int
}

func NewLatchedTriggers(c LatchedTriggersConfig) (*LatchedTriggers, error) {
	if c.Count < 1 || c.Count > 1<<16 || c.Period < 1 || c.ReleaseAt < 0 || c.ReleaseAt >= c.Period ||
		c.Source > TriggerSignal || math.IsNaN(c.RandomRange) || math.IsInf(c.RandomRange, 0) ||
		math.IsNaN(c.RandomBias) || math.IsInf(c.RandomBias, 0) ||
		math.IsNaN(c.Threshold) || math.IsInf(c.Threshold, 0) ||
		c.Source == TriggerRandom && (c.RandomFloat == nil || c.RandomRange <= 0) ||
		c.Source == TriggerSignal && c.Signal == nil {
		return nil, fmt.Errorf("motion: invalid latched trigger bank")
	}
	b := &LatchedTriggers{config: c, values: make([]float64, c.Count),
		latched: make([]bool, c.Count), visible: make([]bool, c.Count)}
	return b, nil
}

func (b *LatchedTriggers) Reset() {
	clear(b.values)
	clear(b.latched)
	clear(b.visible)
	b.timer, b.tick = 0, 0
}

func (b *LatchedTriggers) sample() {
	for i := range b.values {
		if b.config.Source == TriggerRandom {
			b.values[i] = math.Floor(b.config.RandomFloat()*b.config.RandomRange) + b.config.RandomBias
		} else if b.config.Signal(i, b.tick) {
			b.values[i] = 1
		} else {
			b.values[i] = 0
		}
	}
}

// Step returns a borrowed visibility slice prepared for the current frame.
func (b *LatchedTriggers) Step() []bool {
	if b.timer == 0 {
		b.sample()
		b.timer = b.config.Period
	} else if b.config.Source == TriggerSignal && b.config.SignalEveryTick {
		b.sample()
	}
	for i, value := range b.values {
		if value >= b.config.Threshold {
			b.latched[i] = true
		}
		b.visible[i] = b.latched[i]
	}
	b.timer--
	if b.timer == b.config.ReleaseAt {
		clear(b.latched)
	}
	b.tick++
	return b.visible
}

func (b *LatchedTriggers) Visible(index int) bool { return b.visible[index] }
func (b *LatchedTriggers) Timer() int             { return b.timer }
func (b *LatchedTriggers) Tick() int              { return b.tick }
