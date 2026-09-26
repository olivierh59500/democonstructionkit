package motion

import (
	"fmt"
	"math"
)

// TwistingRibbonConfig places source strips along two oscillating faces.
// Position and phase wraps use strict greater-than checks to preserve authored
// boundary frames. Every geometric threshold is independent of the images.
type TwistingRibbonConfig struct {
	Width, StripWidth, IndexStep             int
	Phase, PhaseStep, PhaseWrap, FarStart    int
	NearAmplitude, FarAmplitude              float64
	NearScale, FarScale                      float64
	AngleMultiplier, NearDivisor, FarDivisor float64
	BaseAngle, FaceGap, BaseY                float64
	BackVisibleBelow, BackVisibleAbove       float64
	FrontVisibleAbove, FrontVisibleBelow     float64
	UpperClipStart, UpperClipEnd             float64
	LowerClipStart, LowerClipEnd             float64
	MinimumHeight                            float64
}

// TwistingRibbonSlice is one source strip with optional back/front placements.
// Negative BackScaleY mirrors the back image in destination space.
type TwistingRibbonSlice struct {
	SourceX                                int
	BackY, BackScaleY, FrontY, FrontScaleY float64
	BackVisible, FrontVisible              bool
}

// TwistingRibbon caches one bounded pose array and advances its phase only on
// Advance. A caller can draw the same prepared frame on several destinations.
type TwistingRibbon struct {
	config TwistingRibbonConfig
	phase  int
	poses  []TwistingRibbonSlice
}

func NewTwistingRibbon(c TwistingRibbonConfig) (*TwistingRibbon, error) {
	if c.Width < 1 || c.Width > 8192 || c.StripWidth < 1 || c.Width%c.StripWidth != 0 ||
		c.IndexStep < 0 || c.PhaseStep < 0 || c.PhaseWrap < 1 || c.Phase < 0 || c.Phase > c.PhaseWrap ||
		c.FarStart < 0 || c.FarStart > c.PhaseWrap || c.NearAmplitude == 0 || c.FarAmplitude == 0 ||
		c.NearScale == 0 || c.FarScale == 0 || c.NearDivisor == 0 || c.FarDivisor == 0 {
		return nil, fmt.Errorf("motion: invalid twisting ribbon geometry or clock")
	}
	for _, value := range [...]float64{c.NearAmplitude, c.FarAmplitude, c.NearScale, c.FarScale,
		c.AngleMultiplier, c.NearDivisor, c.FarDivisor, c.BaseAngle, c.FaceGap, c.BaseY,
		c.BackVisibleBelow, c.BackVisibleAbove, c.FrontVisibleAbove, c.FrontVisibleBelow,
		c.UpperClipStart, c.UpperClipEnd, c.LowerClipStart, c.LowerClipEnd, c.MinimumHeight} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite twisting ribbon parameter")
		}
	}
	if c.BackVisibleBelow > c.BackVisibleAbove || c.FrontVisibleAbove > c.FrontVisibleBelow ||
		c.UpperClipStart > c.UpperClipEnd || c.LowerClipStart > c.LowerClipEnd || c.MinimumHeight < 0 {
		return nil, fmt.Errorf("motion: inverted twisting ribbon thresholds")
	}
	r := &TwistingRibbon{config: c, phase: c.Phase, poses: make([]TwistingRibbonSlice, c.Width/c.StripWidth)}
	r.sample()
	return r, nil
}

// Advance selects the next complete pose after the current frame was drawn.
func (r *TwistingRibbon) Advance() {
	if r == nil {
		return
	}
	r.phase += r.config.PhaseStep
	if r.phase > r.config.PhaseWrap {
		r.phase -= r.config.PhaseWrap
	}
	r.sample()
}

func (r *TwistingRibbon) Phase() int                   { return r.phase }
func (r *TwistingRibbon) Poses() []TwistingRibbonSlice { return r.poses }

func (r *TwistingRibbon) sample() {
	c := r.config
	position := r.phase
	for slot := range r.poses {
		amplitude, heightScale := c.NearAmplitude, c.NearScale
		decal := float64(position) * c.AngleMultiplier * math.Pi / c.NearDivisor
		if position >= c.FarStart {
			amplitude, heightScale = c.FarAmplitude, c.FarScale
			decal = float64(position-c.FarStart) * c.AngleMultiplier * math.Pi / c.FarDivisor
		}
		position += c.IndexStep
		if position > c.PhaseWrap {
			position -= c.PhaseWrap
		}
		a1 := math.Mod(c.BaseAngle+decal, 2*math.Pi)
		a2 := math.Mod(a1+c.FaceGap, 2*math.Pi)
		y1 := math.Floor(amplitude*math.Sin(a1) + .5)
		y2 := math.Floor(amplitude*math.Sin(a2) + .5)
		pose := TwistingRibbonSlice{SourceX: slot * c.StripWidth}
		if a1 > c.BackVisibleAbove || a1 < c.BackVisibleBelow {
			from, to := y1, y2
			if a1 > c.LowerClipStart && a1 < c.LowerClipEnd {
				from = -amplitude
			}
			if a1 > c.UpperClipStart && a1 < c.UpperClipEnd {
				to = amplitude
			}
			height := (to - from) / amplitude / heightScale
			if height > c.MinimumHeight {
				pose.BackVisible = true
				pose.BackY, pose.BackScaleY = amplitude+c.BaseY+to, -height
			}
		}
		if a1 > c.FrontVisibleAbove && a1 < c.FrontVisibleBelow {
			from, to := y2, y1
			if a1 > c.UpperClipStart && a1 < c.UpperClipEnd {
				to = amplitude
			}
			if a1 > c.LowerClipStart && a1 < c.LowerClipEnd {
				from = -amplitude
			}
			height := (to - from) / amplitude / heightScale
			if height > c.MinimumHeight {
				pose.FrontVisible = true
				pose.FrontY, pose.FrontScaleY = amplitude+c.BaseY+from, height
			}
		}
		r.poses[slot] = pose
	}
}
