package authoring

import (
	"fmt"
	"image/color"

	"github.com/olivierh59500/democonstructionkit/effects"
)

type JellyCubeStep struct {
	Mode     string  `json:"mode"`
	Duration float64 `json:"duration"`
}
type JellyDeformation struct {
	Wobble      *float64 `json:"wobble,omitempty"`
	Ripple      *float64 `json:"ripple,omitempty"`
	Squash      *float64 `json:"squash,omitempty"`
	Twist       *float64 `json:"twist,omitempty"`
	Translation *float64 `json:"translation,omitempty"`
}

// JellyCube saves one complete independently animated cube. Omitted properties
// inherit the chosen preset; pointers distinguish an explicit zero from omission.
// HalfEdge is the distance from model center to a face, in world units.
type JellyCube struct {
	Preset        string            `json:"preset,omitempty"`
	Center        *Point            `json:"center,omitempty"`
	HalfEdge      *float64          `json:"halfEdge,omitempty"`
	Focal         *float64          `json:"focal,omitempty"`
	Depth         *float64          `json:"depth,omitempty"`
	Colors        *[6][4]uint8      `json:"colors,omitempty"`
	Steps         []JellyCubeStep   `json:"steps,omitempty"`
	Speed         *float64          `json:"speed,omitempty"`
	RotationSpeed *float64          `json:"rotationSpeed,omitempty"`
	Phase         *float64          `json:"phase,omitempty"`
	Delay         *float64          `json:"delay,omitempty"`
	ZoomDuration  *float64          `json:"zoomDuration,omitempty"`
	Transition    *float64          `json:"transition,omitempty"`
	Deformation   *JellyDeformation `json:"deformation,omitempty"`
}

func jellyConfig(s JellyCube) (effects.JellyCubeConfig, error) {
	c := effects.DefaultJellyCubeConfig()
	switch s.Preset {
	case "", "default":
	case "dma-is-back":
		c = effects.DMAJellyCubeConfig()
	default:
		return c, fmt.Errorf("unknown jelly cube preset %q", s.Preset)
	}
	set := func(target *float64, value *float64) {
		if value != nil {
			*target = *value
		}
	}
	if s.Center != nil {
		c.X, c.Y = s.Center.X, s.Center.Y
	}
	set(&c.Size, s.HalfEdge)
	set(&c.CameraFOV, s.Focal)
	set(&c.CameraOffset, s.Depth)
	set(&c.Speed, s.Speed)
	set(&c.RotationSpeed, s.RotationSpeed)
	set(&c.Phase, s.Phase)
	set(&c.Delay, s.Delay)
	set(&c.ZoomDuration, s.ZoomDuration)
	set(&c.Transition, s.Transition)
	if s.Colors != nil {
		for i, v := range *s.Colors {
			c.Colors[i] = color.RGBA{R: v[0], G: v[1], B: v[2], A: v[3]}
		}
	}
	if s.Steps != nil {
		if len(s.Steps) > 256 {
			return c, fmt.Errorf("too many cube steps")
		}
		c.Steps = nil
		for _, s := range s.Steps {
			var mode effects.JellyCubeMode
			switch s.Mode {
			case "normal":
				mode = effects.JellyNormal
			case "tumble":
				mode = effects.JellyTumble
			case "pulsate":
				mode = effects.JellyPulsate
			case "swing":
				mode = effects.JellySwing
			case "bounce":
				mode = effects.JellyBounce
			default:
				return c, fmt.Errorf("unknown cube mode %q", s.Mode)
			}
			c.Steps = append(c.Steps, effects.JellyCubeStep{Mode: mode, Duration: s.Duration})
		}
	}
	if d := s.Deformation; d != nil {
		set(&c.Deformation.Wobble, d.Wobble)
		set(&c.Deformation.Ripple, d.Ripple)
		set(&c.Deformation.Squash, d.Squash)
		set(&c.Deformation.Twist, d.Twist)
		set(&c.Deformation.Translation, d.Translation)
	}
	return c, c.Validate()
}
