package motion

import (
	"fmt"
	"math"
)

// RecurrentRowWaveConfig places successive rendered strips on a flat baseline
// until a staggered time threshold, then samples a cosine wave. The wave angle
// advances by a sine/cosine recurrence after each rendered strip. This makes
// skipped or clipped strips retain the source painter's phase behavior.
type RecurrentRowWaveConfig struct {
	Base, Flat, Amplitude              float64
	RevealTime, IndexDelay             float64
	SampleTimeStep                     float64
	StartAngle, TimeDivisor, AngleStep float64
	FrameStart, FrameStep              float64 // Scene time before the first draw and per simulation update.
}

// RecurrentRowWave is reset with Begin before each Draw. At must be called in
// painter order, once for each strip that will actually be rendered.
type RecurrentRowWave struct {
	config           RecurrentRowWaveConfig
	sinStep, cosStep float64
	time, sin, cos   float64
	frameTime        float64
}

func NewRecurrentRowWave(c RecurrentRowWaveConfig) (*RecurrentRowWave, error) {
	for _, v := range [...]float64{c.Base, c.Flat, c.Amplitude, c.RevealTime, c.IndexDelay,
		c.SampleTimeStep, c.StartAngle, c.TimeDivisor, c.AngleStep, c.FrameStart, c.FrameStep} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("motion: nonfinite recurrent row wave parameter")
		}
	}
	if c.TimeDivisor == 0 {
		return nil, fmt.Errorf("motion: recurrent row wave needs a time divisor")
	}
	sinStep, cosStep := math.Sincos(c.AngleStep)
	return &RecurrentRowWave{config: c, sinStep: sinStep, cosStep: cosStep, frameTime: c.FrameStart}, nil
}

// AdvanceFrame moves the scene clock once without changing the strip recurrence.
// Call it only on frames where this effect is active.
func (w *RecurrentRowWave) AdvanceFrame() error {
	next := w.frameTime + w.config.FrameStep
	if math.IsNaN(next) || math.IsInf(next, 0) {
		return fmt.Errorf("motion: invalid recurrent row wave frame time")
	}
	w.frameTime = next
	return nil
}

// FrameTime reports the scene time that BeginFrame will sample.
func (w *RecurrentRowWave) FrameTime() float64 { return w.frameTime }

// SetFrameStep changes the pace without resetting the current scene time.
func (w *RecurrentRowWave) SetFrameStep(step float64) error {
	if math.IsNaN(step) || math.IsInf(step, 0) {
		return fmt.Errorf("motion: invalid recurrent row wave frame step")
	}
	w.config.FrameStep = step
	return nil
}

// SetFrameTime seeks to a cue without changing the strip program or tempo.
func (w *RecurrentRowWave) SetFrameTime(time float64) error {
	if math.IsNaN(time) || math.IsInf(time, 0) {
		return fmt.Errorf("motion: invalid recurrent row wave frame time")
	}
	w.frameTime = time
	return nil
}

// ResetFrame returns to the configured start time, retaining the current step.
func (w *RecurrentRowWave) ResetFrame() { w.frameTime = w.config.FrameStart }

// BeginFrame restarts the strip recurrence at the owned scene time. The
// original Begin(time) remains available for externally scheduled scenes.
func (w *RecurrentRowWave) BeginFrame() error { return w.Begin(w.frameTime) }

// Begin restarts the draw-order recurrence at an authored scene time.
func (w *RecurrentRowWave) Begin(time float64) error {
	if math.IsNaN(time) || math.IsInf(time, 0) {
		return fmt.Errorf("motion: invalid recurrent row wave time")
	}
	w.time = time
	w.sin, w.cos = math.Sincos(w.config.StartAngle + time/w.config.TimeDivisor)
	return nil
}

func (w *RecurrentRowWave) At(index int) float64 {
	y := w.config.Flat
	if w.time > w.config.RevealTime-float64(index)*w.config.IndexDelay {
		y = w.config.Amplitude * w.cos
	}
	w.time += w.config.SampleTimeStep
	w.sin, w.cos = w.sin*w.cosStep+w.cos*w.sinStep, w.cos*w.cosStep-w.sin*w.sinStep
	return w.config.Base + y
}
