package motion

import (
	"fmt"
	"time"
)

// SampledSignalsConfig stores one compact channel mask per source tick. Masks
// use consecutive little-endian bytes, with channel zero in the lowest bit.
// LoopStart is a frame index: the introduction plays once before that region
// repeats. Rate is the source clock, independent of the application's updates.
type SampledSignalsConfig struct {
	Channels, Rate int
	Masks          []byte
	Loop           bool
	LoopStart      int
}

// SampledSignals owns its masks and samples an external clock without advancing
// it. An audio player's audible position can drive sprites, palettes or layers.
// Sampling supports paused clocks, skipped updates and backwards seeks.
type SampledSignals struct {
	config SampledSignalsConfig
	stride int
	frames int
}

func NewSampledSignals(c SampledSignalsConfig) (*SampledSignals, error) {
	if c.Channels < 1 || c.Channels > 64 || c.Rate < 1 || c.Rate > 8000 {
		return nil, fmt.Errorf("motion: invalid sampled signal channels or rate")
	}
	stride := (c.Channels + 7) / 8
	if len(c.Masks) == 0 || len(c.Masks)%stride != 0 || len(c.Masks)/stride > 1<<24 ||
		c.LoopStart < 0 || c.LoopStart >= len(c.Masks)/stride {
		return nil, fmt.Errorf("motion: invalid sampled signal bank or loop start")
	}
	if used := c.Channels % 8; used != 0 {
		for i := stride - 1; i < len(c.Masks); i += stride {
			if c.Masks[i]>>used != 0 {
				return nil, fmt.Errorf("motion: sampled signal mask exceeds channel count")
			}
		}
	}
	c.Masks = append([]byte(nil), c.Masks...)
	return &SampledSignals{config: c, stride: stride, frames: len(c.Masks) / stride}, nil
}

// FrameAt returns -1 after a finite track. Negative positions clamp to the
// opening frame; large positions wrap in bounded time when looping is enabled.
func (s *SampledSignals) FrameAt(position time.Duration) int {
	if s == nil {
		return -1
	}
	position = max(position, 0)
	rate := int64(s.config.Rate)
	frame := int64(position/time.Second)*rate + int64(position%time.Second)*rate/int64(time.Second)
	if frame >= int64(s.frames) {
		if !s.config.Loop {
			return -1
		}
		start := int64(s.config.LoopStart)
		frame = start + (frame-start)%int64(s.frames-s.config.LoopStart)
	}
	return int(frame)
}

func (s *SampledSignals) At(position time.Duration, channel int) bool {
	if s == nil || channel < 0 || channel >= s.config.Channels {
		return false
	}
	frame := s.FrameAt(position)
	if frame < 0 {
		return false
	}
	return s.config.Masks[frame*s.stride+channel/8]&(1<<uint(channel%8)) != 0
}

func (s *SampledSignals) Duration() time.Duration {
	return time.Duration(int64(s.frames) * int64(time.Second) / int64(s.config.Rate))
}
