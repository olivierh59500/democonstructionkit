package presets

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// DMA3DStarOptions retains the authored starting formula, per-layer material
// and speed, vertical wrap range, and rectangular origin mask as editable data.
type DMA3DStarOptions struct {
	Width, VerticalRange int
	Layers               []ReplicantsStarLayer
	Materials            []sprites.SolidFrame
	Mask                 *image.Rectangle
	Source               *ebiten.Image
}

// DefaultDMA3DStarOptions reproduces the three-layer 640-pixel starfield and
// the center exclusion rectangle at any logical screen size.
func DefaultDMA3DStarOptions(width, height int) DMA3DStarOptions {
	mask := image.Rect(width/2-200, height/2-150, width/2+201, height/2+151)
	return DMA3DStarOptions{
		Width: width, VerticalRange: 280,
		Layers:    []ReplicantsStarLayer{{Count: 35, Speed: 11.2}, {Count: 35, Speed: 5.6}, {Count: 35, Speed: 2.8}},
		Materials: ReplicantsStarMaterials(), Mask: &mask,
	}
}

// DMA3DStarfield uses one strict X wrap per tick. Its Y reset depends on the
// global particle index and one-based tick, including ticks after a long run.
func DMA3DStarfield(c DMA3DStarOptions) (sprites.BatchedSolidFieldConfig, error) {
	if c.Width < 1 || c.VerticalRange < 1 || len(c.Layers) == 0 ||
		len(c.Layers) != len(c.Materials) {
		return sprites.BatchedSolidFieldConfig{}, fmt.Errorf("presets: invalid DMA 3D star field")
	}
	count := 0
	for _, layer := range c.Layers {
		if layer.Count < 1 || layer.Count > 1<<16-count ||
			math.IsNaN(layer.Speed) || math.IsInf(layer.Speed, 0) || layer.Speed <= 0 {
			return sprites.BatchedSolidFieldConfig{}, fmt.Errorf("presets: invalid DMA 3D star layer")
		}
		count += layer.Count
	}
	material := make([]int, count)
	localIndex := make([]int, count)
	velocity := make([]float64, count)
	next := 0
	for layerIndex, layer := range c.Layers {
		for i := range layer.Count {
			material[next], localIndex[next], velocity[next] = layerIndex, i, layer.Speed
			next++
		}
	}
	spawn := func(index int, _ bool) motion.FrameParticle {
		i, speed := localIndex[index], velocity[index]
		width, height := float64(c.Width), float64(c.VerticalRange)
		return motion.FrameParticle{
			X:         math.Mod(float64(i*73+int(speed)*137)*1.234/1000.0*width, width),
			Y:         math.Mod(float64(i*97+int(speed)*211)*2.345/1000.0*height, height),
			VelocityX: speed, Image: material[index],
		}
	}
	wrap := &motion.FrameAxisWrap{Boundary: float64(c.Width), Shift: float64(c.Width),
		OnWrapAt: func(index, tick int, p *motion.FrameParticle) {
			p.Y = math.Mod(float64(index*97+tick), float64(c.VerticalRange))
		}}
	return sprites.BatchedSolidFieldConfig{
		Motion:    motion.FrameFieldConfig{Count: count, Spawn: spawn, WrapX: wrap},
		Materials: c.Materials, Mask: c.Mask, Source: c.Source,
	}, nil
}
