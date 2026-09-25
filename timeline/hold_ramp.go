package timeline

import (
	"fmt"
	"math"
)

// HoldRampConfig describes a finite tick-based presentation. Frames updates
// reach full progress; the following Step reports completion. InteriorOffset
// adjusts only partial frames, preserving unmodified zero and full endpoints.
type HoldRampConfig struct {
	Frames         int
	InteriorOffset float64
}

// HoldRamp owns a finite tick counter and a prepared visual progress value.
// Draw may read Progress repeatedly without advancing the phase.
type HoldRamp struct {
	config   HoldRampConfig
	tick     int
	progress float64
}

func NewHoldRamp(config HoldRampConfig) (*HoldRamp, error) {
	if config.Frames < 1 || config.Frames > 1<<24 || math.IsNaN(config.InteriorOffset) || math.IsInf(config.InteriorOffset, 0) {
		return nil, fmt.Errorf("timeline: invalid hold ramp")
	}
	return &HoldRamp{config: config}, nil
}

// Step returns done after the configured full-progress tick. It does not
// automatically reset, so a scene can select its own next-stage boundary.
func (ramp *HoldRamp) Step() (done bool) {
	if ramp.tick >= ramp.config.Frames {
		return true
	}
	ramp.tick++
	progress := float64(ramp.tick) / float64(ramp.config.Frames)
	if progress > 0 && progress < 1 {
		progress = min(1, max(0, progress+ramp.config.InteriorOffset))
	}
	ramp.progress = progress
	return false
}

func (ramp *HoldRamp) Tick() int         { return ramp.tick }
func (ramp *HoldRamp) Progress() float64 { return ramp.progress }

func (ramp *HoldRamp) Reset() { ramp.tick, ramp.progress = 0, 0 }
