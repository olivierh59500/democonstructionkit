package motion

import (
	"fmt"
	"math"
)

// PhaseSegment appends Samples values to a cyclic phase table. Reset supplies
// the same phase on every sample; otherwise Step accumulates before sampling.
// A nil Reset differs from an explicit reset to zero.
type PhaseSegment struct {
	Samples int
	Step    float64
	Reset   *float64
}

type PhaseSequenceConfig struct {
	Start, Period float64
	Segments      []PhaseSegment
}

// PhaseSequence compiles exact ordered phase updates once. Current and Step
// are allocation-free and independent of rendering and display refresh.
type PhaseSequence struct {
	values []float64
	index  int
}

func NewPhaseSequence(config PhaseSequenceConfig) (*PhaseSequence, error) {
	if !pathFinite(config.Start) || !pathFinite(config.Period) || config.Period <= 0 ||
		len(config.Segments) == 0 || len(config.Segments) > 1<<16 {
		return nil, fmt.Errorf("motion: invalid phase sequence configuration")
	}
	count := 0
	for _, segment := range config.Segments {
		if segment.Samples < 1 || segment.Samples > 1<<20-count || !pathFinite(segment.Step) ||
			segment.Reset != nil && !pathFinite(*segment.Reset) {
			return nil, fmt.Errorf("motion: invalid phase segment")
		}
		count += segment.Samples
	}
	sequence := &PhaseSequence{values: make([]float64, 0, count)}
	phase := config.Start
	for _, segment := range config.Segments {
		for range segment.Samples {
			if segment.Reset != nil {
				phase = *segment.Reset
			} else {
				phase += segment.Step
			}
			if !pathFinite(phase) {
				return nil, fmt.Errorf("motion: phase sequence overflow")
			}
			sequence.values = append(sequence.values, math.Mod(phase, config.Period))
		}
	}
	return sequence, nil
}

// Current returns the phase for the next draw, before Step advances the cursor.
func (sequence *PhaseSequence) Current() float64 { return sequence.values[sequence.index] }
func (sequence *PhaseSequence) Index() int       { return sequence.index }
func (sequence *PhaseSequence) Len() int         { return len(sequence.values) }

func (sequence *PhaseSequence) Step() {
	sequence.index++
	if sequence.index == len(sequence.values) {
		sequence.index = 0
	}
}

func (sequence *PhaseSequence) Reset() { sequence.index = 0 }

// AtTick samples an absolute scene tick without moving the sequential cursor.
func (sequence *PhaseSequence) AtTick(tick int) float64 {
	index := tick % len(sequence.values)
	if index < 0 {
		index += len(sequence.values)
	}
	return sequence.values[index]
}
