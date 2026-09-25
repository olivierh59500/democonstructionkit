package motion

import (
	"fmt"
	"math"
)

// HoldBounceConfig starts at Max and moves inward at negative Velocity after
// its initial HoldTicks. LeadTicks may start the motion before the wait ends.
// The oscillator clamps at Min, returns to Max, then repeats the full hold.
type HoldBounceConfig struct {
	Min, Max, Velocity   float64
	HoldTicks, LeadTicks int
}

// HoldBounce owns a one-dimensional bounce with a pause at its upper limit.
type HoldBounce struct {
	config          HoldBounceConfig
	value, velocity float64
	wait            int
}

func NewHoldBounce(config HoldBounceConfig) (*HoldBounce, error) {
	for _, value := range [...]float64{config.Min, config.Max, config.Velocity} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite hold bounce")
		}
	}
	if config.Min >= config.Max || config.Velocity >= 0 || config.HoldTicks < 0 || config.LeadTicks < 0 || config.LeadTicks > config.HoldTicks {
		return nil, fmt.Errorf("motion: invalid hold bounce bounds or timing")
	}
	bounce := &HoldBounce{config: config}
	bounce.Reset()
	return bounce, nil
}

func (bounce *HoldBounce) At() float64 { return bounce.value }

// Step preserves pre-step threshold checks and one tick of wait decrement.
func (bounce *HoldBounce) Step() float64 {
	if bounce.wait <= bounce.config.LeadTicks {
		bounce.value += bounce.velocity
		if bounce.value <= bounce.config.Min {
			bounce.value = bounce.config.Min
			bounce.velocity = -bounce.config.Velocity
		}
	}
	if bounce.wait <= 0 && bounce.value >= bounce.config.Max {
		bounce.value = bounce.config.Max
		bounce.velocity = bounce.config.Velocity
		bounce.wait = bounce.config.HoldTicks
	}
	bounce.wait--
	return bounce.value
}

func (bounce *HoldBounce) Reset() {
	bounce.value = bounce.config.Max
	bounce.velocity = bounce.config.Velocity
	bounce.wait = bounce.config.HoldTicks
}
