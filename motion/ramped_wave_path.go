package motion

import (
	"fmt"
	"math"
)

// RampedWavePathConfig combines independent horizontal and vertical waves with
// one entrance envelope. Base and each wave's Offset remain fixed while the
// wave amplitudes rise from zero. Sample selects spatial coordinates for paths
// that should follow a moving source rather than a fixed point.
type RampedWavePathConfig struct {
	Base, Sample Point
	X, Y         Wave
	Start, Rise  float64
}

// RampedWavePath is a reusable absolute-time XY trajectory. It can be sampled
// directly by scrolling, sprites or input, or supplied to TrajectoryClock.
type RampedWavePath struct{ config RampedWavePathConfig }

func NewRampedWavePath(config RampedWavePathConfig) (*RampedWavePath, error) {
	values := [...]float64{
		config.Base.X, config.Base.Y, config.Sample.X, config.Sample.Y,
		config.X.Amplitude, config.X.Spatial, config.X.Speed, config.X.Phase, config.X.Offset,
		config.Y.Amplitude, config.Y.Spatial, config.Y.Speed, config.Y.Phase, config.Y.Offset,
		config.Start, config.Rise,
	}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite ramped wave path parameter")
		}
	}
	if config.Rise < 0 {
		return nil, fmt.Errorf("motion: negative ramped wave rise duration")
	}
	return &RampedWavePath{config: config}, nil
}

// At returns a pose without changing state. Rise zero means full amplitude
// immediately; otherwise the ramp reaches full amplitude at Start+Rise.
func (path *RampedWavePath) At(seconds float64) Point {
	if path == nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return Point{}
	}
	c := path.config
	gain := 1.0
	if c.Rise > 0 {
		gain = Linear((seconds - c.Start) / c.Rise)
	}
	xWave, yWave := c.X, c.Y
	xWave.Amplitude *= gain
	yWave.Amplitude *= gain
	return Point{
		X: c.Base.X + xWave.At(c.Sample.X, seconds),
		Y: c.Base.Y + yWave.At(c.Sample.Y, seconds),
	}
}
