package motion

import (
	"fmt"
	"math"
)

// WarpTableClockConfig couples a cyclic row-source lookup with an independent
// cosine displacement for image columns. All coordinates are source pixels;
// the caller chooses row/column thickness and final material composition.
type WarpTableClockConfig struct {
	Horizontal        []float64
	SourceOrigin      int
	StartTick         uint64
	VerticalStart     float64
	VerticalStep      float64
	VerticalSpatial   float64
	VerticalAmplitude float64
}

// WarpTableClock owns two synchronized phases. The same instance may drive a
// text surface and any number of logos, each with independent strip variation.
type WarpTableClock struct {
	config WarpTableClockConfig
	table  []float64
	tick   uint64
	phase  float64
}

func NewWarpTableClock(config WarpTableClockConfig) (*WarpTableClock, error) {
	if len(config.Horizontal) == 0 || len(config.Horizontal) > 1<<20 {
		return nil, fmt.Errorf("motion: invalid warp lookup table")
	}
	for _, value := range [...]float64{config.VerticalStart, config.VerticalStep, config.VerticalSpatial, config.VerticalAmplitude} {
		if !pathFinite(value) {
			return nil, fmt.Errorf("motion: nonfinite warp clock parameter")
		}
	}
	for _, value := range config.Horizontal {
		if !pathFinite(value) || math.Abs(value+float64(config.SourceOrigin)) >= 1<<30 {
			return nil, fmt.Errorf("motion: invalid warp source offset")
		}
	}
	clock := &WarpTableClock{config: config, table: append([]float64(nil), config.Horizontal...),
		tick: config.StartTick, phase: config.VerticalStart}
	clock.config.Horizontal = nil
	return clock, nil
}

// SampleX preserves modulo and float-to-int truncation after the table value
// and source origin are summed. Signed rows support independent logo offsets.
func (clock *WarpTableClock) SampleX(row int) int {
	period := len(clock.table)
	index := (int(clock.tick%uint64(period)) + row) % period
	if index < 0 {
		index += period
	}
	return int(clock.table[index] + float64(clock.config.SourceOrigin))
}

// OffsetY samples the current cosine phase before the next Step.
func (clock *WarpTableClock) OffsetY(column int) float64 {
	return math.Cos(clock.phase+float64(column)*clock.config.VerticalSpatial) * clock.config.VerticalAmplitude
}

func (clock *WarpTableClock) Step(speed float64) error {
	if clock == nil || !pathFinite(speed) || clock.tick == ^uint64(0) {
		return fmt.Errorf("motion: invalid warp clock step")
	}
	next := clock.phase + clock.config.VerticalStep*speed
	if !pathFinite(next) {
		return fmt.Errorf("motion: warp phase overflow")
	}
	clock.tick++
	clock.phase = next
	return nil
}

func (clock *WarpTableClock) Tick() uint64   { return clock.tick }
func (clock *WarpTableClock) Phase() float64 { return clock.phase }
func (clock *WarpTableClock) Len() int       { return len(clock.table) }

// Table exposes borrowed read-only samples for authoring or fidelity checks.
func (clock *WarpTableClock) Table() []float64 { return clock.table }

func (clock *WarpTableClock) SetTick(tick uint64) { clock.tick = tick }
func (clock *WarpTableClock) SetPhase(phase float64) error {
	if !pathFinite(phase) {
		return fmt.Errorf("motion: nonfinite warp phase")
	}
	clock.phase = phase
	return nil
}
func (clock *WarpTableClock) Reset() {
	clock.tick, clock.phase = clock.config.StartTick, clock.config.VerticalStart
}
