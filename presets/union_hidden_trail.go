package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/sprites"
	"github.com/olivierh59500/democonstructionkit/timeline"
)

// UnionHiddenTrail keeps the four independent pointer images at delays 0, 20,
// 40 and 60, drawing the oldest first so the live pointer remains on top.
func UnionHiddenTrail(images []*ebiten.Image) sprites.DelayedTrailConfig {
	return sprites.DelayedTrailConfig{
		Images: images, Delays: []int{0, 20, 40, 60}, Capacity: 61,
		Initial: geometry.Vec2{X: 384, Y: 268}, Reverse: true,
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver,
	}
}

// UnionHiddenPalette advances once per visual tick. Sample Current for the
// upper bar, then Step for the crosshair to preserve their one-color offset.
func UnionHiddenPalette(count int) timeline.PacedIndexConfig {
	return timeline.PacedIndexConfig{Count: count, Every: 1, First: 1}
}
