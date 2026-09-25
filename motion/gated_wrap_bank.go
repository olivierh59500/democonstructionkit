package motion

import (
	"fmt"
	"math"
)

// GatedWrapBankConfig gives each wrapped lane an activation time. One velocity
// may be broadcast to all lanes. By default a lane starts only when time is
// strictly greater than its gate; Inclusive also starts at equality.
type GatedWrapBankConfig struct {
	Start, Velocity, Gates []float64
	Inclusive              bool
	Lower, Upper           *WrapLimit
}

// GatedWrapBank owns independent clocks for atlas frames, rasters or sprites.
// Reading At and Frame does not change any lane or allocate an image.
type GatedWrapBank struct {
	bank      *WrapBank
	velocity  []float64
	gates     []float64
	inclusive bool
}

func NewGatedWrapBank(config GatedWrapBankConfig) (*GatedWrapBank, error) {
	if len(config.Start) == 0 || len(config.Gates) != len(config.Start) || len(config.Velocity) != 1 && len(config.Velocity) != len(config.Start) {
		return nil, fmt.Errorf("motion: invalid gated wrap dimensions")
	}
	for _, value := range append(append([]float64(nil), config.Gates...), config.Velocity...) {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite gated wrap setting")
		}
	}
	bank, err := NewWrapBank(WrapBankConfig{
		Start: config.Start, Velocity: []float64{0}, Lower: config.Lower, Upper: config.Upper,
	})
	if err != nil {
		return nil, err
	}
	velocity := make([]float64, len(config.Start))
	for i := range velocity {
		index := i
		if len(config.Velocity) == 1 {
			index = 0
		}
		velocity[i] = config.Velocity[index]
	}
	return &GatedWrapBank{bank: bank, velocity: velocity, gates: append([]float64(nil), config.Gates...), inclusive: config.Inclusive}, nil
}

func (gated *GatedWrapBank) Len() int             { return gated.bank.Len() }
func (gated *GatedWrapBank) At(index int) float64 { return gated.bank.At(index) }
func (gated *GatedWrapBank) Frame(index int) int  { return int(math.Floor(gated.At(index))) }

// StepAt advances active lanes once at the caller's logical scene time.
func (gated *GatedWrapBank) StepAt(time float64) error {
	if math.IsNaN(time) || math.IsInf(time, 0) {
		return fmt.Errorf("motion: nonfinite gated wrap time")
	}
	for index, gate := range gated.gates {
		gated.bank.velocity[index] = 0
		if time > gate || gated.inclusive && time == gate {
			gated.bank.velocity[index] = gated.velocity[index]
		}
	}
	gated.bank.Step()
	return nil
}

func (gated *GatedWrapBank) Reset() { gated.bank.Reset() }
