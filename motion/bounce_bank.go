package motion

import (
	"fmt"
	"math"
)

// BounceBankConfig describes independent one-dimensional oscillators. One
// velocity is broadcast to every start position; otherwise each position must
// have its own velocity. By default, direction reverses only after crossing a
// bound, preserving the overshoot used by many raster and sprite effects.
type BounceBankConfig struct {
	Start, Velocity []float64
	Min, Max        float64
	Inclusive       bool // Reverse when touching a bound as well as crossing it.
	Clamp           bool // Clamp an overshoot to the nearest bound.
}

// BounceBank keeps the positions and directions of a reusable raster, sprite
// or logo train. Step changes state; At and Draw callbacks only read it.
type BounceBank struct {
	config    BounceBankConfig
	positions []float64
	velocity  []float64
}

func NewBounceBank(config BounceBankConfig) (*BounceBank, error) {
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	if len(config.Start) == 0 || len(config.Start) > 1_000_000 || len(config.Velocity) != 1 && len(config.Velocity) != len(config.Start) ||
		!finite(config.Min) || !finite(config.Max) || config.Min >= config.Max {
		return nil, fmt.Errorf("motion: invalid bounce bank size or bounds")
	}
	for _, start := range config.Start {
		if !finite(start) || start < config.Min || start > config.Max {
			return nil, fmt.Errorf("motion: invalid bounce start")
		}
	}
	for _, velocity := range config.Velocity {
		if !finite(velocity) {
			return nil, fmt.Errorf("motion: nonfinite bounce velocity")
		}
	}
	config.Start = append([]float64(nil), config.Start...)
	config.Velocity = append([]float64(nil), config.Velocity...)
	bank := &BounceBank{config: config, positions: make([]float64, len(config.Start)), velocity: make([]float64, len(config.Start))}
	bank.Reset()
	return bank, nil
}

func (bank *BounceBank) Len() int { return len(bank.positions) }

// At returns the current position of one item; it does not advance motion.
func (bank *BounceBank) At(index int) float64 { return bank.positions[index] }

// Step advances every item once without allocating.
func (bank *BounceBank) Step() {
	for index := range bank.positions {
		position := bank.positions[index] + bank.velocity[index]
		crossed := position < bank.config.Min || position > bank.config.Max
		if bank.config.Inclusive {
			crossed = position <= bank.config.Min || position >= bank.config.Max
		}
		if crossed {
			bank.velocity[index] = -bank.velocity[index]
			if bank.config.Clamp {
				position = math.Max(bank.config.Min, math.Min(bank.config.Max, position))
			}
		}
		bank.positions[index] = position
	}
}

// Reset restores the configured positions and directions.
func (bank *BounceBank) Reset() {
	copy(bank.positions, bank.config.Start)
	for index := range bank.velocity {
		velocityIndex := index
		if len(bank.config.Velocity) == 1 {
			velocityIndex = 0
		}
		bank.velocity[index] = bank.config.Velocity[velocityIndex]
	}
}
