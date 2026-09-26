package motion

import (
	"fmt"
	"math"
)

// RibbonWrapMode selects a one-step historical reset threshold. A continuous
// seamless loop belongs to scrolling.Config.Repeat instead of this clock.
type RibbonWrapMode uint8

const (
	RibbonWrapBelow RibbonWrapMode = iota // Reset below -Length.
	RibbonWrapAbove                       // Reset above Length+Extent.
	RibbonWrapNone                        // Keep traveling without a reset.
)

// RibbonClockConfig describes movement in source pixels per simulation tick.
// Offset and Restart are independent so a message may enter from either side.
type RibbonClockConfig struct {
	Length, Offset, Velocity, Restart, Extent float64
	Multiplier                                float64
	Wrap                                      RibbonWrapMode
	Inclusive                                 bool
}

// RibbonClock owns one scalar transport without font or image resources.
type RibbonClock struct {
	config     RibbonClockConfig
	offset     float64
	multiplier float64
}

func NewRibbonClock(c RibbonClockConfig) (*RibbonClock, error) {
	if c.Wrap > RibbonWrapNone || c.Length <= 0 || c.Extent < 0 || c.Multiplier < 0 {
		return nil, fmt.Errorf("motion: invalid ribbon clock geometry")
	}
	for _, value := range [...]float64{c.Length, c.Offset, c.Velocity, c.Restart, c.Extent, c.Multiplier, c.Length + c.Extent} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("motion: nonfinite ribbon clock parameter")
		}
	}
	return &RibbonClock{config: c, offset: c.Offset, multiplier: c.Multiplier}, nil
}

func (c *RibbonClock) Step() error {
	next := c.offset + c.config.Velocity*c.multiplier
	if math.IsNaN(next) || math.IsInf(next, 0) {
		return fmt.Errorf("motion: ribbon clock overflow")
	}
	switch c.config.Wrap {
	case RibbonWrapBelow:
		boundary := -c.config.Length
		if next < boundary || c.config.Inclusive && next == boundary {
			next = c.config.Restart
		}
	case RibbonWrapAbove:
		boundary := c.config.Length + c.config.Extent
		if next > boundary || c.config.Inclusive && next == boundary {
			next = c.config.Restart
		}
	}
	c.offset = next
	return nil
}

func (c *RibbonClock) Offset() float64 { return c.offset }
func (c *RibbonClock) Reset()          { c.offset = c.config.Offset }

func (c *RibbonClock) SetMultiplier(value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return fmt.Errorf("motion: invalid ribbon multiplier")
	}
	c.multiplier = value
	return nil
}

func (c *RibbonClock) SetVelocity(value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("motion: invalid ribbon velocity")
	}
	c.config.Velocity = value
	return nil
}
