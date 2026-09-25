package motion

import (
	"fmt"
	"math"
)

// WaveClockConfig binds a reusable spatial/time wave to an independently
// advanced phase. Call At before Step when the source screen draws first.
type WaveClockConfig struct {
	Wave        Wave
	Start, Step float64
}

// WaveClock keeps one authored wave phase without touching a renderer.
type WaveClock struct {
	config WaveClockConfig
	phase  float64
}

func NewWaveClock(config WaveClockConfig) (*WaveClock, error) {
	for _, value := range [...]float64{config.Wave.Amplitude, config.Wave.Spatial, config.Wave.Speed, config.Wave.Phase, config.Wave.Offset, config.Start, config.Step} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite wave clock setting")
		}
	}
	return &WaveClock{config: config, phase: config.Start}, nil
}

func (clock *WaveClock) At(position float64) float64 {
	return clock.config.Wave.At(position, clock.phase)
}

// Phase reports the phase used by the next At call.
func (clock *WaveClock) Phase() float64 { return clock.phase }

// Step advances the phase once, independently of how many times At was called.
func (clock *WaveClock) Step() { clock.phase += clock.config.Step }

// SetStep changes the next advance without discarding the current phase.
func (clock *WaveClock) SetStep(step float64) error {
	if math.IsNaN(step) || math.IsInf(step, 0) {
		return fmt.Errorf("motion: nonfinite wave clock step")
	}
	clock.config.Step = step
	return nil
}

// SetPhase jumps to an authored phase, for example at a timeline cue.
func (clock *WaveClock) SetPhase(phase float64) error {
	if math.IsNaN(phase) || math.IsInf(phase, 0) {
		return fmt.Errorf("motion: nonfinite wave clock phase")
	}
	clock.phase = phase
	return nil
}

// Reset restores the configured starting phase and retains the current step.
func (clock *WaveClock) Reset() { clock.phase = clock.config.Start }
