package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// GlowPainterConfig describes a reusable layered halo around a sprite group.
// The base image is drawn last; glow layers run from outermost to innermost.
// Anchor is measured in source-image pixels. PositionScale supports groups
// animated in a smaller logical viewport than the output surface.
type GlowPainterConfig struct {
	PositionScale, BaseScale, LayerScaleStep, GlowAlpha float64
	AnchorX, AnchorY                                    float64
	Layers                                              int
	GlowFilter, BaseFilter                              ebiten.Filter
	Blend                                               ebiten.Blend
}

type GlowPainter struct{ config GlowPainterConfig }

func NewGlowPainter(config GlowPainterConfig) (*GlowPainter, error) {
	for _, value := range []float64{config.PositionScale, config.BaseScale, config.LayerScaleStep, config.GlowAlpha, config.AnchorX, config.AnchorY} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("sprites: nonfinite glow parameter")
		}
	}
	if config.Layers < 0 || config.Layers > 32 || config.BaseScale <= 0 || config.PositionScale <= 0 || config.LayerScaleStep < 0 || config.GlowAlpha < 0 || config.GlowAlpha > 1 {
		return nil, fmt.Errorf("sprites: invalid glow painter configuration")
	}
	return &GlowPainter{config: config}, nil
}

// DrawGroup renders prepared group poses without advancing their animation.
// Frames are borrowed from the group, and drawing allocates no retained image.
func (painter *GlowPainter) DrawGroup(dst *ebiten.Image, group *Group) {
	if painter == nil || dst == nil || group == nil {
		return
	}
	for n := range group.poses {
		index := n
		if group.config.Reverse {
			index = len(group.poses) - 1 - n
		}
		pose := group.poses[index]
		if pose.Frame < 0 || pose.Frame >= len(group.config.Frames) || pose.Opacity <= 0 || !finiteField(pose.X) || !finiteField(pose.Y) || !finiteField(pose.ScaleX) || !finiteField(pose.ScaleY) || !finiteField(pose.Angle) {
			continue
		}
		image := group.config.Frames[pose.Frame]
		if image == nil {
			continue
		}
		for layer := painter.config.Layers; layer > 0; layer-- {
			var options ebiten.DrawImageOptions
			scale := painter.config.BaseScale + float64(layer)*painter.config.LayerScaleStep
			options.GeoM.Translate(-painter.config.AnchorX, -painter.config.AnchorY)
			options.GeoM.Scale(scale*pose.ScaleX, scale*pose.ScaleY)
			options.GeoM.Rotate(pose.Angle)
			options.GeoM.Translate(pose.X*painter.config.PositionScale, pose.Y*painter.config.PositionScale)
			options.ColorScale.ScaleAlpha(float32(painter.config.GlowAlpha / float64(layer) * pose.Opacity))
			options.Filter = painter.config.GlowFilter
			options.Blend = painter.config.Blend
			dst.DrawImage(image, &options)
		}
		var options ebiten.DrawImageOptions
		options.GeoM.Translate(-painter.config.AnchorX, -painter.config.AnchorY)
		options.GeoM.Scale(painter.config.BaseScale*pose.ScaleX, painter.config.BaseScale*pose.ScaleY)
		options.GeoM.Rotate(pose.Angle)
		options.GeoM.Translate(pose.X*painter.config.PositionScale, pose.Y*painter.config.PositionScale)
		options.ColorScale.ScaleAlpha(float32(pose.Opacity))
		options.Filter = painter.config.BaseFilter
		options.Blend = painter.config.Blend
		dst.DrawImage(image, &options)
	}
}
