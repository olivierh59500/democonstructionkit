package presets

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// ReplicantsStarLayer sets population and horizontal speed for one material.
type ReplicantsStarLayer struct {
	Count int
	Speed float64
}

// ReplicantsStarOptions keeps layer speeds/counts, bounds and random spawning
// independent of the renderer. RandomInt must return a value in [0,n).
type ReplicantsStarOptions struct {
	Width, Height int
	Layers        []ReplicantsStarLayer
	RandomInt     func(n int) int
}

func DefaultReplicantsStarOptions(randomInt func(int) int) ReplicantsStarOptions {
	return ReplicantsStarOptions{Width: 640, Height: 280, RandomInt: randomInt,
		Layers: []ReplicantsStarLayer{{Count: 35, Speed: 11.2},
			{Count: 35, Speed: 5.6}, {Count: 35, Speed: 2.8}}}
}

// ReplicantsStarMaterials returns independent color recipes for the three
// depth-like star layers. Any compatible image sequence may replace them.
func ReplicantsStarMaterials() []sprites.SolidFrame {
	return []sprites.SolidFrame{
		{Width: 2, Height: 2, Color: color.RGBA{R: 0xE0, G: 0xA0, B: 0xA0, A: 0xFF}},
		{Width: 2, Height: 2, Color: color.RGBA{R: 0xC0, G: 0x60, B: 0x60, A: 0xFF}},
		{Width: 2, Height: 2, Color: color.RGBA{R: 0x80, G: 0x40, B: 0x40, A: 0xFF}},
	}
}

// ReplicantsStars shares the same bounded animated field as DOM's atlas stars.
// It selects a fixed image per particle and wraps X once before respawning Y.
func ReplicantsStars(frames []*ebiten.Image, c ReplicantsStarOptions) (sprites.AnimatedFieldConfig, error) {
	if c.Width < 1 || c.Height < 1 || len(c.Layers) == 0 || len(c.Layers) > 256 || c.RandomInt == nil ||
		(len(frames) != 0 && len(frames) < len(c.Layers)) {
		return sprites.AnimatedFieldConfig{}, fmt.Errorf("presets: invalid Replicants star field")
	}
	count := 0
	for _, layer := range c.Layers {
		if layer.Count < 1 || layer.Count > 1<<16-count || math.IsNaN(layer.Speed) || math.IsInf(layer.Speed, 0) || layer.Speed <= 0 {
			return sprites.AnimatedFieldConfig{}, fmt.Errorf("presets: invalid Replicants star layer")
		}
		count += layer.Count
	}
	material := make([]int, count)
	velocity := make([]float64, count)
	next := 0
	for index, layer := range c.Layers {
		for range layer.Count {
			material[next], velocity[next] = index, layer.Speed
			next++
		}
	}
	spawn := func(index int, _ bool) motion.FrameParticle {
		return motion.FrameParticle{X: float64(c.RandomInt(c.Width)), Y: float64(c.RandomInt(c.Height)),
			VelocityX: velocity[index], Image: material[index]}
	}
	wrap := &motion.FrameAxisWrap{Boundary: float64(c.Width), Shift: float64(c.Width),
		OnWrap: func(_ int, p *motion.FrameParticle) { p.Y = float64(c.RandomInt(c.Height)) }}
	return sprites.AnimatedFieldConfig{
		Motion: motion.FrameFieldConfig{Count: count, Spawn: spawn, WrapX: wrap},
		Frames: frames, ImageByParticle: true,
	}, nil
}
