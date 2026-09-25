package timeline

import (
	"fmt"
	"math"
)

// CueRange is a finite or open-ended time interval. The default is [Start,End);
// OpenStart and IncludeEnd independently control its two exact boundaries.
type CueRange struct {
	Start, End            float64
	OpenStart, IncludeEnd bool
}

// CueRanges owns a validated ordered range table. Sampling allocates nothing
// and returns the authored range index for scene data such as text or material.
type CueRanges struct{ ranges []CueRange }

func NewCueRanges(ranges []CueRange) (*CueRanges, error) {
	if len(ranges) == 0 || len(ranges) > 1<<16 {
		return nil, fmt.Errorf("timeline: invalid cue range count")
	}
	for i, current := range ranges {
		if math.IsNaN(current.Start) || math.IsInf(current.Start, 0) || math.IsNaN(current.End) || current.Start < 0 || current.End <= current.Start {
			return nil, fmt.Errorf("timeline: invalid cue range")
		}
		if i > 0 {
			previous := ranges[i-1]
			if previous.End > current.Start || previous.End == current.Start && previous.IncludeEnd && !current.OpenStart {
				return nil, fmt.Errorf("timeline: overlapping cue ranges")
			}
		}
	}
	return &CueRanges{ranges: append([]CueRange(nil), ranges...)}, nil
}

func (program *CueRanges) Len() int { return len(program.ranges) }

func (program *CueRanges) At(seconds float64) (index int, window CueRange, active bool) {
	if math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return -1, CueRange{}, false
	}
	for i, candidate := range program.ranges {
		if seconds < candidate.Start || seconds == candidate.Start && candidate.OpenStart {
			continue
		}
		if seconds > candidate.End || seconds == candidate.End && !candidate.IncludeEnd {
			continue
		}
		return i, candidate, true
	}
	return -1, CueRange{}, false
}

// SteppedEnvelopeConfig indexes an image/color bank from MaxIndex to zero on
// entry and back toward MaxIndex before exit. Step is seconds per bank level.
type SteppedEnvelopeConfig struct {
	MaxIndex                    int
	EntryLength, ExitLead, Step float64
}

type SteppedEnvelope struct{ config SteppedEnvelopeConfig }

func NewSteppedEnvelope(config SteppedEnvelopeConfig) (*SteppedEnvelope, error) {
	if config.MaxIndex < 0 || config.MaxIndex > 1<<16 || config.EntryLength < 0 || config.ExitLead < 0 || config.Step <= 0 ||
		math.IsNaN(config.EntryLength) || math.IsNaN(config.ExitLead) || math.IsNaN(config.Step) ||
		math.IsInf(config.EntryLength, 0) || math.IsInf(config.ExitLead, 0) || math.IsInf(config.Step, 0) {
		return nil, fmt.Errorf("timeline: invalid stepped envelope")
	}
	return &SteppedEnvelope{config: config}, nil
}

// At returns a borrowed palette/image index. End may be +Inf for a final hold.
func (envelope *SteppedEnvelope) At(seconds, start, end float64) int {
	c := envelope.config
	if seconds < start {
		return c.MaxIndex
	}
	if seconds < start+c.EntryLength {
		return max(0, min(c.MaxIndex, c.MaxIndex-int((seconds-start)/c.Step)))
	}
	if seconds >= end-c.ExitLead {
		return max(0, min(c.MaxIndex, int((seconds-end+c.ExitLead)/c.Step)))
	}
	return 0
}
