// Package motion provides deterministic animation functions in seconds and radians.
package motion

import "math"

// Wrap returns a value in [0, period), including for negative coordinates.
func Wrap(value, period float64) float64 {
	if period <= 0 || math.IsNaN(value) || math.IsInf(value, 0) || math.IsInf(period, 0) {
		return 0
	}
	return value - math.Floor(value/period)*period
}

// Wave is a spatial sine or cosine traveling with angular speed in radians per
// caller-selected time unit. Cos selects cosine, Rectify takes its absolute
// value before signed Amplitude, and Offset translates the result.
type Wave struct {
	Amplitude, Spatial, Speed, Phase float64
	Offset                           float64
	Cos, Rectify                     bool
}

// At samples a wave at a position and an absolute time.
func (w Wave) At(position, seconds float64) float64 {
	phase := position*w.Spatial + seconds*w.Speed + w.Phase
	var wave float64
	if w.Cos {
		wave = math.Cos(phase)
	} else {
		wave = math.Sin(phase)
	}
	if w.Rectify {
		wave = math.Abs(wave)
	}
	value := w.Amplitude * wave
	if w.Offset != 0 {
		value += w.Offset
	}
	return value
}

// Waves combines independent harmonics without allocating when sampled.
type Waves []Wave

func (w Waves) At(position, seconds float64) float64 {
	sum := 0.0
	for _, v := range w {
		sum += v.At(position, seconds)
	}
	return sum
}

// Table samples a periodic original-demo lookup table with linear interpolation.
// Rate converts seconds into table entries; Phase is measured in entries.
type Table struct {
	Values      []float64
	Rate, Phase float64
}

func (t Table) At(seconds float64) float64 {
	if len(t.Values) == 0 {
		return 0
	}
	x := Wrap(seconds*t.Rate+t.Phase, float64(len(t.Values)))
	i := int(x)
	return Lerp(t.Values[i], t.Values[(i+1)%len(t.Values)], x-float64(i))
}

// Ease transforms normalized progress. Implementations clamp to [0, 1].
type Ease func(float64) float64

func Linear(t float64) float64 { return math.Max(0, math.Min(1, t)) }
func Smooth(t float64) float64 { t = Linear(t); return t * t * (3 - 2*t) }
func Sine(t float64) float64   { return (1 - math.Cos(math.Pi*Linear(t))) / 2 }
func ElasticOut(t float64) float64 {
	t = Linear(t)
	if t == 0 || t == 1 {
		return t
	}
	return math.Pow(2, -10*t)*math.Sin((10*t-.75)*2*math.Pi/3) + 1
}
func ElasticIn(t float64) float64  { return 1 - ElasticOut(1-Linear(t)) }
func Lerp(a, b, t float64) float64 { return a + (b-a)*t }

// Tween samples an absolute-time animation. Nonpositive duration is an instant cut.
func Tween(from, to, start, duration, seconds float64, ease Ease) float64 {
	if seconds < start {
		return from
	}
	if duration <= 0 {
		return to
	}
	if ease == nil {
		ease = Linear
	}
	return Lerp(from, to, ease(Linear((seconds-start)/duration)))
}
