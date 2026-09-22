// Package modulation describes reusable parameter animation as serializable data.
// A host supplies audible music time and named inputs once per Update; effects
// share the resulting signals without reading or decoding audio inside Draw.
package modulation

import (
	"fmt"
	"math"
	"sort"
)

type TimeBase string

const (
	Seconds TimeBase = "seconds"
	Beats   TimeBase = "beats"
)

type Waveform string

const (
	Sine     Waveform = "sine"
	Cosine   Waveform = "cosine"
	Triangle Waveform = "triangle"
	Saw      Waveform = "saw"
	Square   Waveform = "square"
)

type Interpolation string

const (
	Linear Interpolation = "linear"
	Smooth Interpolation = "smooth"
	Hold   Interpolation = "hold"
)

// Context is immutable while sampling a frame. Inputs may contain YM voice
// levels, module row events, a measured audio envelope or user controls.
type Context struct {
	Seconds, Beats float64
	Inputs         map[string]float64
}

// MusicContext computes a tempo grid from audible playback position. Supplying
// Position() from the audio output avoids using the decoder's buffered cursor.
// A manual clock can supply the same seconds for deterministic offline rendering.
func MusicContext(seconds, bpm, beatOffset float64, inputs map[string]float64) Context {
	return Context{Seconds: seconds, Beats: seconds*bpm/60 + beatOffset, Inputs: inputs}
}

type Oscillator struct {
	Shape     Waveform `json:"shape"`
	Amplitude float64  `json:"amplitude"`
	Frequency float64  `json:"frequency"` // Cycles per selected time unit.
	Phase     float64  `json:"phase"`     // Cycles, allowing direct spacing between instances.
}
type Key struct {
	Time          float64       `json:"time"`
	Value         float64       `json:"value"`
	Interpolation Interpolation `json:"interpolation,omitempty"`
}
type Input struct {
	Name string  `json:"name"`
	Gain float64 `json:"gain"`
}
type Range struct{ Min, Max float64 }

// Spec is editor-friendly data, without function pointers. The output is Base
// plus the sampled key track, oscillators and weighted inputs. Clamp is optional.
// Loop repeats only the key track; oscillator phases remain continuous.
type Spec struct {
	Base        float64      `json:"base"`
	TimeBase    TimeBase     `json:"timeBase,omitempty"`
	Oscillators []Oscillator `json:"oscillators,omitempty"`
	Keys        []Key        `json:"keys,omitempty"`
	Loop        float64      `json:"loop,omitempty"`
	Inputs      []Input      `json:"inputs,omitempty"`
	Clamp       *Range       `json:"clamp,omitempty"`
}

// Signal owns a validated immutable configuration. Sampling performs no allocation.
type Signal struct{ spec Spec }

func New(c Spec) (*Signal, error) {
	if !finite(c.Base) || !finite(c.Loop) || c.Loop < 0 || (c.TimeBase != "" && c.TimeBase != Seconds && c.TimeBase != Beats) {
		return nil, fmt.Errorf("modulation: invalid base, loop or time unit")
	}
	for _, o := range c.Oscillators {
		if !finite(o.Amplitude) || !finite(o.Frequency) || !finite(o.Phase) {
			return nil, fmt.Errorf("modulation: nonfinite oscillator")
		}
		switch o.Shape {
		case Sine, Cosine, Triangle, Saw, Square:
		default:
			return nil, fmt.Errorf("modulation: unknown waveform %q", o.Shape)
		}
	}
	for i, k := range c.Keys {
		if !finite(k.Time) || !finite(k.Value) || k.Time < 0 || (i > 0 && k.Time <= c.Keys[i-1].Time) {
			return nil, fmt.Errorf("modulation: key times must increase")
		}
		switch k.Interpolation {
		case "", Linear, Smooth, Hold:
		default:
			return nil, fmt.Errorf("modulation: unknown interpolation")
		}
	}
	if c.Loop > 0 && len(c.Keys) > 0 && c.Keys[len(c.Keys)-1].Time > c.Loop {
		return nil, fmt.Errorf("modulation: key exceeds loop period")
	}
	for _, input := range c.Inputs {
		if input.Name == "" || !finite(input.Gain) {
			return nil, fmt.Errorf("modulation: invalid input binding")
		}
	}
	if c.Clamp != nil {
		if !finite(c.Clamp.Min) || !finite(c.Clamp.Max) || c.Clamp.Min > c.Clamp.Max {
			return nil, fmt.Errorf("modulation: invalid clamp")
		}
		copy := *c.Clamp
		c.Clamp = &copy
	}
	c.Keys = append([]Key(nil), c.Keys...)
	c.Oscillators = append([]Oscillator(nil), c.Oscillators...)
	c.Inputs = append([]Input(nil), c.Inputs...)
	return &Signal{spec: c}, nil
}

func (s *Signal) At(c Context) float64 {
	t := c.Seconds
	if s.spec.TimeBase == Beats {
		t = c.Beats
	}
	if !finite(t) {
		t = 0
	}
	v := s.spec.Base + s.keyAt(t)
	for _, o := range s.spec.Oscillators {
		phase := t*o.Frequency + o.Phase
		if !finite(phase) {
			continue
		}
		phase -= math.Floor(phase)
		wave := 0.0
		switch o.Shape {
		case Sine:
			wave = math.Sin(phase * 2 * math.Pi)
		case Cosine:
			wave = math.Cos(phase * 2 * math.Pi)
		case Triangle:
			wave = 1 - 4*math.Abs(phase-.5)
		case Saw:
			wave = 2*phase - 1
		case Square:
			wave = 1
			if phase >= .5 {
				wave = -1
			}
		}
		v += o.Amplitude * wave
	}
	for _, input := range s.spec.Inputs {
		if x := c.Inputs[input.Name]; finite(x) {
			v += x * input.Gain
		}
	}
	if s.spec.Clamp != nil {
		v = math.Max(s.spec.Clamp.Min, math.Min(s.spec.Clamp.Max, v))
	}
	return v
}

func (s *Signal) keyAt(t float64) float64 {
	keys := s.spec.Keys
	if len(keys) == 0 {
		return 0
	}
	if s.spec.Loop > 0 {
		t -= math.Floor(t/s.spec.Loop) * s.spec.Loop
	}
	if t <= keys[0].Time {
		return keys[0].Value
	}
	i := sort.Search(len(keys), func(i int) bool { return keys[i].Time > t })
	if i == len(keys) {
		return keys[i-1].Value
	}
	a, b := keys[i-1], keys[i]
	u := (t - a.Time) / (b.Time - a.Time)
	switch a.Interpolation {
	case Hold:
		u = 0
	case Smooth:
		u = u * u * (3 - 2*u)
	}
	return a.Value + (b.Value-a.Value)*u
}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
