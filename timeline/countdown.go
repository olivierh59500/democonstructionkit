package timeline

import (
	"fmt"
	"math"
)

// CountdownConfig describes a two-stage display countdown. The second counter
// starts decrementing on the tick when the first reaches zero. Hold keeps the
// final display visible, while FadeLead starts an independent cue before blank.
type CountdownConfig struct {
	First, Second, Hold, FadeLead int
}

type CountdownState struct {
	First, Second int
	Blank         bool
}

// Countdown owns validated thresholds but no display surface or tick clock.
// At is a stateless sample, so multiple render passes see the same values.
type Countdown struct {
	config    CountdownConfig
	blankTick int
	fadeStart int
}

func NewCountdown(config CountdownConfig) (*Countdown, error) {
	if config.First < 0 || config.Second < 0 || config.Hold < 0 || config.FadeLead < 0 || config.First > math.MaxInt-config.Second || config.First+config.Second > math.MaxInt-config.Hold {
		return nil, fmt.Errorf("timeline: invalid countdown dimensions")
	}
	end := config.First + config.Second
	return &Countdown{config: config, blankTick: end + config.Hold, fadeStart: end - config.FadeLead}, nil
}

func (countdown *Countdown) BlankTick() int     { return countdown.blankTick }
func (countdown *Countdown) FadeStartTick() int { return countdown.fadeStart }

func (countdown *Countdown) At(tick int) CountdownState {
	return CountdownState{
		First:  max(0, countdown.config.First-tick),
		Second: countdown.config.Second - max(0, tick-countdown.config.First+1),
		Blank:  tick >= countdown.blankTick,
	}
}
