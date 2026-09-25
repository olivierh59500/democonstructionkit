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
	HoldTicks   int
}

// WaveClock keeps one authored wave phase without touching a renderer.
type WaveClock struct {
	config WaveClockConfig
	phase  float64
	hold   int
}

func NewWaveClock(config WaveClockConfig) (*WaveClock, error) {
	for _, value := range [...]float64{config.Wave.Amplitude, config.Wave.Spatial, config.Wave.Speed, config.Wave.Phase, config.Wave.Offset, config.Start, config.Step} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite wave clock setting")
		}
	}
	if config.HoldTicks < 0 {
		return nil, fmt.Errorf("motion: negative wave clock hold")
	}
	return &WaveClock{config: config, phase: config.Start, hold: config.HoldTicks}, nil
}

func (clock *WaveClock) At(position float64) float64 {
	return clock.config.Wave.At(position, clock.phase)
}

// Phase reports the phase used by the next At call.
func (clock *WaveClock) Phase() float64 { return clock.phase }

// Step consumes an initial hold tick or advances the phase exactly once.
func (clock *WaveClock) Step() {
	if clock.hold > 0 {
		clock.hold--
		return
	}
	clock.phase += clock.config.Step
}

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

// HoldRemaining reports ticks before the next phase advance.
func (clock *WaveClock) HoldRemaining() int { return clock.hold }

// SetHold schedules another pause without changing the current phase.
func (clock *WaveClock) SetHold(ticks int) error {
	if ticks < 0 {
		return fmt.Errorf("motion: negative wave clock hold")
	}
	clock.hold = ticks
	return nil
}

// Reset restores the configured starting phase and hold, retaining the step.
func (clock *WaveClock) Reset() {
	clock.phase, clock.hold = clock.config.Start, clock.config.HoldTicks
}
