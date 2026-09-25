package sprites

import (
	"fmt"
	"math"
)

// FrameSequence chooses an authored image order over a duration. A looping
// sequence wraps time; a finite sequence holds its last frame after Duration.
// Values are caller time units, allowing seconds or a music clock.
type FrameSequence struct {
	Duration float64
	Indices  []int
	Loop     bool
}

// NewFrameSequence validates and owns a copy of the authored frame order.
func NewFrameSequence(config FrameSequence) (FrameSequence, error) {
	if len(config.Indices) == 0 || len(config.Indices) > 1<<20 || math.IsNaN(config.Duration) || math.IsInf(config.Duration, 0) || config.Duration < 0 {
		return FrameSequence{}, fmt.Errorf("sprites: invalid frame sequence")
	}
	for _, index := range config.Indices {
		if index < 0 {
			return FrameSequence{}, fmt.Errorf("sprites: negative frame index")
		}
	}
	config.Indices = append([]int(nil), config.Indices...)
	return config, nil
}

// Current samples without mutating state or allocating. Empty sequences return
// zero for compatibility with optional animation slots.
func (sequence FrameSequence) Current(t float64) int {
	if len(sequence.Indices) == 0 {
		return 0
	}
	if sequence.Duration <= 0 {
		return sequence.Indices[0]
	}
	ct := math.Max(0, t)
	if !sequence.Loop && ct >= sequence.Duration {
		return sequence.Indices[len(sequence.Indices)-1]
	}
	if sequence.Loop {
		ct = math.Mod(ct, sequence.Duration)
	}
	cp := math.Min(ct/sequence.Duration, 1)
	frame := int(math.Floor(float64(len(sequence.Indices)) * cp))
	if frame >= len(sequence.Indices) {
		frame = len(sequence.Indices) - 1
	}
	return sequence.Indices[frame]
}
