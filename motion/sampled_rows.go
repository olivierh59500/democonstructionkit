package motion

import (
	"fmt"
	"math"
)

// SampledRowPose describes one source crop and its destination transform.
// SourceY may leave the image bounds when an authored row should disappear.
type SampledRowPose struct {
	SourceY, SourceHeight float64
	X, Y, ScaleX, ScaleY  float64
}

// SampledRowProgram can choose any per-copy, per-row pose from an absolute
// scene clock. It stays independent of source images and drawing backends.
type SampledRowProgram interface {
	Sample(time float64, copyIndex, row int) SampledRowPose
}

// SampledRowFunc lets a production provide a custom trajectory without a new
// controller type. Named programs remain data-driven alternatives.
type SampledRowFunc func(time float64, copyIndex, row int) SampledRowPose

func (f SampledRowFunc) Sample(time float64, copyIndex, row int) SampledRowPose {
	return f(time, copyIndex, row)
}

// OuterSineRowsConfig makes several row trains share one breathing source-row
// height while each copy keeps its own X origin and phase offset.
type OuterSineRowsConfig struct {
	Bases, PhaseOffsets                                   []float64
	BaseY, HeightBase, HeightAmplitude, HeightTimeDivisor float64
	BounceAmplitude, BounceTimeDivisor                    float64
	HorizontalAmplitude, HorizontalDivisor                float64
	SourceHeight                                          float64
}

type OuterSineRows struct{ config OuterSineRowsConfig }

func NewOuterSineRows(c OuterSineRowsConfig) (*OuterSineRows, error) {
	if len(c.Bases) == 0 || len(c.Bases) != len(c.PhaseOffsets) || len(c.Bases) > 256 ||
		c.HeightTimeDivisor == 0 || c.BounceTimeDivisor == 0 || c.HorizontalDivisor == 0 || c.SourceHeight <= 0 {
		return nil, fmt.Errorf("motion: invalid outer sine row dimensions")
	}
	for _, value := range append(append([]float64(nil), c.Bases...), c.PhaseOffsets...) {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite row origin")
		}
	}
	for _, value := range [...]float64{c.BaseY, c.HeightBase, c.HeightAmplitude, c.HeightTimeDivisor,
		c.BounceAmplitude, c.BounceTimeDivisor, c.HorizontalAmplitude, c.HorizontalDivisor, c.SourceHeight} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite outer row parameter")
		}
	}
	c.Bases = append([]float64(nil), c.Bases...)
	c.PhaseOffsets = append([]float64(nil), c.PhaseOffsets...)
	return &OuterSineRows{config: c}, nil
}

func (p *OuterSineRows) Sample(time float64, copyIndex, row int) SampledRowPose {
	c := p.config
	height := c.HeightBase + (1+math.Sin(time/c.HeightTimeDivisor))*c.HeightAmplitude
	phase := c.PhaseOffsets[copyIndex]
	bounce := float64(int(c.BounceAmplitude * math.Cos(phase+time/c.BounceTimeDivisor)))
	offset := float64(int(c.HorizontalAmplitude * math.Sin(phase+(time+float64(row)*height)/c.HorizontalDivisor)))
	return SampledRowPose{
		SourceY: float64(int(float64(row) * height)), SourceHeight: c.SourceHeight,
		X: c.Bases[copyIndex] + offset, Y: c.BaseY + bounce + float64(row), ScaleX: 1, ScaleY: 1,
	}
}

// ZoomSineRowsConfig samples source rows in reverse while giving each row an
// independent zoom and vertical path. All divisors and amplitudes are editable.
type ZoomSineRowsConfig struct {
	BaseX, BaseY, ReverseSourceStart, SourceFactor, SourceHeight float64
	ZoomBase, ZoomAmplitude, ZoomFrequency                       float64
	ZoomTimeDivisor, ZoomIndexDivisor, XShift                    float64
	YAmplitude, YFrequency, YTimeDivisor, YIndexDivisor          float64
}

type ZoomSineRows struct{ config ZoomSineRowsConfig }

func NewZoomSineRows(c ZoomSineRowsConfig) (*ZoomSineRows, error) {
	if c.SourceHeight <= 0 || c.ZoomTimeDivisor == 0 || c.ZoomIndexDivisor == 0 ||
		c.YTimeDivisor == 0 || c.YIndexDivisor == 0 {
		return nil, fmt.Errorf("motion: invalid zoom row source or clock")
	}
	for _, value := range [...]float64{c.BaseX, c.BaseY, c.ReverseSourceStart, c.SourceFactor, c.SourceHeight,
		c.ZoomBase, c.ZoomAmplitude, c.ZoomFrequency, c.ZoomTimeDivisor, c.ZoomIndexDivisor, c.XShift,
		c.YAmplitude, c.YFrequency, c.YTimeDivisor, c.YIndexDivisor} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite zoom row parameter")
		}
	}
	return &ZoomSineRows{config: c}, nil
}

func (p *ZoomSineRows) Sample(time float64, _ int, row int) SampledRowPose {
	c := p.config
	i := float64(row)
	zoom := math.Sin(c.ZoomFrequency*(time/c.ZoomTimeDivisor+i/c.ZoomIndexDivisor))*c.ZoomAmplitude + c.ZoomBase
	y := c.YAmplitude * math.Sin(c.YFrequency*(time/c.YTimeDivisor+i*zoom/c.YIndexDivisor))
	return SampledRowPose{
		SourceY: float64(int((c.ReverseSourceStart - i) * c.SourceFactor)), SourceHeight: c.SourceHeight,
		X: c.BaseX - zoom*c.XShift, Y: c.BaseY + y, ScaleX: zoom, ScaleY: zoom,
	}
}
