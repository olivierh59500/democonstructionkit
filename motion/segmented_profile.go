package motion

import (
	"fmt"
	"math"
)

type ProfileRound uint8

const (
	ProfileHalfUp ProfileRound = iota
	ProfileFloor
	ProfileNone
)

// ProfileSegment compiles one wave cycle into an authored lookup table. A
// rectified segment uses Wave.At, including its absolute-value convention;
// ordinary segments retain amplitude*sin(phase)/2 + Offset arithmetic.
type ProfileSegment struct {
	Samples, Cycles   int
	Amplitude, Offset float64
	Rectify           bool
}

type SegmentedProfileConfig struct {
	Segments       []ProfileSegment
	AppendTail     bool
	TailValue      float64
	Start, Restart int
	WrapMargin     int
	VisibleSamples int
	Round          ProfileRound
}

// SegmentedProfile owns its read-only wave table and strict-wrap cursor. A
// trailing authored sentinel can extend the counter's limit without ever
// becoming a sampled visible strip.
type SegmentedProfile struct {
	config SegmentedProfileConfig
	table  []float64
	index  int
}

func NewSegmentedProfile(c SegmentedProfileConfig) (*SegmentedProfile, error) {
	if len(c.Segments) == 0 || len(c.Segments) > 1<<16 ||
		c.VisibleSamples < 1 || c.WrapMargin < c.VisibleSamples ||
		c.Start < 0 || c.Restart < 0 || c.Round > ProfileNone ||
		math.IsNaN(c.TailValue) || math.IsInf(c.TailValue, 0) {
		return nil, fmt.Errorf("motion: invalid segmented profile")
	}
	count := 0
	for _, segment := range c.Segments {
		if segment.Samples < 1 || segment.Samples > 1<<20-count || segment.Cycles < 0 ||
			math.IsNaN(segment.Amplitude) || math.IsInf(segment.Amplitude, 0) ||
			math.IsNaN(segment.Offset) || math.IsInf(segment.Offset, 0) {
			return nil, fmt.Errorf("motion: invalid profile segment")
		}
		count += segment.Samples
	}
	if c.AppendTail {
		count++
	}
	if c.WrapMargin >= count || c.Start > count-c.WrapMargin || c.Restart > count-c.WrapMargin {
		return nil, fmt.Errorf("motion: invalid profile wrap or visible range")
	}
	c.Segments = append([]ProfileSegment(nil), c.Segments...)
	p := &SegmentedProfile{config: c, table: make([]float64, 0, count), index: c.Start}
	for _, segment := range c.Segments {
		cycles := segment.Cycles
		if cycles == 0 {
			cycles = 1
		}
		phase := 0.0
		step := 2 * math.Pi / float64(segment.Samples)
		if cycles != 1 {
			step *= float64(cycles)
		}
		wave := Wave{Amplitude: segment.Amplitude, Speed: 1, Rectify: true}
		for range segment.Samples {
			value := segment.Amplitude*math.Sin(phase)/2 + segment.Offset
			if segment.Rectify {
				value = wave.At(0, phase)
			}
			switch c.Round {
			case ProfileHalfUp:
				value = math.Floor(value + .5)
			case ProfileFloor:
				value = math.Floor(value)
			}
			p.table = append(p.table, value)
			phase += step
		}
	}
	if c.AppendTail {
		p.table = append(p.table, c.TailValue)
	}
	return p, nil
}

func (p *SegmentedProfile) Step() {
	p.index++
	if p.index > len(p.table)-p.config.WrapMargin {
		p.index = p.config.Restart
	}
}

func (p *SegmentedProfile) At(offset int) float64 {
	if offset < 0 || offset >= p.config.VisibleSamples {
		return 0
	}
	return p.table[p.index+offset]
}

func (p *SegmentedProfile) Index() int       { return p.index }
func (p *SegmentedProfile) Table() []float64 { return p.table }
func (p *SegmentedProfile) Reset()           { p.index = p.config.Start }
