package motion

import (
	"fmt"
	"math"
)

// ScaledTextClockConfig synchronizes several differently scaled renderers of
// one message. FontAt contains the selected bank for each visible character.
// BaseSpeeds gives the reference bank's pixels per update for each active bank.
// Scales are horizontal glyph scales relative to the source font cell width.
type ScaledTextClockConfig struct {
	FontAt          []int
	Scales          []float64
	BaseSpeeds      []float64
	ViewportWidth   float64
	TileWidth       float64
	StartOffset     float64
	SpeedMultiplier float64
	InitialBank     int
	Lookahead       int
	WrapInclusive   bool
}

// ScaledTextClock owns transport and font-cue timing without any image state.
// All banks use one reference offset, so a size switch keeps the same letter
// aligned to the right edge. Step and accessors allocate nothing.
type ScaledTextClock struct {
	config     ScaledTextClockConfig
	offsets    []float64
	active     int
	baseOffset float64
	multiplier float64
	lookahead  int
}

func NewScaledTextClock(c ScaledTextClockConfig) (*ScaledTextClock, error) {
	count := len(c.Scales)
	if count == 0 || count > 256 || len(c.BaseSpeeds) != count || len(c.FontAt) == 0 || len(c.FontAt) > 1<<20 ||
		c.InitialBank < 0 || c.InitialBank >= count || c.Lookahead < 0 ||
		!finiteScaledClock(c.ViewportWidth) || c.ViewportWidth <= 0 ||
		!finiteScaledClock(c.TileWidth) || c.TileWidth <= 0 ||
		!finiteScaledClock(c.StartOffset) || !finiteScaledClock(c.SpeedMultiplier) || c.SpeedMultiplier < 0 {
		return nil, fmt.Errorf("motion: invalid scaled text clock")
	}
	for i := range c.Scales {
		if !finiteScaledClock(c.Scales[i]) || c.Scales[i] <= 0 ||
			!finiteScaledClock(c.BaseSpeeds[i]) || c.BaseSpeeds[i] < 0 {
			return nil, fmt.Errorf("motion: invalid scaled text bank %d", i)
		}
		cell := c.TileWidth * c.Scales[i]
		ratio := c.Scales[i] / c.Scales[0]
		if !finiteScaledClock(cell) || cell <= 0 || !finiteScaledClock(ratio) ||
			!finiteScaledClock(c.StartOffset*ratio+(1-ratio)*c.ViewportWidth) ||
			c.ViewportWidth/cell > 1<<30 || c.BaseSpeeds[i]*c.SpeedMultiplier/cell > 1<<30 {
			return nil, fmt.Errorf("motion: unrepresentable scaled text bank %d", i)
		}
	}
	length := float64(len(c.FontAt)) * c.TileWidth * c.Scales[0]
	if !finiteScaledClock(length) || length <= 0 || !finiteScaledClock(length+c.ViewportWidth) ||
		math.Abs(c.StartOffset/(c.TileWidth*c.Scales[0])) > 1<<30 {
		return nil, fmt.Errorf("motion: unrepresentable scaled text span")
	}
	for _, bank := range c.FontAt {
		if bank < 0 || bank >= count {
			return nil, fmt.Errorf("motion: font cue selects an unknown bank")
		}
	}
	c.Scales = append([]float64(nil), c.Scales...)
	c.BaseSpeeds = append([]float64(nil), c.BaseSpeeds...)
	c.FontAt = append([]int(nil), c.FontAt...)
	clock := &ScaledTextClock{config: c, offsets: make([]float64, count),
		active: c.InitialBank, baseOffset: c.StartOffset, multiplier: c.SpeedMultiplier}
	clock.syncOffsets()
	return clock, nil
}

func finiteScaledClock(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func (c *ScaledTextClock) syncOffsets() {
	baseScale := c.config.Scales[0]
	for i, scale := range c.config.Scales {
		ratio := scale / baseScale
		c.offsets[i] = c.baseOffset*ratio + (1-ratio)*c.config.ViewportWidth
	}
}

// Step moves the reference bank, applies its authored one-step wrap, then
// samples the font cue at the right-edge lookahead of the currently active
// bank. A cue changes the speed on the following update, not retroactively.
func (c *ScaledTextClock) Step() {
	c.baseOffset -= c.config.BaseSpeeds[c.active] * c.multiplier
	length := float64(len(c.config.FontAt)) * c.config.TileWidth * c.config.Scales[0]
	if c.config.WrapInclusive && c.baseOffset <= -length || !c.config.WrapInclusive && c.baseOffset < -length {
		c.baseOffset += length + c.config.ViewportWidth
	}
	c.syncOffsets()
	cellWidth := c.config.TileWidth * c.config.Scales[c.active]
	left := int(math.Floor(-c.offsets[c.active] / cellWidth))
	if left < 0 {
		left = 0
	}
	visible := int(math.Ceil(c.config.ViewportWidth/cellWidth)) + c.config.Lookahead
	position := left + visible
	if position >= len(c.config.FontAt) {
		position = len(c.config.FontAt) - 1
	}
	c.lookahead = position
	c.active = c.config.FontAt[position]
}

// SetSpeedMultiplier lets controls or music cues change movement without
// resetting the text position or active font.
func (c *ScaledTextClock) SetSpeedMultiplier(multiplier float64) error {
	if !finiteScaledClock(multiplier) || multiplier < 0 {
		return fmt.Errorf("motion: invalid scaled text speed")
	}
	for i, speed := range c.config.BaseSpeeds {
		if !finiteScaledClock(speed*multiplier) || speed*multiplier/(c.config.TileWidth*c.config.Scales[i]) > 1<<30 {
			return fmt.Errorf("motion: unrepresentable speed for text bank %d", i)
		}
	}
	c.multiplier = multiplier
	return nil
}

func (c *ScaledTextClock) ActiveBank() int          { return c.active }
func (c *ScaledTextClock) Offset(bank int) float64  { return c.offsets[bank] }
func (c *ScaledTextClock) LookaheadIndex() int      { return c.lookahead }
func (c *ScaledTextClock) SpeedMultiplier() float64 { return c.multiplier }
