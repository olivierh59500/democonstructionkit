package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// LatchedSprite maps one borrowed image to a trigger channel. Zero scale and
// opacity default to one; positions and anchors are logical scene pixels.
type LatchedSprite struct {
	Image                        *ebiten.Image
	Channel                      int
	X, Y, AnchorX, AnchorY       float64
	ScaleX, ScaleY, AngleDegrees float64
	Opacity                      float64
}

type LatchedOverlayConfig struct {
	Motion  motion.LatchedTriggersConfig
	Sprites []LatchedSprite
	Filter  ebiten.Filter
	Blend   ebiten.Blend
}

// LatchedOverlay prepares trigger visibility once per Update, then draws any
// number of ordered sprites through those channels without owning their art.
type LatchedOverlay struct {
	motion *motion.LatchedTriggers
	config LatchedOverlayConfig
}

func NewLatchedOverlay(c LatchedOverlayConfig) (*LatchedOverlay, error) {
	if len(c.Sprites) == 0 || len(c.Sprites) > 1<<16 {
		return nil, fmt.Errorf("sprites: invalid latched overlay population")
	}
	clock, err := motion.NewLatchedTriggers(c.Motion)
	if err != nil {
		return nil, err
	}
	c.Sprites = append([]LatchedSprite(nil), c.Sprites...)
	for i := range c.Sprites {
		s := &c.Sprites[i]
		if s.Image == nil || s.Channel < 0 || s.Channel >= c.Motion.Count {
			return nil, fmt.Errorf("sprites: invalid latched sprite %d", i)
		}
		for _, value := range [...]float64{s.X, s.Y, s.AnchorX, s.AnchorY,
			s.ScaleX, s.ScaleY, s.AngleDegrees, s.Opacity} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, fmt.Errorf("sprites: nonfinite latched sprite pose")
			}
		}
		if s.ScaleX == 0 {
			s.ScaleX = 1
		}
		if s.ScaleY == 0 {
			s.ScaleY = 1
		}
		if s.Opacity == 0 {
			s.Opacity = 1
		}
		if s.Opacity < 0 || s.Opacity > 1 {
			return nil, fmt.Errorf("sprites: invalid latched sprite opacity")
		}
	}
	if c.Blend == (ebiten.Blend{}) {
		c.Blend = ebiten.BlendSourceOver
	}
	return &LatchedOverlay{motion: clock, config: c}, nil
}

func (o *LatchedOverlay) Update(kit.Frame) error { o.motion.Step(); return nil }

func (o *LatchedOverlay) Draw(dst *ebiten.Image) {
	if o == nil || dst == nil {
		return
	}
	for _, sprite := range o.config.Sprites {
		if !o.motion.Visible(sprite.Channel) {
			continue
		}
		var options ebiten.DrawImageOptions
		options.Filter, options.Blend = o.config.Filter, o.config.Blend
		options.GeoM.Translate(-sprite.AnchorX, -sprite.AnchorY)
		options.GeoM.Scale(sprite.ScaleX, sprite.ScaleY)
		options.GeoM.Rotate(sprite.AngleDegrees * math.Pi / 180)
		options.GeoM.Translate(sprite.X, sprite.Y)
		options.ColorScale.ScaleAlpha(float32(sprite.Opacity))
		dst.DrawImage(sprite.Image, &options)
	}
}

func (o *LatchedOverlay) Motion() *motion.LatchedTriggers { return o.motion }
