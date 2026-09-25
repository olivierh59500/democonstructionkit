package motion

import (
	"fmt"
	"math"
)

// WrapLimit replaces a position after it crosses Boundary. Inclusive also
// wraps when the position lands exactly on the boundary.
type WrapLimit struct {
	Boundary, Restart float64
	Inclusive         bool
}

// WrapBankConfig gives independent layers their initial offsets and speeds.
// One velocity is broadcast to every layer; otherwise each has its own.
// Lower and Upper are optional, but at least one must be supplied. When both
// trigger in one update, the upper rule runs first, followed by the lower rule.
type WrapBankConfig struct {
	Start, Velocity []float64
	Lower, Upper    *WrapLimit
}

// WrapBank is a reusable stateful transport for backgrounds, rasters, sprites
// and logos. Drawing or reading At never advances the transport.
type WrapBank struct {
	config    WrapBankConfig
	positions []float64
	velocity  []float64
}

func NewWrapBank(config WrapBankConfig) (*WrapBank, error) {
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	if len(config.Start) == 0 || len(config.Start) > 1_000_000 || len(config.Velocity) != 1 && len(config.Velocity) != len(config.Start) ||
		config.Lower == nil && config.Upper == nil {
		return nil, fmt.Errorf("motion: invalid wrap bank size or missing limit")
	}
	if config.Lower != nil {
		lower := *config.Lower
		if !finite(lower.Boundary) || !finite(lower.Restart) {
			return nil, fmt.Errorf("motion: nonfinite lower wrap limit")
		}
		config.Lower = &lower
	}
	if config.Upper != nil {
		upper := *config.Upper
		if !finite(upper.Boundary) || !finite(upper.Restart) {
			return nil, fmt.Errorf("motion: nonfinite upper wrap limit")
		}
		config.Upper = &upper
	}
	if config.Lower != nil && config.Upper != nil && config.Lower.Boundary >= config.Upper.Boundary {
		return nil, fmt.Errorf("motion: overlapping wrap limits")
	}
	for _, value := range append(append([]float64(nil), config.Start...), config.Velocity...) {
		if !finite(value) {
			return nil, fmt.Errorf("motion: nonfinite wrap position or velocity")
		}
	}
	config.Start = append([]float64(nil), config.Start...)
	config.Velocity = append([]float64(nil), config.Velocity...)
	bank := &WrapBank{config: config, positions: make([]float64, len(config.Start)), velocity: make([]float64, len(config.Start))}
	bank.Reset()
	return bank, nil
}

func (bank *WrapBank) Len() int { return len(bank.positions) }

// At returns a layer's current offset without changing it.
func (bank *WrapBank) At(index int) float64 { return bank.positions[index] }

// Velocity returns a layer's current speed in units per Step.
func (bank *WrapBank) Velocity(index int) float64 { return bank.velocity[index] }

// Step advances all layers once without allocating.
func (bank *WrapBank) Step() {
	for index := range bank.positions {
		position := bank.positions[index] + bank.velocity[index]
		if limit := bank.config.Upper; limit != nil && (position > limit.Boundary || limit.Inclusive && position >= limit.Boundary) {
			position = limit.Restart
		}
		if limit := bank.config.Lower; limit != nil && (position < limit.Boundary || limit.Inclusive && position <= limit.Boundary) {
			position = limit.Restart
		}
		bank.positions[index] = position
	}
}

// SetVelocity changes one speed without resetting its current position.
func (bank *WrapBank) SetVelocity(index int, velocity float64) error {
	if index < 0 || index >= len(bank.velocity) || math.IsNaN(velocity) || math.IsInf(velocity, 0) {
		return fmt.Errorf("motion: invalid wrap velocity")
	}
	bank.velocity[index] = velocity
	return nil
}

// AddVelocity adjusts one speed while retaining the current phase.
func (bank *WrapBank) AddVelocity(index int, delta float64) error {
	if index < 0 || index >= len(bank.velocity) || math.IsNaN(delta) || math.IsInf(delta, 0) || math.IsNaN(bank.velocity[index]+delta) || math.IsInf(bank.velocity[index]+delta, 0) {
		return fmt.Errorf("motion: invalid wrap velocity change")
	}
	bank.velocity[index] += delta
	return nil
}

// ReverseAll changes direction while preserving every position.
func (bank *WrapBank) ReverseAll() {
	for index := range bank.velocity {
		bank.velocity[index] = -bank.velocity[index]
	}
}

// Reset restores the configured positions and speeds.
func (bank *WrapBank) Reset() {
	copy(bank.positions, bank.config.Start)
	for index := range bank.velocity {
		velocityIndex := index
		if len(bank.config.Velocity) == 1 {
			velocityIndex = 0
		}
		bank.velocity[index] = bank.config.Velocity[velocityIndex]
	}
}
