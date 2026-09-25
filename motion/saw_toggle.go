package motion

import (
	"fmt"
	"math"
)

// SawToggleConfig advances a scalar to a strict boundary, jumps to Restart,
// and alternates a boolean face on each jump. Negative Velocity uses the lower
// boundary instead. The restart is absolute, so no overshoot is retained.
type SawToggleConfig struct {
	Start, Velocity, Boundary, Restart float64
	StartAlternate                     bool
}

// SawToggle is a reusable scalar/face clock for image flips and staged motion.
type SawToggle struct {
	config    SawToggleConfig
	value     float64
	alternate bool
}

func NewSawToggle(config SawToggleConfig) (*SawToggle, error) {
	for _, value := range [...]float64{config.Start, config.Velocity, config.Boundary, config.Restart} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite saw toggle setting")
		}
	}
	if config.Velocity == 0 || config.Velocity > 0 && (config.Restart >= config.Boundary || config.Start > config.Boundary) ||
		config.Velocity < 0 && (config.Restart <= config.Boundary || config.Start < config.Boundary) {
		return nil, fmt.Errorf("motion: invalid saw toggle direction or bounds")
	}
	cycle := &SawToggle{config: config}
	cycle.Reset()
	return cycle, nil
}

func (cycle *SawToggle) At() float64     { return cycle.value }
func (cycle *SawToggle) Alternate() bool { return cycle.alternate }
func (cycle *SawToggle) Reset() {
	cycle.value, cycle.alternate = cycle.config.Start, cycle.config.StartAlternate
}
func (cycle *SawToggle) Step() {
	cycle.value += cycle.config.Velocity
	if cycle.config.Velocity > 0 && cycle.value > cycle.config.Boundary ||
		cycle.config.Velocity < 0 && cycle.value < cycle.config.Boundary {
		cycle.value = cycle.config.Restart
		cycle.alternate = !cycle.alternate
	}
}
