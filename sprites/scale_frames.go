package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// ScaleFramesConfig pre-renders a bounded image bank. Empty Scales selects the
// evenly spaced sequence (index+1)/Count; explicit Scales permits any zoom
// progression. The caller owns and deallocates the resulting images.
type ScaleFramesConfig struct {
	Source *ebiten.Image
	Count  int
	Scales []float64
	Filter ebiten.Filter
}

func NewScaleFrames(c ScaleFramesConfig) ([]*ebiten.Image, error) {
	if c.Source == nil || c.Source.Bounds().Empty() || c.Count < 1 || c.Count > 1024 ||
		(len(c.Scales) != 0 && len(c.Scales) != c.Count) {
		return nil, fmt.Errorf("sprites: invalid scale frame source or count")
	}
	frames := make([]*ebiten.Image, 0, c.Count)
	for i := 0; i < c.Count; i++ {
		scale := float64(i+1) / float64(c.Count)
		if len(c.Scales) > 0 {
			scale = c.Scales[i]
		}
		if math.IsNaN(scale) || math.IsInf(scale, 0) || scale <= 0 ||
			float64(c.Source.Bounds().Dx())*scale > 8192 || float64(c.Source.Bounds().Dy())*scale > 8192 {
			for _, frame := range frames {
				frame.Deallocate()
			}
			return nil, fmt.Errorf("sprites: invalid scale frame %d", i)
		}
		width := max(1, int(float64(c.Source.Bounds().Dx())*scale))
		height := max(1, int(float64(c.Source.Bounds().Dy())*scale))
		frame := ebiten.NewImage(width, height)
		var op ebiten.DrawImageOptions
		op.Filter = c.Filter
		op.GeoM.Scale(scale, scale)
		frame.DrawImage(c.Source, &op)
		frames = append(frames, frame)
	}
	return frames, nil
}
