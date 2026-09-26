package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// SampledSpriteTrainConfig gives one borrowed image a frame region per sprite.
// Motion chooses trajectories and timing; artwork, filter and scale are free
// to differ between productions using the same coordinate sampler.
type SampledSpriteTrainConfig struct {
	Motion         motion.SampledSpriteTrainConfig
	Image          *ebiten.Image
	Regions        []composite.Region
	ScaleX, ScaleY float64
	Filter         ebiten.Filter
	Blend          ebiten.Blend
}

// SampledSpriteTrain prepares poses once per Update and paints only its cached
// cropped frames. It owns neither the source atlas nor an intermediate surface.
type SampledSpriteTrain struct {
	motion *motion.SampledSpriteTrain
	config SampledSpriteTrainConfig
}

func NewSampledSpriteTrain(c SampledSpriteTrainConfig) (*SampledSpriteTrain, error) {
	if c.Image == nil || len(c.Regions) < c.Motion.Count ||
		math.IsNaN(c.ScaleX) || math.IsInf(c.ScaleX, 0) ||
		math.IsNaN(c.ScaleY) || math.IsInf(c.ScaleY, 0) {
		return nil, fmt.Errorf("sprites: invalid sampled sprite image or scale")
	}
	if c.ScaleX == 0 {
		c.ScaleX = 1
	}
	if c.ScaleY == 0 {
		c.ScaleY = 1
	}
	for _, region := range c.Regions {
		if math.IsNaN(region.X) || math.IsNaN(region.Y) ||
			math.IsNaN(region.Width) || math.IsNaN(region.Height) ||
			region.Width <= 0 || region.Height <= 0 ||
			region.X < float64(c.Image.Bounds().Min.X) ||
			region.Y < float64(c.Image.Bounds().Min.Y) ||
			region.X+region.Width > float64(c.Image.Bounds().Max.X) ||
			region.Y+region.Height > float64(c.Image.Bounds().Max.Y) {
			return nil, fmt.Errorf("sprites: invalid sampled sprite crop")
		}
	}
	clock, err := motion.NewSampledSpriteTrain(c.Motion)
	if err != nil {
		return nil, err
	}
	c.Regions = append([]composite.Region(nil), c.Regions...)
	if c.Blend == (ebiten.Blend{}) {
		c.Blend = ebiten.BlendSourceOver
	}
	return &SampledSpriteTrain{motion: clock, config: c}, nil
}

func (t *SampledSpriteTrain) Update(kit.Frame) error { t.motion.Step(); return nil }

func (t *SampledSpriteTrain) Draw(dst *ebiten.Image) {
	if t == nil || dst == nil {
		return
	}
	for _, pose := range t.motion.Samples() {
		var options ebiten.DrawImageOptions
		options.Filter, options.Blend = t.config.Filter, t.config.Blend
		options.GeoM.Scale(t.config.ScaleX, t.config.ScaleY)
		options.GeoM.Translate(pose.X, pose.Y)
		composite.DrawRegion(dst, t.config.Image, t.config.Regions[pose.Frame], &options)
	}
}

func (t *SampledSpriteTrain) Motion() *motion.SampledSpriteTrain { return t.motion }
